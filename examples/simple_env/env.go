//go:generate goclap -type mycli
package simple_env

import "os"

// Print a string
type mycli struct {
	// The input string.
	//
	// clap:env INPUT_STRING
	input string
}

func configFromEnv() mycli {
	originalValue := os.Getenv("INPUT_STRING")
	defer func() {
		os.Setenv("INPUT_STRING", originalValue)
	}()
	os.Setenv("INPUT_STRING", "abc123")

	c := mycli{}
	c.parse(nil)

	return c
}
