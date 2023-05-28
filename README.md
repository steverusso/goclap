# goclap

[![Go Reference](https://pkg.go.dev/badge/github.com/steverusso/goclap.svg)](https://pkg.go.dev/github.com/steverusso/goclap)
[![GitHub CI](https://github.com/steverusso/goclap/actions/workflows/ci.yaml/badge.svg)](https://github.com/steverusso/goclap/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/steverusso/goclap)](https://goreportcard.com/report/github.com/steverusso/goclap)

```
go install github.com/steverusso/goclap@latest
```

A pre-build tool to generate **c**ommand **l**ine **a**rgument **p**arsing code from Go
comments. The idea is inspired by the [`clap` Rust
crate](https://github.com/clap-rs/clap), specifically its use of documentation and proc
macros.

## Example

The following is taken from [`examples/simple/main.go`](./examples/simple/main.go).

```go
//go:generate goclap -type mycli

...

// Print a string with the option to make it uppercase.
type mycli struct {
	// Make the input string all uppercase.
	//
	// clap:opt upper,u
	toUpper bool
	// The input string.
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
```

By running `go generate` (assuming `goclap` is installed), the `mycli` struct, its fields,
and their comments will be used to generate code for parsing command line arguments into a
`mycli`. That code will be placed in a file named `clap.gen.go` (see [the `simple`
example's one](./examples/simple/clap.gen.go)). The program can then be built with `go
build`.

Running `./simple -u hello` will output "HELLO", and running `./simple -h` will output the
following help message:

```
./simple - print a string with the option to make it uppercase

usage:
   ./simple [options] <input>

options:
   -u, --upper   make the input string all uppercase
   -h, --help    show this help message

arguments:
   <input>   the input string
```

## Building

To just build the project as is, run `go build`. If you have
[`task`](https://github.com/go-task/task) installed, as well as the necessary additional
tools ([`goimports`](https://pkg.go.dev/golang.org/x/tools/cmd/goimports),
[`gofumpt`](https://github.com/mvdan/gofumpt), [`staticcheck`](https://staticcheck.io/)),
then you can run `task` to fmt, lint and build the project.

## Projects Using Goclap

* [lockbook-x/lbcli](https://github.com/steverusso/lockbook-x/tree/master/lbcli)

## License

This is free and unencumbered software released into the public domain. Please
see the [UNLICENSE](./UNLICENSE) file for more information.
