// Code generated by goclap ((devel)-20230226-b9dc42917d41-(with unstaged changes)); DO NOT EDIT
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

func clapParseBool(s string) bool {
	if s == "" || s == "true" {
		return true
	}
	if s != "false" {
		claperr("invalid boolean value '%%s'\n", s)
		os.Exit(1)
	}
	return false
}

func optParts(arg string) (string, string, bool) {
	if arg == "-" {
		claperr("emtpy option ('-') found\n")
		os.Exit(1)
	}
	if arg[0] == '-' {
		arg = arg[1:]
	}
	if arg[0] == '-' {
		arg = arg[1:]
	}
	if eqIdx := strings.IndexByte(arg, '='); eqIdx != -1 {
		name := arg[:eqIdx]
		eqVal := ""
		if eqIdx < len(arg) {
			eqVal = arg[eqIdx+1:]
		}
		return name, eqVal, true
	}
	return arg, "", false
}

func (*strops) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s - perform different string operations

usage:
   %[1]s [options] <input>

options:
   --upper, -u    make the input string all uppercase
   --prefix, -p   add this prefix to the final string
   --help, -h     show this help message

arguments:
   <input>   the string on which to operate
`, os.Args[0])
}

func (c *strops) parse(args []string) {
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
		k, eqv, hasEq := optParts(args[i][1:])

		switch k {
		case "upper", "u":
			c.toUpper = clapParseBool(eqv)
		case "prefix", "p":
			if hasEq {
				c.prefix = eqv
			} else if i == len(args)-1 {
				claperr("string option '%s' needs a value\n", k)
				os.Exit(1)
			} else {
				i++
				c.prefix = args[i]
			}
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
