package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
	"unicode"
)

var (
	//go:embed tmpls/base-unexported.go.tmpl
	baseUnexportedTmplText string

	//go:embed tmpls/usagefunc.go.tmpl
	usgFnTmplText string

	//go:embed tmpls/parsefunc.go.tmpl
	parseFnTmplText string
)

func generate(incVersion bool, pkgName string, usgTextWidth int, usgLayoutKind string, root *command) ([]byte, error) {
	g, err := newGenerator(usgTextWidth, usgLayoutKind)
	if err != nil {
		return nil, fmt.Errorf("initializing generator: %w", err)
	}
	if err = g.writeBase(incVersion, pkgName, root); err != nil {
		return nil, err
	}
	if err = g.genCommandCode(root); err != nil {
		return nil, err
	}
	return g.buf.Bytes(), nil
}

type generator struct {
	buf           bytes.Buffer
	usgTextWidth  int
	usgLayoutKind string
	usgFnTmpl     *template.Template
	parseFnTmpl   *template.Template
}

func newGenerator(usgTextWidth int, usgLayoutKind string) (generator, error) {
	usgFnTmpl := template.Must(template.New("usagefunc").Parse(usgFnTmplText))

	parseFuncs := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}
	parseFnTmpl, err := template.New("parsefunc").Funcs(parseFuncs).Parse(parseFnTmplText)
	if err != nil {
		return generator{}, fmt.Errorf("parsing template: %w", err)
	}

	return generator{
		usgTextWidth:  usgTextWidth,
		usgLayoutKind: usgLayoutKind,
		usgFnTmpl:     usgFnTmpl,
		parseFnTmpl:   parseFnTmpl,
	}, nil
}

type headerData struct {
	PkgName      string
	Version      string
	HasBool      bool
	HasFloat     bool
	HasInt       bool
	HasUint      bool
	HasNumber    bool
	HasSubcmds   bool
	Types        typeSet
	NeedsEnvCode bool
}

func (g *generator) writeBase(incVersion bool, pkgName string, root *command) error {
	ts := typeSet{}
	root.getTypes(ts)

	hasFloat := ts.HasAny("float32", "float64")
	hasInt := ts.HasAny("int", "int8", "int16", "int32", "int64")
	hasUint := ts.HasAny("uint", "uint8", "uint16", "uint32", "uint64")

	data := headerData{
		PkgName:      pkgName,
		Types:        ts,
		HasBool:      ts.HasAny("bool"),
		HasFloat:     hasFloat,
		HasInt:       hasInt,
		HasUint:      hasUint,
		HasNumber:    hasFloat || hasInt || hasUint,
		HasSubcmds:   root.HasSubcmds(),
		NeedsEnvCode: root.HasEnvArgOrOptSomewhere(),
	}
	if incVersion {
		data.Version = getBuildVersionInfo().String()
	}

	baseTmpl := template.Must(template.New("clapbase").Parse(baseUnexportedTmplText))
	if err := baseTmpl.Execute(&g.buf, &data); err != nil {
		return fmt.Errorf("executing base template: %w", err)
	}
	return nil
}

func (c *command) getTypes(ts typeSet) {
	for _, o := range c.Opts {
		if o.Name != "h" {
			ts[o.FieldType] = struct{}{}
		}
	}
	for _, a := range c.Args {
		ts[a.FieldType] = struct{}{}
	}
	for _, sc := range c.Subcmds {
		sc.getTypes(ts)
	}
}

type typeSet map[basicType]struct{}

func (ts typeSet) HasAny(names ...basicType) bool {
	for i := range names {
		if _, ok := ts[names[i]]; ok {
			return true
		}
	}
	return false
}

func (t basicType) ClapValueType() string {
	switch t {
	case "bool":
		return "Bool"
	case "string":
		return "String"
	case "float32", "float64":
		return "Float"
	case "int", "int8", "int16", "int32", "int64", "rune":
		return "Int"
	case "uint", "uint8", "uint16", "uint32", "uint64", "byte":
		return "Uint"
	default:
		panic("unknown basic type: " + t)
	}
}

func (g *generator) genCommandCode(c *command) error {
	for i := range c.Subcmds {
		if err := g.genCommandCode(&c.Subcmds[i]); err != nil {
			return err
		}
	}
	if err := g.genCmdUsageFunc(c); err != nil {
		return fmt.Errorf("generating '%s' usage help: %w", c.TypeName, err)
	}
	if err := g.genCmdParseFunc(c); err != nil {
		return fmt.Errorf("generating '%s' parse func: %w", c.TypeName, err)
	}
	return nil
}

type usgTmplData struct {
	OptUsgs    []string
	ArgUsgs    []string
	SubcmdUsgs []string

	*command
}

