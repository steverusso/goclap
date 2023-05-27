// generated by goclap; DO NOT EDIT

package main

import (
	"fmt"
	"os"
	"strconv"
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

func clapParseBool(s string) bool {
	if s == "" || s == "true" {
		return true
	}
	if s != "false" {
		claperr("invalid boolean value '%s'\n", s)
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

type clapOpt struct {
	long  string
	short string
	v     any
}

func parseOpts(args []string, u clapUsagePrinter, data []clapOpt) int {
	var i int
argsLoop:
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv, hasEq := optParts(args[i][1:])
		for z := range data {
			if k == data[z].long || k == data[z].short {
				if v, ok := data[z].v.(*bool); ok {
					*v = clapParseBool(eqv)
				} else {
					var val string
					if hasEq {
						val = eqv
					} else if i == len(args)-1 {
						claperr("option '%s' needs an argument\n", k)
						os.Exit(1)
					} else {
						i++
						val = args[i]
					}
					err := clapParseInto(data[z].v, val)
					if err != nil {
						claperr("invalid argument for option '%s': %v\n", k, err)
						os.Exit(1)
					}
				}
				continue argsLoop
			}
		}
		if k == "h" || k == "help" {
			u.printUsage(os.Stdout)
			os.Exit(0)
		}
		claperr("unknown option '%s'\n", k)
		os.Exit(1)
	}
	return i
}

func clapParseInto(v any, s string) error {
	if v, ok := v.(*string); ok {
		*v = s
		return nil
	}
	var (
		i64 int64
		u64 uint64
		err error
	)
	switch v := v.(type) {
	case *int:
		i64, err = strconv.ParseInt(s, 10, 0)
		*v = int(i64)
	case *int8:
		i64, err = strconv.ParseInt(s, 10, 8)
		*v = int8(i64)
	case *int16:
		i64, err = strconv.ParseInt(s, 10, 16)
		*v = int16(i64)
	case *int32:
		i64, err = strconv.ParseInt(s, 10, 32)
		*v = int32(i64)
	case *int64:
		*v, err = strconv.ParseInt(s, 10, 64)
	case *uint:
		u64, err = strconv.ParseUint(s, 10, 0)
		*v = uint(u64)
	case *uint8:
		u64, err = strconv.ParseUint(s, 10, 8)
		*v = uint8(u64)
	case *uint16:
		u64, err = strconv.ParseUint(s, 10, 16)
		*v = uint16(u64)
	case *uint32:
		u64, err = strconv.ParseUint(s, 10, 32)
		*v = uint32(u64)
	case *uint64:
		*v, err = strconv.ParseUint(s, 10, 64)
	case *uintptr:
		u64, err = strconv.ParseUint(s, 10, 64)
		*v = uintptr(u64)
	case *float32:
		var f float64
		f, err = strconv.ParseFloat(s, 32)
		*v = float32(f)
	case *float64:
		*v, err = strconv.ParseFloat(s, 64)
	}
	return err
}

func (*strops) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s - perform different string operations

usage:
   %[1]s [options] <input>

options:
   -u, --upper           make the `+"`"+`input`+"`"+` string all uppercase
   -r, --reverse         reverse the final string
       --repeat  <n>     repeat the string this many times
       --prefix  <str>   add this prefix to the final string
   -h, --help            show this help message

arguments:
   <input>   the string on which to operate
`, os.Args[0])
}

func (c *strops) parse(args []string) {
	if len(args) > 0 && len(args) == len(os.Args) {
		args = args[1:]
	}
	i := parseOpts(args, c, []clapOpt{
		{"upper", "u", &c.toUpper},
		{"reverse", "r", &c.reverse},
		{"repeat", "", &c.repeat},
		{"prefix", "", &c.prefix},
	})
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<input>")
	}
	clapParseInto(&c.input, args[0])
}
