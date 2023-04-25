package main

// Any changes to this file likely necessitate changes to the project's README.
//go:generate goclap -type mycli

import (
	"fmt"
	"os"
)

// Print a string with the option to make it uppercase.
type mycli struct {
	// The user provided username.
	//
	// clap:opt uname,u
	// clap:env MY_UNAME
	username string
	// The user provided password.
	//
	// clap:opt passwd,p
	// clap:env MY_PASSWD
	password string
}

func main() {
	c := mycli{}
	c.parse(os.Args)
	fmt.Printf("username: %q\n", c.username)
	fmt.Printf("password: %q\n", c.password)
}
