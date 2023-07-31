# strops (example)

This is an example that has boolean and string options along with a string positional argument.
To get started, run `go build` and then `./strops -h`.

## Usage

```
./strops - Perform different string operations

usage:
   ./strops [options] <input>

options:
   -upper           Make the `input` string all uppercase
   -reverse         Reverse the final string
   -repeat  <n>     Repeat the string this many times
   -prefix  <str>   Add this prefix to the final string
   -h               show this help message

arguments:
   <input>   The string on which to operate
```

## Try It

```shell
./strops hello      # hello
./strops -u hello   # HELLO

./strops --prefix "hello, " there   # hello, there
```
