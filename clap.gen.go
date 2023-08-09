// generated by goclap; DO NOT EDIT

package main

import (
	"os"

	"github.com/steverusso/goclap/clap"
)

func (*goclap) UsageHelp() string {
	return `goclap - Pre-build tool to generate command line argument parsing code from Go comments

usage:
   goclap [options]

options:
   -type  <arg>       The root command struct name
   -srcdir  <arg>     Directory of source files to parse (default ".")
   -include-version   Include goclap's version info in the generated code
   -out  <arg>        Output file path (default "./clap.gen.go")
   -version           Print version info and exit
   -h                 Show this help message`
}

func (c *goclap) Parse(args []string) {
	if len(args) > 0 && len(args) == len(os.Args) {
		args = args[1:]
	}

	p := clap.NewCommandParser("goclap")
	p.CustomUsage = c.UsageHelp
	p.Flag("type", clap.NewString(&c.rootCmdType))
	p.Flag("srcdir", clap.NewString(&c.srcDir))
	p.Flag("include-version", clap.NewBool(&c.incVersion))
	p.Flag("out", clap.NewString(&c.outFilePath))
	p.Flag("version", clap.NewBool(&c.version))
	p.Parse(args)
}
