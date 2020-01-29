Toy compiler for subset of OCaml, which generates MIPS-like assembly that can be executed by https://github.com/kkty/simulator.

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

Generates an image with ray tracing with some optimization.

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

- The optimizer (immediate-value optimization, etc.) is run for 5 times.
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

## Credits

Grammar files and example programs are by https://github.com/esumii/min-caml with some modifications.
