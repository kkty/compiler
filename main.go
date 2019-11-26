package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/kkty/mincaml-go/ast"
	"github.com/kkty/mincaml-go/emit"
	"github.com/kkty/mincaml-go/interpreter"
	"github.com/kkty/mincaml-go/ir"
	"github.com/kkty/mincaml-go/mir"
	"github.com/kkty/mincaml-go/parser"
	"github.com/kkty/mincaml-go/typing"
)

func main() {
	interpret := flag.Bool("i", false, "interprets program instead of generating assembly")
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
	main, functions = ir.Inline(main, functions, *inline, types)
	main = ir.RemoveRedundantVariables(main, functions)
	if *interpret {
		interpreter.Execute(functions, main, os.Stdout, os.Stdin)
	} else {
		emit.Emit(functions, main, types, os.Stdout)
	}
}
