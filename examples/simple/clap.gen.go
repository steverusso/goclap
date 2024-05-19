// generated by goclap; DO NOT EDIT

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

type clapCommand struct {
	usage func() string
	opts  []clapInput
	args  []clapInput
}

type clapInput struct {
	name     string
	value    flag.Value
	required bool
}

func clapFatalf(cmdName, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "error: %s.\nRun '%s -h' for usage.\n", msg, cmdName)
	os.Exit(2)
}

func (cc *clapCommand) parse(args []string) ([]string, error) {
	f := flag.FlagSet{Usage: func() {}}
	f.SetOutput(io.Discard)
	for i := range cc.opts {
		o := &cc.opts[i]
		f.Var(o.value, o.name, "")
	}

	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Println(cc.usage())
			os.Exit(0)
		}
		return nil, err
	}

	rest := f.Args()

	if len(cc.args) > 0 {
		for i := range cc.args {
			arg := &cc.args[i]
			if len(rest) <= i {
				if arg.required {
					return nil, fmt.Errorf("missing required arg '%s'", arg.name)
				}
				return nil, nil
			}
			if err := arg.value.Set(rest[i]); err != nil {
				return nil, fmt.Errorf("parsing positional argument '%s': %v", arg.name, err)
			}
		}
		return nil, nil
	}

	return rest, nil
}

type clapBool bool

func clapNewBool(p *bool) *clapBool { return (*clapBool)(p) }

func (v *clapBool) String() string { return strconv.FormatBool(bool(*v)) }

func (v *clapBool) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return fmt.Errorf(`invalid boolean value "%s"`, s)
	}
	*v = clapBool(b)
	return err
}

func (*clapBool) IsBoolFlag() bool { return true }

type clapString string

func clapNewString(p *string) *clapString { return (*clapString)(p) }

func (v *clapString) String() string { return string(*v) }

func (v *clapString) Set(s string) error {
	*v = clapString(s)
	return nil
}

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
	p := clapCommand{
		usage: c.UsageHelp,
		opts: []clapInput{
			{name: "upper", value: clapNewBool(&c.toUpper)},
		},
		args: []clapInput{
			{name: "<input>", value: clapNewString(&c.input), required: true},
		},
	}
	_, err := p.parse(args)
	if err != nil {
		clapFatalf("mycli", err.Error())
	}
}
