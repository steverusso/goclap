package clap

import (
	"flag"
	"fmt"
	"os"
)

type Command struct {
	Path string

	CustomUsage func() string

	flags []Input
	args  []Input
}

type Input struct {
	name       string
	envVarName string
	value      flag.Value
	isRequired bool
	isPresent  bool
}

func (in *Input) Env(name string) *Input {
	in.envVarName = name
	return in
}

func (in *Input) Require() *Input {
	in.isRequired = true
	return in
}

func NewCommand(path string) Command {
	return Command{
		Path: path,
	}
}

func (c *Command) Flag(name string, v flag.Value) *Input {
	c.flags = append(c.flags, Input{
		name:  name,
		value: v,
	})
	return &c.flags[len(c.flags)-1]
}

func (c *Command) Arg(name string, v flag.Value) *Input {
	c.args = append(c.args, Input{
		name:  name,
		value: v,
	})
	return &c.args[len(c.args)-1]
}

func (c *Command) Fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v.\nRun '%s -h' for usage.", err, c.Path)
	os.Exit(2)
}

func (c *Command) Parse(args []string) {
	f := flag.FlagSet{Usage: func() {}}
	for _, opt := range c.flags {
		if opt.envVarName != "" {
			ok, err := parseEnv(opt.value, opt.envVarName)
			if err != nil {
				c.Fatal(err)
			}
			if ok {
				opt.isPresent = true
			}
		}
		f.Var(opt.value, opt.name, "")
	}

	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Println(c.CustomUsage())
			os.Exit(0)
		}
		c.Fatal(err)
	}

	// TODO(steve): check for missing required flags

	rest := f.Args()
	for i, arg := range c.args {
		if len(rest) < i {
			if arg.isRequired {
				c.Fatal(fmt.Errorf("missing required arg '%s'", arg.name))
			}
			return
		}
		if arg.envVarName != "" {
			ok, err := parseEnv(arg.value, arg.envVarName)
			if err != nil {
				c.Fatal(err)
			}
			if ok {
				arg.isPresent = true
			}
		}
		if err := arg.value.Set(rest[0]); err != nil {
			c.Fatal(fmt.Errorf("parsing positional argument '%s': %w", arg.name, err))
		}
	}
}

func parseEnv(v flag.Value, name string) (bool, error) {
	s, ok := os.LookupEnv(name)
	if !ok {
		return false, nil
	}
	if err := v.Set(s); err != nil {
		return true, fmt.Errorf("parsing env var '%s': %w", name, err)
	}
	return true, nil
}
