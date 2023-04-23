//go:generate goclap -type mycli
package simple_env

// Print a string
type mycli struct {
	// The input string.
	//
	// clap:env INPUT_STRING
	input string
}
