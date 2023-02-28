package main

import (
	_ "embed"
	"fmt"
	"os"
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
		return generator{}, fmt.Errorf("parsing template: %w", err)
	}
	parseFnTmpl, err := template.New("parsefunc").Funcs(parseFuncs).Parse(parseFnTmplText)
	if err != nil {
		return generator{}, fmt.Errorf("parsing template: %w", err)
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
		return fmt.Errorf("'%s': %w", c.TypeName, err)
	}
	if err := g.genCmdParseFunc(c); err != nil {
		return fmt.Errorf("'%s': %w", c.TypeName, err)
	}
	return nil
}

func (g *generator) genCmdUsageFunc(c *command) error {
	err := g.usgFnTmpl.Execute(g.out, c)
	if err != nil {
		return err
	}
	return nil
}

func (g *generator) genCmdParseFunc(c *command) error {
	err := g.parseFnTmpl.Execute(g.out, c)
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

func (c *command) OptNameColWidth() int {
	w := 0
	for _, o := range c.Opts {
		if l := len(o.UsgNames()); l > w {
			w = l
		}
	}
	return w
}

// OptArgNameColWidth returns the length of the longest option argument name out of this
// command's options. In a usage message, the argument name column is in between the
// option's name(s) and description columns.
func (c *command) OptArgNameColWidth() int {
	w := 0
	for _, o := range c.Opts {
		if !o.FieldType.IsBool() {
			if l := len(o.UsgArgName()); l > w {
				w = l
			}
		}
	}
	return w
}

func (c *command) ArgNamesColWidth() int {
	w := 0
	for _, a := range c.Args {
		if l := len(a.UsgName()); l > w {
			w = l
		}
	}
	return w
}

func (c *command) SubcmdNamesColWidth() int {
	w := 0
	for _, sc := range c.Subcmds {
		if l := len(sc.UsgName()); l > w {
			w = l
		}
	}
	return w
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

func (c *command) NeedsVarEqv() bool {
	for _, o := range c.Opts {
		if o.Long != "help" {
			return true
		}
	}
	return false
}

func (c *command) NeedsVarHasEq() bool {
	for _, o := range c.Opts {
		if o.FieldType.IsString() {
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
	_, ok := arg.Data.getConfig("arg_required")
	return ok
}

// UsgArgName returns the usage text of an option argument for non-boolean options. For
// example, if there's a string option named `file`, the usage might look something like
// `--file <arg>` where "<arg>" is the usage argument name text.
func (o *option) UsgArgName() string {
	if o.FieldType.IsBool() {
		return ""
	}
	if name, ok := o.Data.getConfig("opt_arg_name"); ok {
		return "<" + name + ">"
	}
	return "<arg>"
}

func (o *option) UsgNames() string {
	long := o.Long
	if long != "" {
		long = "--" + long
	}
	short := "  "
	if o.Short != "" {
		short = "-" + o.Short
	}
	comma := "  "
	if o.Long != "" && o.Short != "" {
		comma = ", "
	}
	return short + comma + long
}

func (o *option) QuotedPlainNames() string {
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
	return long + comma + short
}

func (c *command) IsRoot() bool     { return c.FieldName == "%[1]s" }
func (c *command) HasSubcmds() bool { return len(c.Subcmds) > 0 }
func (c *command) HasOptions() bool { return len(c.Opts) > 0 }
func (c *command) HasArgs() bool    { return len(c.Args) > 0 }
