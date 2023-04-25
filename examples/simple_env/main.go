package main

// Any changes to this file likely necessitate changes to the project's README.
//go:generate goclap -type mycli

import (
	"fmt"
	"os"
)

// Print a string with a prefix.
type mycli struct {
	// The value to prepend to the input string.
	//
	// clap:opt prefix,p
	// clap:env MY_PREFIX
	prefix string
	// The user provided input.
	//
	// clap:env MY_INPUT
	input string
}

func main() {
	c := mycli{}
	c.parse(os.Args)
	fmt.Printf("'%s%s'\n", c.prefix, c.input)
}
