package test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/kkty/mincaml-go/ast"
	"github.com/kkty/mincaml-go/emit"
	"github.com/kkty/mincaml-go/ir"
	"github.com/kkty/mincaml-go/mir"
	"github.com/kkty/mincaml-go/parser"
	"github.com/kkty/mincaml-go/typing"
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
	} {
		t.Run(file, func(t *testing.T) {
			b, err := ioutil.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			program := string(b)
			astNode := parser.Parse(program)
			ast.AlphaTransform(astNode)
			mirNode := mir.Generate(astNode)
			types := typing.GetTypes(mirNode)
			main, functions, _ := ir.Generate(mirNode, types)
			main, _ = ir.Inline(main, functions, 5, types, false)
			for i := 0; i < 10; i++ {
				main = ir.RemoveRedundantVariables(main, functions)
				main = ir.Immediate(main, functions)
				main = ir.Reorder(main, functions)
			}

			for _, function := range functions {
				assert.Equal(t, 0, len(function.FreeVariables()))
			}
			assert.Equal(t, 0, len(main.FreeVariables(map[string]struct{}{})))

			emit.Emit(functions, main, types, &bytes.Buffer{})
		})
	}
}
