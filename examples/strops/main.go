package main

//go:generate goclap -type strops

// Any changes to this file likely necessitate changes to the project's README.

import (
	"bytes"
	"fmt"
	"os"
)

// Perform different string operations.
type strops struct {
	// Make the `input` string all uppercase.
	//
	// clap:opt upper
	toUpper bool
	// Reverse the final string.
	//
	// clap:opt reverse
	reverse bool
	// Repeat the string this many times.
	//
	// clap:opt repeat
	// clap:opt_arg_name n
	repeat int
	// Add this prefix to the final string.
	//
	// clap:opt prefix
	// clap:opt_arg_name str
	prefix string
	// The string on which to operate.
	//
	// clap:arg_required
	input string
}

func main() {
	c := strops{}
	c.Parse(os.Args[1:])

	b := []byte(c.input)
	if c.prefix != "" {
		b = append([]byte(c.prefix), b...)
	}

	if c.reverse {
		n := len(b)
		for i := 0; i < n/2; i++ {
			b[i], b[n-1-i] = b[n-1-i], b[i]
		}
	}
	if c.toUpper {
		b = bytes.ToUpper(b)
	}
	if c.repeat > 0 {
		b = bytes.Repeat(b, c.repeat)
	}

	fmt.Println(string(b))
}
