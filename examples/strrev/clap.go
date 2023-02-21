// This file is generated via 'go generate'; DO NOT EDIT
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
	claperr("not enough args: no \033[1;33m%s\033[0m provided\n", name)
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

func (*mycli) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s - reverse a string and maybe make it uppercase

usage:
   %[1]s [options] <input>

options:
   --upper, -u   make the input string all uppercase
   --help, -h    show this help message

arguments:
   <input>   the string to reverse
`, os.Args[0])
}

func (c *mycli) parse(args []string) {
	if len(args) > 0 && len(args) == len(os.Args) {
		args = args[1:]
	}
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "upper", "u":
			c.toUpper = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<input>")
	}
	c.input = args[0]
}
