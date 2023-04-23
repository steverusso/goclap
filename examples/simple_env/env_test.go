package simple_env

// Any changes to this file likely necessitate changes to the project's README.

import (
	"fmt"
	"os"
)

func ExampleEnvCLI() {
	originalValue := os.Getenv("INPUT_STRING")
	defer func() {
		os.Setenv("INPUT_STRING", originalValue)
	}()
	os.Setenv("INPUT_STRING", "abc123")

	c := mycli{}
	c.parse(nil)

	fmt.Println(c.input)
	// Output: abc123
}
