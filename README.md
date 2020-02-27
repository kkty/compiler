Toy compiler for subset of OCaml, which generates assembly code that can be executed with https://github.com/kkty/simulator.

## Features

- Support for subset of OCaml
  - `test/*.ml` files will give you some ideas of available syntax and built-in functions.
- Constant folding
  - `let rec double i = i + i in let six = double 3 in ...` will be converted to `... let six = 6 in ...`.
- Inline expansion
  - `let rec double i = i + i in let y = double x in ...` will be converted to `... let y = x + x in ...`.
- Reordering of variable assignments
  - `let i = ... in if ... then (i is used here) else (i is not used here)` will be converted to `if ... then let i = ... in (i is used here) else (i is not used here)`.
- Removal of unused variables
  - `let i = (code without side effects) in (code which does not use i)` will be converted to `(code which does not use i)`.
- Register allocation with graph coloring
- Visualization of IR (see below)
- Interpreter of IR (see below)

## Requirements

- Go >= 1.13

## Install

```console
$ go get -u github.com/kkty/compiler
$ go get -u github.com/kkty/simulator # to execute assembly
```

## Usage

```console
$ compiler --help
Usage of compiler:
  -debug
        enables debugging output
  -graph
        outputs graph in dot format
  -i    interprets program instead of generating assembly
  -inline int
        number of inline expansions
  -iter int
        number of iterations for optimization
```

### Examples

Calculates the 10th fibonacci number.

```console
$ compiler <this_repository>/test/fib.ml > program.s
$ simulator program.s
89
```

---

Generates an image with ray tracing (with some optimization flags).

```console
$ compiler <this_repository>/test/min-rt.ml -iter 5 -inline 50 > program.s
$ simulator program.s > out.ppm << EOF
0 0 0 0 30
1 0 0
255
0 1 2 0 40 10 40 0 -40 0 1 0.2 64 255 255 0
4 3 1 0 30 30 30 0 0 0 1 1 255 255 255 255
-1
0 -1
1 -1
-1
99 0 1 -1
-1
EOF
$ <open out.ppm>
```

- The optimizer (constant folding, etc.) is run for 5 times.
- Inline expansion is applied to all the non-recursive functions and 50 recursive functions in the program. 
- With the `-debug` option, you can see which functions are inlined and how much the program size has got reduced.

---

Visualizes the fibonacci program as graphs.

```console
$ compiler <this_repository>/test/fib.ml -graph
$ ls graphs
fib_17.dot main.dot
$ dot -Tpdf main.dot > graph.pdf
$ <open graph.pdf>
```

---

Calculates the greatest common divisor of 72 and 120 with the built-in interpreter.

```console
$ compiler -i <this_repository>/gcd.ml
24
```

- A program is converted to IR and then is interpreted.

## Credits

Grammar files and example programs are by https://github.com/esumii/min-caml with some modifications.
