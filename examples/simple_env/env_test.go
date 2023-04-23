package simple_env

// Any changes to this file likely necessitate changes to the project's README.

import (
	"fmt"
)

func Example_configFromEnv() {
	c := configFromEnv()

	fmt.Println(c.input)
	// Output: abc123
}
