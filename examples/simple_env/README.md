# simple_env (example)

This is a dead simple example that has a string option and a string positional argument,
both of which can be set via environment variables.
To get started, run `go build` and then `./simple_env -h`.

## Usage

```
./simple_env - Print a string with a prefix

usage:
   ./simple_env [options] [input]

options:
   -prefix  <arg>   The value to prepend to the input string [$MY_PREFIX]
   -count  <arg>    Print the output this many extra times [$MY_COUNT]
   -h               show this help message

arguments:
   [input]   The user provided input [$MY_INPUT]
```

## Try It

```shell
./simple_env hello              # 'hello'
MY_INPUT="hello" ./simple_env   # 'hello'

MY_PREFIX="hello, " ./simple_env there   # 'hello, there'
```
