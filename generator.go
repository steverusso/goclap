package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"
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
	out         *os.File
	usgFnTmpl   *template.Template
	parseFnTmpl *template.Template
}

func newGenerator(out *os.File) (generator, error) {
	usgFnTmpl, err := template.New("usagefunc").Parse(usgFnTmplText)
	if err != nil {
		return generator{}, fmt.Errorf("parsing usage func template: %w", err)
	}
	parseFnTmpl, err := template.New("parsefunc").Funcs(parseFuncs).Parse(parseFnTmplText)
	if err != nil {
		return generator{}, fmt.Errorf("parsing parse func template: %w", err)
	}
	return generator{
		out:         out,
		usgFnTmpl:   usgFnTmpl,
		parseFnTmpl: parseFnTmpl,
	}, nil
}

func (g *generator) writeHeader(hasSubcmds bool) error {
	type headerData struct {
		Version    string
		HasSubcmds bool
	}

	t, err := template.New("header").Parse(headerTmplText)
	if err != nil {
		return fmt.Errorf("parsing header template: %w", err)
	}
	err = t.Execute(g.out, headerData{
		HasSubcmds: hasSubcmds,
		Version:    getBuildVersionInfo().String(),
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
		return fmt.Errorf("usage func for '%s': %w", c.TypeName, err)
	}
	if err := g.genCmdParseFunc(c); err != nil {
		return fmt.Errorf("parse func for '%s': %w", c.TypeName, err)
	}
	return nil
}

func (g *generator) genCmdUsageFunc(c *command) error {
	err := g.usgFnTmpl.Execute(g.out, c)
	if err != nil {
		return fmt.Errorf("executing template: %w", err)
	}
	return nil
}

func (g *generator) genCmdParseFunc(c *command) error {
	err := g.parseFnTmpl.Execute(g.out, c)
	if err != nil {
		return fmt.Errorf("executing template: %w", err)
	}
	return nil
}

func (o *optInfo) IsBool() bool { return o.fieldType == typBool }

func (a *argInfo) IsString() bool { return a.fieldType == typString }

func (c *command) IsRoot() bool { return c.FieldName == "%[1]s" }

func (c *command) Parents() string {
	s := strings.Join(c.parentNames, " ")
	if s != "" {
		s += " "
	}
	return s
}

func (c *command) UsageLines() []string {
	us := make([]string, 0, 2)
	for _, cfg := range c.data.configs {
		if cfg.key == "cmd_usage" {
			us = append(us, c.DocName()+" "+cfg.val)
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
			argsSlot += " " + arg.DocString()
		}
	}
	return []string{
		fmt.Sprintf("%s%s%s%s", c.DocName(), optionsSlot, commandSlot, argsSlot),
	}
}

func (c *command) Blurb() string { return c.data.blurb }
func (o *optInfo) Blurb() string { return o.data.blurb }
func (a *argInfo) Blurb() string { return a.data.blurb }

func (c *command) OptNamesColWidth() int {
	w := 0
	for _, o := range c.Opts {
		if l := len(o.DocNames()); l > w {
			w = l
		}
	}
	return w
}

func (c *command) ArgNamesColWidth() int {
	w := 0
	for _, a := range c.Args {
		if l := len(a.DocString()); l > w {
			w = l
		}
	}
	return w
}

func (c *command) SubcmdNamesColWidth() int {
	w := 0
	for _, sc := range c.Subcmds {
		if l := len(sc.DocName()); l > w {
			w = l
		}
	}
	return w
}

func (c *command) RequiredArgs() []argInfo {
	reqs := make([]argInfo, 0, len(c.Args))
	for _, arg := range c.Args {
		if arg.IsRequired() {
			reqs = append(reqs, arg)
		}
	}
	return reqs
}

func (c *command) HasOptField() bool {
	for _, o := range c.Opts {
		if o.Long != "help" {
			return true
		}
	}
	return false
}

func (c *command) HasSubcmds() bool { return len(c.Subcmds) > 0 }
func (c *command) HasOptions() bool { return len(c.Opts) > 0 }
func (c *command) HasArgs() bool    { return len(c.Args) > 0 }

func (arg *argInfo) DocString() string {
	if arg.IsRequired() {
		return "<" + arg.name + ">"
	}
	return "[" + arg.name + "]"
}

func (arg *argInfo) IsRequired() bool {
	_, ok := arg.data.getConfig("arg_required")
	return ok
}

func (o *optInfo) DocNames() string {
	long := o.Long
	if long != "" {
		long = "--" + long
	}
	short := o.Short
	if short != "" {
		short = "-" + short
	}
	comma := ""
	if o.Long != "" && o.Short != "" {
		comma = ", "
	}
	return fmt.Sprintf("%s%s%s", long, comma, short)
}

func (o *optInfo) QuotedPlainNames() string {
	long := o.Long
	if long != "" {
		long = "\"" + long + "\""
	}
	short := o.Short
	if short != "" {
		short = "\"" + short + "\""
	}
	comma := ""
	if o.Long != "" && o.Short != "" {
		comma = ", "
	}
	return fmt.Sprintf("%s%s%s", long, comma, short)
}

func (g *generator) printf(format string, a ...any) {
	fmt.Fprintf(g.out, format, a...)
}
