// generated by goclap; DO NOT EDIT

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

type clapParser struct {
	usg  func(to *os.File)
	args []string
	idx  int

	optName  string
	optEqVal string
	optHasEq bool
}

func (p *clapParser) exitUsgGood() {
	p.usg(os.Stdout)
	os.Exit(0)
}

func (p *clapParser) stageOpt() bool {
	if p.optName != "" {
		p.idx++
	}
	if p.idx > len(p.args)-1 {
		p.optName = ""
		return false
	}
	arg := p.args[p.idx]
	if arg[0] != '-' {
		p.optName = ""
		return false
	}
	arg = arg[1:]
	if arg == "" {
		claperr("emtpy option ('-') found\n")
		os.Exit(1)
	}
	if arg[0] == '-' {
		arg = arg[1:]
	}
	if arg == "" {
		p.idx++
		p.optName = ""
		return false
	}

	p.optEqVal = ""
	if eqIdx := strings.IndexByte(arg, '='); eqIdx != -1 {
		p.optName = arg[:eqIdx]
		if eqIdx < len(arg) {
			p.optEqVal = arg[eqIdx+1:]
		}
		p.optHasEq = true
	} else {
		p.optName = arg
		p.optHasEq = false
	}
	return true
}

func (p *clapParser) nextStr() string {
	if p.optName != "" {
		if p.optHasEq {
			return p.optEqVal
		}
		if p.idx == len(p.args)-1 {
			claperr("option '%s' needs an argument\n", p.optName)
			os.Exit(1)
		}
		p.idx++
		return p.args[p.idx]
	}
	p.idx++
	return p.args[p.idx-1]
}

func (p *clapParser) exitBadInput(typ string, err error) {
	var forWhat string
	switch {
	case p.optName != "":
		forWhat = "option '" + p.optName + "'"
	default:
		forWhat = "argument"
	}
	claperr("invalid %s for %s: %v\n", typ, forWhat, err)
	os.Exit(1)
}

func (p *clapParser) thisBool() bool {
	s := p.optEqVal
	if s == "" || s == "true" || s == "1" {
		return true
	}
	if s != "false" && s != "0" {
		p.exitBadInput("bool", fmt.Errorf("%q not recognized as a boolean", s))
	}
	return false
}

func (*goclap) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s - pre-build tool to generate command line argument parsing code from Go comments

usage:
   %[1]s [options]

options:
   -v, --version           print version info and exit
       --include-version   include the version info in the generated code
       --type  <arg>       the root command struct name
       --srcdir  <arg>     directory of source files to parse (default ".")
   -o, --out  <arg>        output file path (default "./clap.gen.go")
   -h, --help              show this help message
`, os.Args[0])
}

func (c *goclap) parse(args []string) {
	if len(args) > 0 && len(args) == len(os.Args) {
		args = args[1:]
	}
	p := clapParser{usg: c.printUsage, args: args}
	for p.stageOpt() {
		switch p.optName {
		case "version", "v":
			c.version = p.thisBool()
		case "include-version":
			c.incVersion = p.thisBool()
		case "type":
			c.rootCmdType = p.nextStr()
		case "srcdir":
			c.srcDir = p.nextStr()
		case "out", "o":
			c.outFilePath = p.nextStr()
		case "help", "h":
			p.exitUsgGood()
		default:
			claperr("unknown option '%s'\n", p.optName)
			os.Exit(1)
		}
	}
}
