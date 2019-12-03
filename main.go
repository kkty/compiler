package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/kkty/compiler/ast"
	"github.com/kkty/compiler/emit"
	"github.com/kkty/compiler/interpreter"
	"github.com/kkty/compiler/ir"
	"github.com/kkty/compiler/mir"
	"github.com/kkty/compiler/parser"
	"github.com/kkty/compiler/typing"
)

func main() {
	interpret := flag.Bool("i", false, "interprets program instead of generating assembly")
	debug := flag.Bool("debug", false, "enables debugging output")
	inline := flag.Int("inline", 0, "number of inline expansions")
	flag.Parse()

	b, err := ioutil.ReadFile(flag.Arg(0))

	if err != nil {
		log.Fatal(err)
	}
	program := string(b)
	astNode := parser.Parse(program)
	ast.AlphaTransform(astNode)
	mirNode := mir.Generate(astNode)
	types := typing.GetTypes(mirNode)
	main, functions, _ := ir.Generate(mirNode, types)
	main, functions = ir.Inline(main, functions, *inline, types, *debug)
	for i := 0; i < 10; i++ {
		if *debug {
			fmt.Fprintf(os.Stderr, "optimizing (i=%d)\n", i)
		}
		main = ir.RemoveRedundantVariables(main, functions)
		main = ir.Immediate(main, functions)
		main = ir.Reorder(main, functions)
	}
	if *interpret {
		interpreter.Execute(functions, main, os.Stdout, os.Stdin)
	} else {
		emit.Emit(functions, main, types, os.Stdout)
	}
}
