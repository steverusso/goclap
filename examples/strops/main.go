//go:generate goclap strops
package main

// Any changes to this file likely necessitate changes to the project's README.

import (
	"bytes"
	"fmt"
	"os"
)

// perform different string operations
type strops struct {
	// make the input string all uppercase
	//
	// clap:opt upper,u
	toUpper bool
	// add this prefix to the final string
	//
	// clap:opt prefix,p
	prefix string
	// the string on which to operate
	//
	// clap:arg_required
	input string
}

func main() {
	c := strops{}
	c.parse(os.Args)

	b := []byte(c.input)
	if c.toUpper {
		b = bytes.ToUpper(b)
	}

	if c.prefix != "" {
		b = append([]byte(c.prefix), b...)
	}

	fmt.Println(string(b))
}
