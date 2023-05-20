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
	//go:embed tmpls/header.go.tmpl
	headerTmplText string

	//go:embed tmpls/usagefunc.go.tmpl
	usgFnTmplText string

	//go:embed tmpls/parsefunc.go.tmpl
	parseFnTmplText string
)

var parseFuncs = template.FuncMap{
	"add": func(a, b int) int { return a + b },
}

type generator struct {
	pkgName     string
	incVersion  bool
	buf         bytes.Buffer
	usgFnTmpl   *template.Template
	parseFnTmpl *template.Template
}

func newGenerator(pkgName string, incVersion bool) (generator, error) {
	usgFnTmpl, err := template.New("usagefunc").Parse(usgFnTmplText)
	if err != nil {
		return generator{}, fmt.Errorf("parsing template: %w", err)
	}
	parseFnTmpl, err := template.New("parsefunc").Funcs(parseFuncs).Parse(parseFnTmplText)
	if err != nil {
		return generator{}, fmt.Errorf("parsing template: %w", err)
	}
	return generator{
		pkgName:     pkgName,
		incVersion:  incVersion,
		usgFnTmpl:   usgFnTmpl,
		parseFnTmpl: parseFnTmpl,
	}, nil
}

func (g *generator) writeHeader(root *command) error {
	type headerData struct {
		PkgName    string
		IncVersion bool
		Version    string
		RootCmd    *command
	}

	t, err := template.New("header").Parse(headerTmplText)
	if err != nil {
		return fmt.Errorf("parsing header template: %w", err)
	}
	err = t.Execute(&g.buf, headerData{
		PkgName:    g.pkgName,
		IncVersion: g.incVersion,
		Version:    getBuildVersionInfo().String(),
		RootCmd:    root,
	})
	if err != nil {
		return fmt.Errorf("executing header template: %w", err)
	}
	return nil
}

func (g *generator) generate(c *command) error {
	for i := range c.Subcmds {
		if err := g.generate(&c.Subcmds[i]); err != nil {
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
	if c.HasArgs() {
		for _, arg := range c.Args {
			argsSlot += " " + arg.UsgName()
		}
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
	ww := wordWrapper{}
	paras := c.Data.overview
	var s strings.Builder
	for i := range paras {
		s.WriteString(ww.wrap(paras[i], 3, 90))
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
			})
		}
	}
	for i := range c.Args {
		if name, ok := c.Args[i].data.getConfig("env"); ok {
			envs = append(envs, clapEnvValue{
				VarName:   name,
				FieldName: c.Args[i].FieldName,
			})
		}
	}
	return envs
}

func (c *command) RequiredArgs() []argument {
	reqs := make([]argument, 0, len(c.Args))
	for _, arg := range c.Args {
		if arg.IsRequired() {
			reqs = append(reqs, arg)
		}
	}
	return reqs
}

func (c *command) HasNonHelpOpt() bool {
	for _, o := range c.Opts {
		if o.Long != "help" {
			return true
		}
	}
	return false
}

func (arg *argument) UsgName() string {
	if arg.IsRequired() {
		return "<" + arg.name + ">"
	}
	return "[" + arg.name + "]"
}

func (arg *argument) IsRequired() bool {
	_, ok := arg.data.getConfig("arg_required")
	return ok
}

// Usg returns an argument's usage message text given how wide the name column should be.
func (a *argument) Usg(nameWidth int) string {
	var envName string
	if v, ok := a.data.getConfig("env"); ok {
		envName = " [$" + v + "]"
	}
	return fmt.Sprintf("%-*s   %s%s", nameWidth, a.UsgName(), a.data.Blurb, envName)
}

// Usg returns an option's usage message text given how wide the name column should be.
func (o *option) Usg(nameWidth int) string {
	var envName string
	if v, ok := o.data.getConfig("env"); ok {
		envName = " [$" + v + "]"
	}
	return fmt.Sprintf("%-*s   %s%s", nameWidth, o.usgNamesAndArg(), o.data.Blurb, envName)
}

func (o *option) usgNamesAndArg() string {
	argName := o.usgArgName()

	s := strings.Builder{}
	s.Grow(len(o.Short) + len(o.Long) + len(argName) + 4)
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
	if argName != "" {
		s.WriteString("  ")
		s.WriteString(argName)
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

// HasReqArgSomewhere returns true if this command or one of its subcommands contains a
// required positional argument.
func (c *command) HasReqArgSomewhere() bool {
	if c.HasRequiredArgs() {
		return true
	}
	for _, ch := range c.Subcmds {
		if ch.HasReqArgSomewhere() {
			return true
		}
	}
	return false
}

func (c *command) HasRequiredArgs() bool {
	for _, arg := range c.Args {
		if arg.IsRequired() {
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
func (c *command) HasOptions() bool { return len(c.Opts) > 0 }
func (c *command) HasArgs() bool    { return len(c.Args) > 0 }

type wordWrapper struct {
	indent string
	word   strings.Builder
	line   strings.Builder
	result strings.Builder
}

func (ww *wordWrapper) wrap(v string, indentWidth, width int) string {
	ww.indent = strings.Repeat(" ", indentWidth)
	ww.word.Grow(width)
	ww.line.Grow(width)
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
		if ww.line.Len()+ww.word.Len() > width {
			ww.takeLineAndReset()
		}
		ww.takeWordAndReset()
		ww.line.WriteRune(c)
	}
	if ww.word.Len() > 0 {
		if ww.line.Len()+ww.word.Len() > width {
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
	ww.result.WriteString(ww.line.String())
	ww.result.WriteRune('\n')
	ww.line.Reset()
	ww.line.WriteString(ww.indent)
}
