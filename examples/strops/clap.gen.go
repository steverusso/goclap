// generated by goclap; DO NOT EDIT

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/steverusso/goclap/clap"
)

func (*strops) usage() string {
	return `strops - Perform different string operations

usage:
   strops [options] <input>

options:
   -upper           Make the ` + "`" + `input` + "`" + ` string all uppercase
   -reverse         Reverse the final string
   -repeat  <n>     Repeat the string this many times
   -prefix  <str>   Add this prefix to the final string
   -h               Show this help message

arguments:
   <input>   The string on which to operate`
}

func (c *strops) parse(args []string) {
	if len(args) > 0 && len(args) == len(os.Args) {
		args = args[1:]
	}

	var err error

	f := flag.FlagSet{Usage: func() {}}
	f.Var(clap.NewBool(&c.toUpper), "upper", "")
	f.Var(clap.NewBool(&c.reverse), "reverse", "")
	f.Var(clap.NewInt(&c.repeat), "repeat", "")
	f.Var(clap.NewString(&c.prefix), "prefix", "")
	if err = f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Println(c.usage())
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "error: %v.\nRun 'strops -h' for usage.", err)
		os.Exit(2)
	}

	rest := f.Args()
	if len(rest) < 1 {
		fmt.Fprintf(os.Stderr, "error: missing arg <input>")
		os.Exit(2)
	}
	if err = clap.NewString(&c.input).Set(rest[0]); err != nil {
		fmt.Fprintf(os.Stderr, "error: parsing arg: %v\n", err)
	}
}
