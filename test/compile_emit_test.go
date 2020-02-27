package test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/kkty/compiler/ast"
	"github.com/kkty/compiler/emit"
	"github.com/kkty/compiler/ir"
	"github.com/kkty/compiler/parser"
	"github.com/kkty/compiler/stringset"
	"github.com/stretchr/testify/assert"
)

func TestCompileAndEmit(t *testing.T) {
	for _, file := range []string{
		"./ack.ml",
		"./fib.ml",
		"./gcd.ml",
		"./mandelbrot.ml",
		"./matmul.ml",
		"./min-rt.ml",
		"./array.ml",
	} {
		t.Run(file, func(t *testing.T) {
			b, err := ioutil.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			program := string(b)
			astNode := parser.Parse(program)
			ast.AlphaTransform(astNode)
			types := ast.GetTypes(astNode)
			main, functions, globals, _ := ir.Generate(astNode, types)
			main, _ = ir.Inline(main, functions, 5, types, false)
			for i := 0; i < 5; i++ {
				main = ir.RemoveRedundantAssignments(main, functions)
				main = ir.Immediate(main, functions)
				main = ir.Reorder(main, functions)
			}

			globalNames := stringset.New()
			for name := range globals {
				globalNames.Add(name)
			}

			for _, function := range append(functions, &ir.Function{
				Body: main,
			}) {
				for freeVariable := range function.FreeVariables() {
					if !globalNames.Has(freeVariable) {
						assert.Fail(t, freeVariable)
					}
				}
			}

			emit.AllocateRegisters(main, functions, globals, types)
			emit.Emit(functions, main, globals, types, &bytes.Buffer{})
		})
	}
}
