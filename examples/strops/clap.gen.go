// generated by goclap; DO NOT EDIT

package main

import (
	"os"

	"github.com/steverusso/goclap/clap"
)

func (*strops) UsageHelp() string {
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

func (c *strops) Parse(args []string) {
	if len(args) > 0 && len(args) == len(os.Args) {
		args = args[1:]
	}
	p := clap.NewCommandParser("strops")
	p.CustomUsage = c.UsageHelp
	p.Flag("upper", clap.NewBool(&c.toUpper))
	p.Flag("reverse", clap.NewBool(&c.reverse))
	p.Flag("repeat", clap.NewInt(&c.repeat))
	p.Flag("prefix", clap.NewString(&c.prefix))
	p.Arg("<input>", clap.NewString(&c.input)).Require()
	p.Parse(args)
}
