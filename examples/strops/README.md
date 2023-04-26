# strops (example)

This is an example that has boolean and string options along with a string positional argument.
To get started, run `go build` and then `./strops -h`.

## Usage

```
./strops - perform different string operations

usage:
   ./strops [options] <input>

options:
   -u, --upper           make the `input` string all uppercase
   -r, --reverse         reverse the final string
       --prefix  <str>   add this prefix to the final string
   -h, --help            show this help message

arguments:
   <input>   the string on which to operate
```

## Try It

```shell
./strops hello      # hello
./strops -u hello   # HELLO

./strops --prefix "hello, " there   # hello, there
```
