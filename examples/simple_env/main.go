package main

//go:generate goclap -type mycli

// Any changes to this file likely necessitate changes to the project's README.

import (
	"fmt"
	"os"
)

// Print a string with a prefix.
type mycli struct {
	// The value to prepend to the input string.
	//
	// clap:opt prefix
	// clap:env MY_PREFIX
	prefix string
	// Print the output this many extra times.
	//
	// clap:opt count
	// clap:env MY_COUNT
	count uint
	// The user provided input.
	//
	// clap:env MY_INPUT
	input string
}

func main() {
	c := mycli{}
	c.Parse(os.Args[1:])

	for i := uint(0); i < c.count+1; i++ {
		fmt.Printf("'%s%s'\n", c.prefix, c.input)
	}
}
