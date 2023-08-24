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

func (p *CommandParser) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "error: %s.\nRun '%s -h' for usage.\n", msg, p.path)
	os.Exit(2)
}

func (p *CommandParser) Parse(args []string) []string {
	f := flag.FlagSet{Usage: func() {}}
	for _, opt := range p.flags {
		if err := opt.parseEnv(); err != nil {
			p.Fatalf("%v", err)
		}
		f.Var(opt.value, opt.name, "")
	}

	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Println(p.CustomUsage())
			os.Exit(0)
		}
		p.Fatalf("%v", err)
	}

	// TODO(steve): check for missing required flags when supported

	rest := f.Args()

	if len(p.args) > 0 {
		for i, arg := range p.args {
			if len(rest) < i {
				if arg.isRequired {
					p.Fatalf("missing required arg '%s'", arg.name)
				}
				return nil
			}
			if err := arg.parseEnv(); err != nil {
				p.Fatalf("%v", err)
			}
			if err := arg.value.Set(rest[i]); err != nil {
				p.Fatalf("parsing positional argument '%s': %v", arg.name, err)
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
