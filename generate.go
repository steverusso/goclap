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
		return nil, fmt.Errorf("generating %w", err)
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
	RootCmd *command
	Types   typeSet

	NeedsEnvCode     bool
	NeedsStrconvCode bool
}

func (g *generator) writeHeader(incVersion bool, pkgName string, root *command) error {
	ts := typeSet{}
	root.getTypes(ts)

	t, err := template.New("header").Parse(headerTmplText)
	if err != nil {
		return fmt.Errorf("parsing header template: %w", err)
	}

	data := headerData{
		PkgName: pkgName,
		RootCmd: root,
		Types:   ts,

		NeedsEnvCode: root.HasEnvArgOrOptSomewhere(),
		NeedsStrconvCode: ts.HasAny("float32", "float64",
			"int", "int8", "int16", "int32", "int64", "rune",
			"uint", "uint8", "uint16", "uint32", "uint64", "byte",
		),
	}
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
		if o.Long != "help" {
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

var clapIterMethodMap = map[basicType]string{
	// int*
	"int":   "nextInt",
	"int8":  "nextInt8",
	"int16": "nextInt16",
	"int32": "nextInt32",
	"int64": "nextInt64",
	// uint*
	"uint":   "nextUint",
	"uint8":  "nextUint8",
	"uint16": "nextUint16",
	"uint32": "nextUint32",
	"uint64": "nextUint64",
	// float*
	"float32": "nextFloat32",
	"float64": "nextFloat64",
	// misc
	"bool":   "thisBool",
	"string": "nextStr",
	"byte":   "nextUint8",
	"rune":   "nextInt32",
}

func (t basicType) ClapIterMethodName() string {
	return clapIterMethodMap[t]
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
	us := make([]string, 0, 2)
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
		if l := len(o.usgNamesAndArg()); l > w {
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

type clapEnvValue struct {
	VarName   string
	FieldName string
	FieldType basicType
}

// EnvVals returns the environment variable name and the field name for any option or
// arguments that use an `env` config.
func (c *command) EnvVals() []clapEnvValue {
	envs := make([]clapEnvValue, 0, len(c.Opts)+len(c.Args))
	for i := range c.Opts {
		if name, ok := c.Opts[i].data.getConfig("env"); ok {
			envs = append(envs, clapEnvValue{
				VarName:   name,
				FieldName: c.Opts[i].FieldName,
				FieldType: c.Opts[i].FieldType,
			})
		}
	}
	for i := range c.Args {
		if name, ok := c.Args[i].data.getConfig("env"); ok {
			envs = append(envs, clapEnvValue{
				VarName:   name,
				FieldName: c.Args[i].FieldName,
				FieldType: c.Args[i].FieldType,
			})
		}
	}
	return envs
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
	paddedNameAndArg := fmt.Sprintf("   %-*s   ", nameWidth, o.usgNamesAndArg())
	desc := o.data.Blurb
	if v, ok := o.data.getConfig("env"); ok {
		desc += " [$" + v + "]"
	}
	return paddedNameAndArg + wrapBlurb(desc, len(paddedNameAndArg), maxUsgLineLen)
}

func (o *option) usgNamesAndArg() string {
	var s strings.Builder
	s.Grow(maxUsgLineLen / 3)
	// short
	if o.Short != "" {
		s.WriteByte('-')
		s.WriteString(o.Short)
	} else {
		s.WriteString("  ")
	}
	// comma
	if o.Long != "" && o.Short != "" {
		s.WriteString(", ")
	} else {
		s.WriteString("  ")
	}
	// long
	if o.Long != "" {
		s.WriteString("--")
		s.WriteString(o.Long)
	}
	// arg name
	if an := o.usgArgName(); an != "" {
		s.WriteString("  ")
		s.WriteString(an)
	}
	return s.String()
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

// QuotedPlainNames returns the option's long and / or short name(s) in double quotes and
// separated by a comma. For example, the default help option would return `"h", "help"`.
func (o *option) QuotedPlainNames() string {
	long := o.Long
	if long != "" {
		long = `"` + long + `"`
	}
	short := o.Short
	if short != "" {
		short = `"` + short + `"`
	}
	comma := ""
	if o.Long != "" && o.Short != "" {
		comma = ", "
	}
	return long + comma + short
}

// Usg returns a command's usage message text given how wide the name column should be.
func (c *command) Usg(nameWidth int) string {
	paddedName := fmt.Sprintf("   %-*s   ", nameWidth, c.UsgName())
	return paddedName + wrapBlurb(c.Data.Blurb, len(paddedName), maxUsgLineLen)
}

// HasReqArgSomewhere returns true if this command or one of its subcommands contains a
// required positional argument.
func (c *command) HasReqArgSomewhere() bool {
	for _, a := range c.Args {
		if a.IsRequired() {
			return true
		}
	}
	for _, ch := range c.Subcmds {
		if ch.HasReqArgSomewhere() {
			return true
		}
	}
	return false
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

func (c *command) IsRoot() bool     { return c.FieldName == "%[1]s" }
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
