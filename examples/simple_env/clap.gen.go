// generated by goclap; DO NOT EDIT

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
)

type clapCommand struct {
	usage func() string
	opts  []clapInput
	args  []clapInput
}

type clapInput struct {
	name     string
	envName  string
	value    flag.Value
	required bool
}

func (in *clapInput) parseEnv() error {
	if in.envName == "" {
		return nil
	}
	s, ok := os.LookupEnv(in.envName)
	if !ok {
		return nil
	}
	if err := in.value.Set(s); err != nil {
		return fmt.Errorf("parsing env var '%s': %w", in.envName, err)
	}
	return nil
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
		if err := o.parseEnv(); err != nil {
			return nil, err
		}
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
			if err := arg.parseEnv(); err != nil {
				return nil, err
			}
		}
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

type clapString string

func clapNewString(p *string) *clapString { return (*clapString)(p) }

func (v *clapString) String() string { return string(*v) }

func (v *clapString) Set(s string) error {
	*v = clapString(s)
	return nil
}

type clapUint[T uint | uint8 | uint16 | uint32 | uint64] struct{ v *T }

func clapNewUint[T uint | uint8 | uint16 | uint32 | uint64](p *T) clapUint[T] { return clapUint[T]{p} }

func (v clapUint[T]) String() string { return strconv.FormatUint(uint64(*v.v), 10) }

func (v clapUint[T]) Set(s string) error {
	u64, err := strconv.ParseUint(s, 0, reflect.TypeFor[T]().Bits())
	if err != nil {
		return numError(err)
	}
	*v.v = T(u64)
	return nil
}

func numError(err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		return ne.Err
	}
	return err
}

func (*mycli) UsageHelp() string {
	return `simple_env - Print a string with a prefix

usage:
   simple_env [options] [input]

options:
   -prefix  <arg>   The value to prepend to the input string [$MY_PREFIX]
   -count  <arg>    Print the output this many extra times [$MY_COUNT]
   -h               Show this help message

arguments:
   [input]   The user provided input [$MY_INPUT]`
}

func (c *mycli) Parse(args []string) {
	p := clapCommand{
		usage: c.UsageHelp,
		opts: []clapInput{
			{name: "prefix", value: clapNewString(&c.prefix), envName: "MY_PREFIX"},
			{name: "count", value: clapNewUint(&c.count), envName: "MY_COUNT"},
		},
		args: []clapInput{
			{name: "[input]", value: clapNewString(&c.input), envName: "MY_INPUT"},
		},
	}
	_, err := p.parse(args)
	if err != nil {
		clapFatalf("simple_env", err.Error())
	}
}
