//go:generate goclap -type mycli
package main

// Any changes to this file likely necessitate changes to the project's README.

import (
	"fmt"
	"os"
	"strings"
)

// print a string with the option to make it uppercase
type mycli struct {
	// make the input string all uppercase
	//
	// clap:opt upper,u
	toUpper bool
	// the input string
	//
	// clap:arg_required
	input string
}

func main() {
	c := mycli{}
	c.parse(os.Args)

	s := c.input
	if c.toUpper {
		s = strings.ToUpper(s)
	}

	fmt.Println(s)
}
