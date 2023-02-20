# goclap

A pre-build tool to generate **c**ommand **l**ine **a**rgument **p**arsing
functionality from doc comments in Go. The idea is inspired by the [`clap` Rust
crate](https://github.com/clap-rs/clap), specifically its use of documentation and
proc macros.

## Building

To just build the project as is, run `go build`. If you have
[`task`](https://github.com/go-task/task),
[`goimports`](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) and
[`gofumpt`](https://github.com/mvdan/gofumpt) installed, you can simply run `task` to fmt,
lint and build the project.

## License

This is free and unencumbered software released into the public domain. Please
see the [UNLICENSE](./UNLICENSE) file for more information.
