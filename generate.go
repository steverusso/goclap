package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
	"unicode"
)

const maxUsgLineLen = 90

var (
	//go:embed tmpls/header.go.tmpl
	headerTmplText string

	//go:embed tmpls/usagefunc.go.tmpl
	usgFnTmplText string

	//go:embed tmpls/parsefunc.go.tmpl
	parseFnTmplText string
)

func generate(incVersion bool, pkgName string, root *command) ([]byte, error) {
	g, err := newGenerator()
	if err != nil {
		return nil, fmt.Errorf("initializing generator: %w", err)
	}
	if err = g.writeHeader(incVersion, pkgName, root); err != nil {
		return nil, err
	}
	if err = g.genCommandCode(root); err != nil {
		return nil, fmt.Errorf("generating: %w", err)
	}
	return g.buf.Bytes(), nil
}

type generator struct {
	buf         bytes.Buffer
	usgFnTmpl   *template.Template
	parseFnTmpl *template.Template
}

func newGenerator() (generator, error) {
	usgFnTmpl, err := template.New("usagefunc").Parse(usgFnTmplText)
	if err != nil {
		return generator{}, fmt.Errorf("parsing template: %w", err)
	}
	parseFuncs := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}
	parseFnTmpl, err := template.New("parsefunc").Funcs(parseFuncs).Parse(parseFnTmplText)
	if err != nil {
		return generator{}, fmt.Errorf("parsing template: %w", err)
	}
	return generator{
		usgFnTmpl:   usgFnTmpl,
		parseFnTmpl: parseFnTmpl,
	}, nil
}

type headerData struct {
	PkgName string
	Version string
}

func (g *generator) writeHeader(incVersion bool, pkgName string, root *command) error {
	ts := typeSet{}
	root.getTypes(ts)

	t, err := template.New("header").Parse(headerTmplText)
	if err != nil {
		return fmt.Errorf("parsing header template: %w", err)
	}

	data := headerData{PkgName: pkgName}
	if incVersion {
		data.Version = getBuildVersionInfo().String()
	}

	if err = t.Execute(&g.buf, data); err != nil {
		return fmt.Errorf("executing header template: %w", err)
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

var clapValueTypes = map[basicType]string{
	// int*
	"int":   "Int",
	"int8":  "Int8",
	"int16": "Int16",
	"int32": "Int32",
	"int64": "Int64",
	// uint*
	"uint":   "Uint",
	"uint8":  "Uint8",
	"uint16": "Uint16",
	"uint32": "Uint32",
	"uint64": "Uint64",
	// float*
	"float32": "Float32",
	"float64": "Float64",
	// misc
	"bool":   "Bool",
	"string": "String",
	"byte":   "Uint8",
	"rune":   "Int32",
}

func (t basicType) ClapValueType() string {
	return clapValueTypes[t]
}

func (g *generator) genCommandCode(c *command) error {
	for i := range c.Subcmds {
		if err := g.genCommandCode(&c.Subcmds[i]); err != nil {
			return err
		}
	}
	if err := g.genCmdUsageFunc(c); err != nil {
		return fmt.Errorf("'%s': %w", c.TypeName, err)
	}
	if err := g.genCmdParseFunc(c); err != nil {
		return fmt.Errorf("'%s': %w", c.TypeName, err)
	}
	return nil
}

func (g *generator) genCmdUsageFunc(c *command) error {
	err := g.usgFnTmpl.Execute(&g.buf, c)
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
		s.WriteString(wrapText(paras[i], 3, maxUsgLineLen))
		if i != len(paras)-1 {
			s.WriteString("\n\n")
		}
	}
	return s.String()
}

func (c *command) OptNameColWidth() int {
	w := 0
	for _, o := range c.Opts {
		if l := len(o.usgNameAndArg()); l > w {
			w = l
		}
	}
	return w
}

func (c *command) ArgNameColWidth() int {
	w := 0
	for _, a := range c.Args {
		if l := len(a.UsgName()); l > w {
			w = l
		}
	}
	return w
}

func (c *command) SubcmdNameColWidth() int {
	w := 0
	for _, sc := range c.Subcmds {
		if l := len(sc.UsgName()); l > w {
			w = l
		}
	}
	return w
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
	if a.IsRequired() {
		return "<" + a.name + ">"
	}
	return "[" + a.name + "]"
}

func (a *argument) IsRequired() bool {
	_, ok := a.data.getConfig("arg_required")
	return ok
}

// Usg returns an argument's usage message text given how wide the name column should be.
func (a *argument) Usg(nameWidth int) string {
	paddedName := fmt.Sprintf("   %-*s   ", nameWidth, a.UsgName())
	desc := a.data.Blurb
	if v, ok := a.data.getConfig("env"); ok {
		desc += " [$" + v + "]"
	}
	return paddedName + wrapBlurb(desc, len(paddedName), maxUsgLineLen)
}

// Usg returns an option's usage message text given how wide the name column should be.
func (o *option) Usg(nameWidth int) string {
	paddedNameAndArg := fmt.Sprintf("   %-*s   ", nameWidth, o.usgNameAndArg())
	desc := o.data.Blurb
	if v, ok := o.data.getConfig("env"); ok {
		desc += " [$" + v + "]"
	}
	return paddedNameAndArg + wrapBlurb(desc, len(paddedNameAndArg), maxUsgLineLen)
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

// Usg returns a command's usage message text given how wide the name column should be.
func (c *command) Usg(nameWidth int) string {
	paddedName := fmt.Sprintf("   %-*s   ", nameWidth, c.UsgName())
	return paddedName + wrapBlurb(c.Data.Blurb, len(paddedName), maxUsgLineLen)
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
