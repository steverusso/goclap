// generated by goclap; DO NOT EDIT

package main

import "github.com/steverusso/goclap/clap"

func (*mycli) UsageHelp() string {
	return `mycli - Print a string with a prefix

usage:
   mycli [options] [input]

options:
   -prefix  <arg>   The value to prepend to the input string [$MY_PREFIX]
   -count  <arg>    Print the output this many extra times [$MY_COUNT]
   -h               Show this help message

arguments:
   [input]   The user provided input [$MY_INPUT]`
}

func (c *mycli) Parse(args []string) {
	p := clap.NewCommandParser("mycli")
	p.CustomUsage = c.UsageHelp
	p.Flag("prefix", clap.NewString(&c.prefix)).Env("MY_PREFIX")
	p.Flag("count", clap.NewUint(&c.count)).Env("MY_COUNT")
	p.Arg("[input]", clap.NewString(&c.input)).Env("MY_INPUT")
	p.Parse(args)
}
