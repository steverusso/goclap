// generated by goclap; DO NOT EDIT

package main

import (
	"os"

	"github.com/steverusso/goclap/clap"
)

func (*mycli) UsageHelp() string {
	return `mycli - Print a string with the option to make it uppercase

usage:
   mycli [options] <input>

options:
   -upper   Make the input string all uppercase
   -h       Show this help message

arguments:
   <input>   The input string`
}

func (c *mycli) Parse(args []string) {
	if len(args) > 0 && len(args) == len(os.Args) {
		args = args[1:]
	}

	cc := clap.NewCommandParser("mycli")
	cc.CustomUsage = c.UsageHelp
	cc.Flag("upper", clap.NewBool(&c.toUpper))
	cc.Arg("<input>", clap.NewString(&c.input)).Require()
	cc.Parse(args)
}
