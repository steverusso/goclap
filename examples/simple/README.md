# simple (example)

This is a dead simple example that has a boolean option and a string positional argument.
To get started, run `go build` and then `./simple -h`.

## Usage

```
./simple - Print a string with the option to make it uppercase

usage:
   ./simple [options] <input>

options:
   -upper   Make the input string all uppercase
   -h       Show this help message

arguments:
   <input>   The input string
```

## Try It

```shell
./simple hello          # hello
./simple -upper hello   # HELLO
```
