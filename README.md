# goclap

```
go install github.com/steverusso/goclap@latest
```

A pre-build tool to generate **c**ommand **l**ine **a**rgument **p**arsing code
from Go comments. The idea is inspired by the [`clap` Rust
crate](https://github.com/clap-rs/clap), specifically its use of documentation
and proc macros.

## Example

The following is taken from [`examples/simple/main.go`](./examples/simple/main.go).

```go
//go:generate goclap -type mycli

...

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
```

By running `go generate` (assuming `goclap` is installed), the `mycli` struct,
its fields, and their comments will be used to generate code for parsing
command line arguments into a `mycli`. That code will be placed in a file named
`clap.go` (see [the `simple` example's one](./examples/simple/clap.go)). The
program can then be built with `go build`.

Running `./simple -u hello` will output "HELLO", and running `./simple -h` will
output the following help message:

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
[`task`](https://github.com/go-task/task),
[`goimports`](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) and
[`gofumpt`](https://github.com/mvdan/gofumpt) installed, you can simply run `task` to fmt,
lint and build the project.

## License

This is free and unencumbered software released into the public domain. Please
see the [UNLICENSE](./UNLICENSE) file for more information.
