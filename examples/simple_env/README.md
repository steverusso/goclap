# simple_env (example)

This is a dead simple example that has a string option and a string positional argument,
both of which can be set via environment variables.
To get started, run `go build` and then `./simple_env -h`.

## Usage

```
./simple_env - print a string with a prefix

usage:
   ./simple_env [options] [input]

options:
   -p, --prefix  <arg>   the value to prepend to the input string [$MY_PREFIX]
   -h, --help            show this help message

arguments:
   [input]   the user provided input [$MY_INPUT]
```

## Try It

```shell
./simple_env hello              # 'hello'
MY_INPUT="hello" ./simple_env   # 'hello'

MY_PREFIX="hello, " ./simple_env there   # 'hello, there'
```
