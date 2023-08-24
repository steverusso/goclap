package main

//go:generate goclap -type mycli

import (
	"fmt"
	"os"
)

// Print a few positional args.
type mycli struct {
	// A float32 positional arg.
	//
	// clap:arg_required
	f32 float32
	// A string positional arg.
	//
	// clap:arg_name text
	// clap:arg_required
	str string
	// A uint16 positional arg.
	//
	// clap:arg_required
	u16 uint16
}

func main() {
	c := mycli{}
	c.Parse(os.Args[1:])

	fmt.Printf("f32: %f\n", c.f32)
	fmt.Printf("str: %s\n", c.str)
	fmt.Printf("u16: %d\n", c.u16)
}
