package clap

import (
	"flag"
	"fmt"
	"os"
)

type CommandParser struct {
	CustomUsage func() string

	path  string
	flags []Input
	args  []Input
}

func NewCommandParser(path string) CommandParser {
	return CommandParser{
		path: path,
	}
}

func (c *CommandParser) Flag(name string, v flag.Value) *Input {
	c.flags = append(c.flags, Input{
		name:  name,
		value: v,
	})
	return &c.flags[len(c.flags)-1]
}

func (c *CommandParser) Arg(name string, v flag.Value) *Input {
	c.args = append(c.args, Input{
		name:  name,
		value: v,
	})
	return &c.args[len(c.args)-1]
}

func (c *CommandParser) Fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v.\nRun '%s -h' for usage.", err, c.path)
	os.Exit(2)
}

func (c *CommandParser) Parse(args []string) []string {
	f := flag.FlagSet{Usage: func() {}}
	for _, opt := range c.flags {
		if err := opt.parseEnv(); err != nil {
			c.Fatal(err)
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

	// TODO(steve): check for missing required flags when supported

	rest := f.Args()

	if len(c.args) > 0 {
		for i, arg := range c.args {
			if len(rest) < i {
				if arg.isRequired {
					c.Fatal(fmt.Errorf("missing required arg '%s'", arg.name))
				}
				return nil
			}
			if err := arg.parseEnv(); err != nil {
				c.Fatal(err)
			}
			if err := arg.value.Set(rest[0]); err != nil {
				c.Fatal(fmt.Errorf("parsing positional argument '%s': %w", arg.name, err))
			}
		}
		return nil
	}

	return f.Args()
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

func (in *Input) parseEnv() error {
	if in.envVarName == "" {
		return nil
	}
	s, ok := os.LookupEnv(in.envVarName)
	if !ok {
		return nil
	}
	in.isPresent = true
	if err := in.value.Set(s); err != nil {
		return fmt.Errorf("parsing env var '%s': %w", in.envVarName, err)
	}
	return nil
}
