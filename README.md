# goclap

```
go install github.com/steverusso/goclap@latest
```

A pre-build tool to generate **c**ommand **l**ine **a**rgument **p**arsing code
from Go comments. The idea is inspired by the [`clap` Rust
crate](https://github.com/clap-rs/clap), specifically its use of documentation
and proc macros.

## Example

The following is taken from [`examples/strrev/main.go`](./examples/strrev/main.go).

```go
//go:generate goclap mycli

...

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
```

By running `go generate` (assuming `goclap` is installed), the `mycli` struct,
its fields, and their comments will be used to generate code for parsing
command line arguments into a `mycli`. That code will be placed in a file named
`clap.go` (see [the `strrev` example one](./examples/strrev/clap.go)). The
program can then be built with `go build`.

Running `./strrev -u hello` will output "OLLEH", and running `./strrev -h` will
output the following help message:

```
./strrev - reverse a string and maybe make it uppercase

usage:
   ./strrev [options] <input>

options:
   --upper, -u   make the input string all uppercase
   --help, -h    show this help message

arguments:
   <input>   the string to reverse
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
