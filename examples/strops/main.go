//go:generate goclap -type strops
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
	// reverse the final string
	//
	// clap:opt reverse,r
	reverse bool
	// add this prefix to the final string
	//
	// clap:opt prefix
	// clap:opt_arg_name str
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

	if c.reverse {
		n := len(b)
		for i := 0; i < n/2; i++ {
			b[i], b[n-1-i] = b[n-1-i], b[i]
		}
	}

	fmt.Println(string(b))
}
