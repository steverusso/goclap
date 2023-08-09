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

func (p *CommandParser) Flag(name string, v flag.Value) *Input {
	p.flags = append(p.flags, Input{
		name:  name,
		value: v,
	})
	return &p.flags[len(p.flags)-1]
}

func (p *CommandParser) Arg(name string, v flag.Value) *Input {
	p.args = append(p.args, Input{
		name:  name,
		value: v,
	})
	return &p.args[len(p.args)-1]
}

func (p *CommandParser) Fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v.\nRun '%s -h' for usage.", err, p.path)
	os.Exit(2)
}

func (p *CommandParser) Parse(args []string) []string {
	f := flag.FlagSet{Usage: func() {}}
	for _, opt := range p.flags {
		if err := opt.parseEnv(); err != nil {
			p.Fatal(err)
		}
		f.Var(opt.value, opt.name, "")
	}

	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Println(p.CustomUsage())
			os.Exit(0)
		}
		p.Fatal(err)
	}

	// TODO(steve): check for missing required flags when supported

	rest := f.Args()

	if len(p.args) > 0 {
		for i, arg := range p.args {
			if len(rest) < i {
				if arg.isRequired {
					p.Fatal(fmt.Errorf("missing required arg '%s'", arg.name))
				}
				return nil
			}
			if err := arg.parseEnv(); err != nil {
				p.Fatal(err)
			}
			if err := arg.value.Set(rest[0]); err != nil {
				p.Fatal(fmt.Errorf("parsing positional argument '%s': %w", arg.name, err))
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
