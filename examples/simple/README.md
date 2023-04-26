# simple (example)

This is a dead simple example that has a boolean option and a string positional argument.
To get started, run `go build` and then `./simple -h`.

## Usage

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

## Try It

```shell
./simple hello      # hello
./simple -u hello   # HELLO
```
