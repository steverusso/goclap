package main

import (
	"fmt"
	"os"
	"strings"
)

type generator struct {
	out *os.File
}

func (g *generator) generate(c *command, isRoot bool) {
	for i := range c.subcmds {
		g.generate(&c.subcmds[i], false)
	}
	g.writePrintUsageFunc(c)
	g.writeCmdParseFunc(c, isRoot)
}

func (g *generator) writePrintUsageFunc(c *command) {
	g.printf("\nfunc (*%s) printUsage(to *os.File) {\n", c.typeName)
	g.printf("\tfmt.Fprintf(to, `")
	g.writeOverview(c)
	g.writeUsageLines(c)
	g.writeOpts(c)
	g.writeArgs(c)
	g.writeSubcmds(c)
	g.printf("`, os.Args[0])\n")
	g.printf("}\n\n")
}

func (g *generator) writeOverview(c *command) {
	parents := strings.Join(c.parentNames, " ")
	if parents != "" {
		parents += " "
	}
	g.printf("%s%s - %s\n\n", parents, c.docName(), c.data.blurb)
}

func (g *generator) writeUsageLines(c *command) {
	g.printf("usage:\n   ")
	if uargs, ok := c.data.getConfig("cmd_usage"); ok {
		g.printf("%s %s\n", c.docName(), uargs)
	} else {
		optionsSlot := " [options]" // Every command has at least the help options for now.
		commandSlot := ""
		if c.hasSubcmds() {
			commandSlot = " <command>"
		}
		argsSlot := ""
		if c.hasArgs() {
			for _, arg := range c.args {
				argsSlot += " " + arg.docString()
			}
		}
		g.printf("%s%s%s%s\n", c.docName(), optionsSlot, commandSlot, argsSlot)
	}
}

func (g *generator) writeSubcmds(c *command) {
	if !c.hasSubcmds() {
		return
	}
	w := 0
	for _, sc := range c.subcmds {
		name := sc.docName()
		if l := len(name); l > w {
			w = l
		}
	}
	g.printf("\ncommands:\n")
	for _, sc := range c.subcmds {
		g.printf("   %-*s   %s\n", w, sc.docName(), sc.data.blurb)
	}
}

func (g *generator) writeArgs(c *command) {
	if !c.hasArgs() {
		return
	}
	w := 0
	for _, arg := range c.args {
		if l := len(arg.docString()); l > w {
			w = l
		}
	}
	g.printf("\narguments:\n")
	for _, arg := range c.args {
		g.printf("   %-*s   %s\n", w, arg.docString(), arg.data.blurb)
	}
}

func (g *generator) writeOpts(c *command) {
	if !c.hasOptions() {
		return
	}
	w := 0
	for _, o := range c.opts {
		if l := len(o.docNames()); l > w {
			w = l
		}
	}
	g.printf("\noptions:\n")
	for _, o := range c.opts {
		g.printf("   %-*s   %s\n", w, o.docNames(), o.data.blurb)
	}
}

func (g *generator) writeCmdParseFunc(c *command, isRoot bool) {
	g.printf("func (c *%s) parse(args []string) {\n", c.typeName)

	if isRoot {
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
		g.printf("\t\tk, eqv := optParts(args[i][1:])\n")
	} else {
		g.printf("\t\tk, _ := optParts(args[i][1:])\n")
	}
	g.printf("\t\tswitch k {\n")
	for _, opt := range c.opts {
		g.printf("\t\tcase %s:\n", opt.quotedPlainNames())
		// Hard code the 'help' case.
		if opt.long == "help" {
			g.printf("\t\t\texitUsgGood(c)\n")
			continue
		}
		switch opt.fieldType {
		case typBool:
			g.printf("\t\t\tc.%s = (eqv == \"\" || eqv == \"true\")\n", opt.fieldName)
		case typString:
		}
	}
	g.printf("\t\t}\n") // end switch
	g.printf("\t}\n")   // end loop

	// Arguments.
	if c.hasArgs() {
		g.printf("\targs = args[i:]\n")
		// Add error handling for missing arguments that are required.
		reqArgs := c.requiredArgs()
		for i := range reqArgs {
			g.printf("\tif len(args) < %d {\n", i+1)
			g.printf("\t\texitMissingArg(c, %q)\n", reqArgs[i].docString())
			g.printf("\t}\n")
		}
		for i, arg := range c.args {
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
	if c.hasSubcmds() {
		g.printf("\tif i >= len(args) {\n")
		g.printf("\t\tc.printUsage(os.Stderr)\n")
		g.printf("\t\tos.Exit(1)\n")
		g.printf("\t}\n")

		g.printf("\tswitch args[i] {\n")

		for _, sc := range c.subcmds {
			g.printf("\tcase %q:\n", sc.docName())
			g.printf("\t\tc.%s = new(%s)\n", sc.fieldName, sc.typeName)
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
	reqs := make([]argInfo, 0, len(c.args))
	for _, arg := range c.args {
		if arg.isRequired() {
			reqs = append(reqs, arg)
		}
	}
	return reqs
}

func (c *command) hasOptField() bool {
	for _, o := range c.opts {
		if o.long != "help" {
			return true
		}
	}
	return false
}

func (c *command) hasSubcmds() bool { return len(c.subcmds) > 0 }
func (c *command) hasOptions() bool { return len(c.opts) > 0 }
func (c *command) hasArgs() bool    { return len(c.args) > 0 }

func (arg *argInfo) docString() string {
	if arg.isRequired() {
		return "<" + arg.name + ">"
	}
	return "[" + arg.name + "]"
}

func (arg *argInfo) isRequired() bool {
	_, ok := arg.data.getConfig("arg_required")
	return ok
}

func (o *optInfo) docNames() string {
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

func (g *generator) writeHeader(hasSubcmds bool) {
	g.printf("// Code generated by goclap (%s); DO NOT EDIT", getBuildVersionInfo())

	g.printf(outFileHeader)
	if hasSubcmds {
		g.printf(`
func exitUnknownCmd(u clapUsagePrinter, name string) {
	claperr("unknown command '%%s'\n", name)
	u.printUsage(os.Stderr)
	os.Exit(1)
}
`)
	}
}

const outFileHeader = `
package main

import (
	"fmt"
	"os"
	"strings"
)

func claperr(format string, a ...any) {
	format = "\033[1;31merror:\033[0m " + format
	fmt.Fprintf(os.Stderr, format, a...)
}

func exitEmptyOpt() {
	claperr("emtpy option ('-') found\n")
	os.Exit(1)
}

type clapUsagePrinter interface {
	printUsage(to *os.File)
}

func exitMissingArg(u clapUsagePrinter, name string) {
	claperr("not enough args: no \033[1;33m%%s\033[0m provided\n", name)
	u.printUsage(os.Stderr)
	os.Exit(1)
}

func exitUsgGood(u clapUsagePrinter) {
	u.printUsage(os.Stdout)
	os.Exit(0)
}

func optParts(arg string) (string, string) {
	if arg == "-" {
		exitEmptyOpt()
	}
	if arg[0] == '-' {
		arg = arg[1:]
	}
	if arg[0] == '-' {
		arg = arg[1:]
	}
	name := arg
	eqVal := ""
	if eqIdx := strings.IndexByte(name, '='); eqIdx != -1 {
		name = arg[:eqIdx]
		eqVal = arg[eqIdx+1:]
	}
	return name, eqVal
}
`