func (g *generator) genCmdUsageFunc(c *command) error {
	optUsgs := make([]string, len(c.Opts))
	{
		var optNameColWidth int
		for _, o := range c.Opts {
			if l := len(o.usgNameAndArg()); l > optNameColWidth {
				optNameColWidth = l
			}
		}
		for i, o := range c.Opts {
			switch g.usgLayoutKind {
			case "roomy":
				content := "   " + o.usgNameAndArg() + "\n"
				content += "      " + wrapBlurb(o.data.Blurb, 6, g.usgTextWidth)
				if v, ok := o.data.getConfig("env"); ok {
					content += "\n\n      [env: " + v + "]"
				}
				if i < len(c.Opts)-1 {
					content += "\n"
				}
				optUsgs[i] = content
			default:
				paddedNameAndArg := fmt.Sprintf("   %-*s   ", optNameColWidth, o.usgNameAndArg())
				desc := o.data.Blurb
				if v, ok := o.data.getConfig("env"); ok {
					desc += " [$" + v + "]"
				}
				optUsgs[i] = paddedNameAndArg + wrapBlurb(desc, len(paddedNameAndArg), g.usgTextWidth)
			}
		}
	}

	argUsgs := make([]string, len(c.Args))
	{
		var argNameColWidth int
		for _, a := range c.Args {
			if l := len(a.UsgName()); l > argNameColWidth {
				argNameColWidth = l
			}
		}
		for i, a := range c.Args {
			switch g.usgLayoutKind {
			case "roomy":
				content := "   " + a.UsgName() + "\n"
				content += "      " + wrapBlurb(a.data.Blurb, 6, g.usgTextWidth)
				if v, ok := a.data.getConfig("env"); ok {
					content += "\n\n      [env: " + v + "]"
				}
				if i < len(c.Args)-1 {
					content += "\n"
				}
				argUsgs[i] = content
			default:
				paddedName := fmt.Sprintf("   %-*s   ", argNameColWidth, a.UsgName())
				desc := a.data.Blurb
				if v, ok := a.data.getConfig("env"); ok {
					desc += " [$" + v + "]"
				}
				argUsgs[i] = paddedName + wrapBlurb(desc, len(paddedName), g.usgTextWidth)
			}
		}
	}

	subcmdUsgs := make([]string, len(c.Subcmds))
	{
		var subcmdNameColWidth int
		for _, sc := range c.Subcmds {
			if l := len(sc.UsgName()); l > subcmdNameColWidth {
				subcmdNameColWidth = l
			}
		}
		for i, sc := range c.Subcmds {
			switch g.usgLayoutKind {
			case "roomy":
				content := "   " + sc.UsgName() + "\n"
				content += "      " + wrapBlurb(sc.Data.Blurb, 6, g.usgTextWidth)
				if i < len(c.Subcmds)-1 {
					content += "\n"
				}
				subcmdUsgs[i] = content
			default:
				paddedName := fmt.Sprintf("   %-*s   ", subcmdNameColWidth, sc.UsgName())
				subcmdUsgs[i] = paddedName + wrapBlurb(sc.Data.Blurb, len(paddedName), g.usgTextWidth)
			}
		}
	}

	err := g.usgFnTmpl.Execute(&g.buf, usgTmplData{
		OptUsgs:    optUsgs,
		ArgUsgs:    argUsgs,
		SubcmdUsgs: subcmdUsgs,
		command:    c,
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *generator) genCmdParseFunc(c *command) error {
	err := g.parseFnTmpl.Execute(&g.buf, c)
	if err != nil {
		return err
	}
	return nil
}

func (c *command) Parents() string {
	s := ""
	for i := range c.parentNames {
		s += c.parentNames[i] + " "
	}
	return s
}

func (c *command) UsageLines() []string {
	var us []string
	for _, cfg := range c.Data.configs {
		if cfg.key == "cmd_usage" {
			us = append(us, c.UsgName()+" "+cfg.val)
		}
	}
	if len(us) > 0 {
		return us
	}
	optionsSlot := " [options]" // Every command has at least the help options for now.
	commandSlot := ""
	if c.HasSubcmds() {
		commandSlot = " <command>"
	}
	argsSlot := ""
	for i := range c.Args {
		argsSlot += " " + c.Args[i].UsgName()
	}
	return []string{
		c.UsgName() + optionsSlot + commandSlot + argsSlot,
	}
}

// QuotedNames returns a comma separated list of this command's name, plus any aliases,
// each in double quotes.
func (c *command) QuotedNames() string {
	s := "\"" + c.UsgName() + "\""
	if csv, ok := c.Data.getConfig("cmd_aliases"); ok {
		for _, alias := range strings.Split(csv, ",") {
			s += ", \"" + strings.TrimSpace(alias) + "\""
		}
	}
	return s
}

func (c *command) Overview() string {
	paras := c.Data.overview
	var s strings.Builder
	for i := range paras {
		s.WriteString("   ")
		// Drop any trailing new lines from the last paragraph so that there won't be any
		// extra space separating it from the options section that follows.
		if i == len(paras)-1 {
			s.WriteString(strings.TrimRight(paras[i], "\n"))
		} else {
			s.WriteString(paras[i])
		}
		// Separate the over paragraphs with a blank line by ensuring each paragraph
		// except the last one ends with two new lines.
		if i != len(paras)-1 && !strings.HasSuffix(paras[i], "\n\n") {
			s.WriteString("\n")
		}
	}
	return s.String()
}

func (o *option) EnvVar() string {
	name, _ := o.data.getConfig("env")
	return name
}

func (a *argument) EnvVar() string {
	name, _ := a.data.getConfig("env")
	return name
}

func (a *argument) UsgName() string {
	name := a.name
	if v, ok := a.data.getConfig("arg_name"); ok {
		name = v
	}
	if a.IsRequired() {
		return "<" + name + ">"
	}
	return "[" + name + "]"
}

func (a *argument) IsRequired() bool {
	_, ok := a.data.getConfig("arg_required")
	return ok
}

func (o *option) usgNameAndArg() string {
	s := "-" + o.Name
	if an := o.usgArgName(); an != "" {
		s += "  " + an
	}
	return s
}

// usgArgName returns the usage text of an option argument for non-boolean options. For
// example, if there's a string option named `file`, the usage might look something like
// `--file <arg>` where "<arg>" is the usage argument name text.
func (o *option) usgArgName() string {
	if o.FieldType.IsBool() {
		return ""
	}
	if name, ok := o.data.getConfig("opt_arg_name"); ok {
		return "<" + name + ">"
	}
	return "<arg>"
}

// HasEnvArgOrOptSomewhere returns true if this command or one of its subcommands contains
// an option or an argument that uses an environment variable config.
func (c *command) HasEnvArgOrOptSomewhere() bool {
	for i := range c.Opts {
		if _, ok := c.Opts[i].data.getConfig("env"); ok {
			return true
		}
	}
	for i := range c.Args {
		if _, ok := c.Args[i].data.getConfig("env"); ok {
			return true
		}
	}
	for i := range c.Subcmds {
		if c.Subcmds[i].HasEnvArgOrOptSomewhere() {
			return true
		}
	}
	return false
}

func (c *command) HasNonHelpOpts() bool {
	for i := range c.Opts {
		if c.Opts[i].Name != "h" {
			return true
		}
	}
	return false
}

func (c *command) HasSubcmds() bool { return len(c.Subcmds) > 0 }

func wrapBlurb(v string, indentLen, lineLen int) string {
	s := wrapText(v, indentLen, lineLen)
	return s[indentLen:]
}

type wordWrapper struct {
	indent string
	word   strings.Builder
	line   strings.Builder
	result strings.Builder
}

func wrapText(v string, indentLen, lineLen int) string {
	var ww wordWrapper
	ww.indent = strings.Repeat(" ", indentLen)
	ww.word.Grow(lineLen)
	ww.line.Grow(lineLen)
	ww.line.WriteString(ww.indent)
	ww.result.Grow(len(v))

	for _, c := range strings.TrimSpace(v) {
		if !unicode.IsSpace(c) {
			ww.word.WriteRune(c)
			continue
		}
		if c == '\n' {
			ww.takeWordAndReset()
			ww.takeLineAndReset()
			continue
		}
		if ww.line.Len()+ww.word.Len() > lineLen {
			ww.takeLineAndReset()
		}
		ww.takeWordAndReset()
		ww.line.WriteRune(c)
	}
	if ww.word.Len() > 0 {
		if ww.line.Len()+ww.word.Len() > lineLen {
			ww.takeLineAndReset()
		}
		ww.takeWordAndReset()
	}
	if ww.line.Len() > 0 {
		ww.result.WriteString(ww.line.String())
		ww.line.Reset()
	}

	res := ww.result.String()
	ww.result.Reset()
	return res
}

func (ww *wordWrapper) takeWordAndReset() {
	ww.line.WriteString(ww.word.String())
	ww.word.Reset()
}

func (ww *wordWrapper) takeLineAndReset() {
	ln := strings.TrimRightFunc(ww.line.String(), unicode.IsSpace) // remove trailing whitespace
	ww.result.WriteString(ln)
	ww.result.WriteRune('\n')
	ww.line.Reset()
	ww.line.WriteString(ww.indent)
}
