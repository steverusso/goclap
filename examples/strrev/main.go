//go:generate goclap mycli
package main

// Any changes to this file likely necessitate changes to the project's README.

import (
	"bytes"
	"fmt"
	"os"
)

// reverse a string and maybe make it uppercase
type mycli struct {
	// make the input string all uppercase
	//
	// clap:opt upper,u
	toUpper bool
	// the string to reverse
	//
	// clap:arg_required
	input string
}

func main() {
	c := mycli{}
	c.parse(os.Args)

	b := []byte(c.input)
	if c.toUpper {
		b = bytes.ToUpper(b)
	}

	n := len(b)
	for i := 0; i < n/2; i++ {
		b[i], b[n-1-i] = b[n-1-i], b[i]
	}
	fmt.Println(string(b))
}
