package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"
)

//go:embed tmpls/header.go.tmpl
var headerTmplText string

//go:embed tmpls/usagefunc.go.tmpl
var usgFnTmplText string

type generator struct {
	out       *os.File
	usgFnTmpl *template.Template
}

func newGenerator(out *os.File) (generator, error) {
	usgFnTmpl, err := template.New("usagefunc").Parse(usgFnTmplText)
	if err != nil {
		return generator{}, fmt.Errorf("parsing usage func template: %w", err)
	}
	return generator{
		out:       out,
		usgFnTmpl: usgFnTmpl,
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
	g.writeCmdParseFunc(c)
	return nil
}

func (g *generator) genCmdUsageFunc(c *command) error {
	err := g.usgFnTmpl.Execute(g.out, c)
	if err != nil {
		return fmt.Errorf("executing template: %w", err)
	}
	return nil
}

func (c *command) IsRoot() bool { return c.fieldName == "%[1]s" }

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

func (g *generator) writeCmdParseFunc(c *command) {
	g.printf("func (c *%s) parse(args []string) {\n", c.TypeName)

	if c.IsRoot() {
		// Drop the program name from args if it's there.
		g.printf("\tif len(args) > 0 && len(args) == len(os.Args) {\n")
		g.printf("\t\targs = args[1:]\n")
		g.printf("\t}\n")
	}

	// Parse options.
	g.printf(`	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
`)
	if c.hasOptField() {
		g.printf("\t\tk, eqv, _ := optParts(args[i][1:])\n")
	} else {
		g.printf("\t\tk, _, _ := optParts(args[i][1:])\n")
	}
	g.printf("\t\tswitch k {\n")
	for _, opt := range c.Opts {
		g.printf("\t\tcase %s:\n", opt.quotedPlainNames())
		// Hard code the 'help' case.
		if opt.long == "help" {
			g.printf("\t\t\texitUsgGood(c)\n")
			continue
		}
		switch opt.fieldType {
		case typBool:
			g.printf("\t\t\tc.%s = parseBool(eqv)\n", opt.fieldName)
		case typString:
		}
	}
	g.printf("\t\t}\n") // end switch
	g.printf("\t}\n")   // end loop

	// Arguments.
	if c.HasArgs() {
		g.printf("\targs = args[i:]\n")
		// Add error handling for missing arguments that are required.
		reqArgs := c.requiredArgs()
		for i := range reqArgs {
			g.printf("\tif len(args) < %d {\n", i+1)
			g.printf("\t\texitMissingArg(c, %q)\n", reqArgs[i].DocString())
			g.printf("\t}\n")
		}
		for i, arg := range c.Args {
			if !arg.isRequired() {
				g.printf("\tif len(args) < %d {\n", i+1)
				g.printf("\t\treturn\n")
				g.printf("\t}\n")
			}
			// Parse positional args based on their type.
			switch arg.fieldType {
			case typString:
				g.printf("\tc.%s = args[%d]\n", arg.fieldName, i)
			default:
				panic(fmt.Sprintf("unsupported arg type '%d' made it to generator", arg.fieldType))
			}
		}
	}

	// Subcommands.
	if c.HasSubcmds() {
		g.printf("\tif i >= len(args) {\n")
		g.printf("\t\tc.printUsage(os.Stderr)\n")
		g.printf("\t\tos.Exit(1)\n")
		g.printf("\t}\n")

		g.printf("\tswitch args[i] {\n")

		for _, sc := range c.Subcmds {
			g.printf("\tcase %q:\n", sc.DocName())
			g.printf("\t\tc.%s = new(%s)\n", sc.fieldName, sc.TypeName)
			g.printf("\t\tc.%s.parse(args[i+1:])\n", sc.fieldName)
		}

		// Default care which means an unknown command.
		g.printf("\tdefault:\n")
		g.printf("\t\texitUnknownCmd(c, args[i])\n")
		g.printf("\t}\n") // end switch
	}

	g.printf("}\n") // Closing curly bracket for parse func.
}

func (c *command) requiredArgs() []argInfo {
	reqs := make([]argInfo, 0, len(c.Args))
	for _, arg := range c.Args {
		if arg.isRequired() {
			reqs = append(reqs, arg)
		}
	}
	return reqs
}

func (c *command) hasOptField() bool {
	for _, o := range c.Opts {
		if o.long != "help" {
			return true
		}
	}
	return false
}

func (c *command) HasSubcmds() bool { return len(c.Subcmds) > 0 }
func (c *command) HasOptions() bool { return len(c.Opts) > 0 }
func (c *command) HasArgs() bool    { return len(c.Args) > 0 }

func (arg *argInfo) DocString() string {
	if arg.isRequired() {
		return "<" + arg.name + ">"
	}
	return "[" + arg.name + "]"
}

func (arg *argInfo) isRequired() bool {
	_, ok := arg.data.getConfig("arg_required")
	return ok
}

func (o *optInfo) DocNames() string {
	long := o.long
	if long != "" {
		long = "--" + long
	}
	short := o.short
	if short != "" {
		short = "-" + short
	}
	comma := ""
	if o.long != "" && o.short != "" {
		comma = ", "
	}
	return fmt.Sprintf("%s%s%s", long, comma, short)
}

func (o *optInfo) quotedPlainNames() string {
	long := o.long
	if long != "" {
		long = "\"" + long + "\""
	}
	short := o.short
	if short != "" {
		short = "\"" + short + "\""
	}
	comma := ""
	if o.long != "" && o.short != "" {
		comma = ", "
	}
	return fmt.Sprintf("%s%s%s", long, comma, short)
}

func (g *generator) printf(format string, a ...any) {
	fmt.Fprintf(g.out, format, a...)
}
