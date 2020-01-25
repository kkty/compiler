package main

import (
	"flag"
	"fmt"
	"github.com/kkty/compiler/ast"
	"github.com/kkty/compiler/emit"
	"github.com/kkty/compiler/interpreter"
	"github.com/kkty/compiler/ir"
	"github.com/kkty/compiler/parser"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	interpret := flag.Bool("i", false, "interprets program instead of generating assembly")
	debug := flag.Bool("debug", false, "enables debugging output")
	graph := flag.Bool("graph", false, "outputs graph in dot format")
	inline := flag.Int("inline", 0, "number of inline expansions")
	iter := flag.Int("iter", 0, "number of iterations for optimization")

	flag.Parse()

	b, err := ioutil.ReadFile(flag.Arg(0))

	if err != nil {
		log.Fatal(err)
	}

	root := parser.Parse(string(b))
	ast.AlphaTransform(root)
	types := ast.GetTypes(root)

	main, functions, _ := ir.Generate(root, types)
	main, functions = ir.Inline(main, functions, *inline, types, *debug)

	for i := 0; i < *iter; i++ {
		if *debug {
			fmt.Fprintf(os.Stderr, "optimizing (i=%d)\n", i)
		}
		main = ir.RemoveRedundantVariables(main, functions)
		main = ir.Immediate(main, functions)
		main = ir.Reorder(main, functions)
	}

	if *graph {
		ir.GenerateGraph(main, functions)
	} else if *interpret {
		interpreter.Execute(functions, main, os.Stdout, os.Stdin)
	} else {
		emit.Emit(functions, main, types, os.Stdout)
	}
}
