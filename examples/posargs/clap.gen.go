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

type clapString string

func clapNewString(p *string) *clapString { return (*clapString)(p) }

func (v *clapString) String() string { return string(*v) }

func (v *clapString) Set(s string) error {
	*v = clapString(s)
	return nil
}

type clapFloat[T float32 | float64] struct{ v *T }

func clapNewFloat[T float32 | float64](p *T) clapFloat[T] { return clapFloat[T]{p} }

func (v clapFloat[T]) String() string {
	return strconv.FormatFloat(float64(*v.v), 'g', -1, reflect.TypeFor[T]().Bits())
}

func (v clapFloat[T]) Set(s string) error {
	f64, err := strconv.ParseFloat(s, reflect.TypeFor[T]().Bits())
	if err != nil {
		return numError(err)
	}
	*v.v = T(f64)
	return err
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
	return `posargs - Print a few positional args

usage:
   posargs [options] <f32> <text> <u16>

options:
   -h   Show this help message

arguments:
   <f32>    A float32 positional arg
   <text>   A string positional arg
   <u16>    A uint16 positional arg`
}

func (c *mycli) Parse(args []string) {
	p := clapCommand{
		usage: c.UsageHelp,
		args: []clapInput{
			{name: "<f32>", value: clapNewFloat(&c.f32), required: true},
			{name: "<text>", value: clapNewString(&c.str), required: true},
			{name: "<u16>", value: clapNewUint(&c.u16), required: true},
		},
	}
	_, err := p.parse(args)
	if err != nil {
		clapFatalf("posargs", err.Error())
	}
}
