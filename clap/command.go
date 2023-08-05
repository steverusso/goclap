package clap

import (
	"flag"
	"fmt"
	"os"
)

type Command struct {
	Path  string
	Flags flag.FlagSet
}

func NewCommand(path string) Command {
	return Command{
		Path:  path,
		Flags: flag.FlagSet{Usage: func() {}},
	}
}

func (c *Command) Flag(name string, v flag.Value) {
	c.Flags.Var(v, name, "")
}

func (c *Command) Fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v.\nRun '%s -h' for usage.", err, c.Path)
	os.Exit(2)
}

func (c *Command) Parse(args []string, usgFn func() string) {
	if err := c.Flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Println(usgFn())
			os.Exit(0)
		}
		c.Fatal(err)
	}
}
