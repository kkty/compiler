package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/kkty/mincaml-go/alpha"
	"github.com/kkty/mincaml-go/emit"
	"github.com/kkty/mincaml-go/interpreter"
	"github.com/kkty/mincaml-go/knormalize"
	"github.com/kkty/mincaml-go/lifting"
	"github.com/kkty/mincaml-go/parser"
	"github.com/kkty/mincaml-go/typing"
)

func main() {
	interpret := flag.Bool("i", false, "interprets program instead of generating assembly")
	flag.Parse()

	b, err := ioutil.ReadFile(flag.Arg(0))

	if err != nil {
		log.Fatal(err)
	}
	program := string(b)
	astNode := parser.Parse(program)
	alpha.AlphaTransform(astNode)
	mirNode := knormalize.KNormalize(astNode)
	types := typing.GetTypes(mirNode)
	main, functions, _ := lifting.Lift(mirNode, types)
	if *interpret {
		interpreter.Execute(functions, main, os.Stdout, os.Stdin)
	} else {
		emit.Emit(functions, main, types, os.Stdout)
	}
}
