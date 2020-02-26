package test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/kkty/compiler/ast"
	"github.com/kkty/compiler/ir"
	"github.com/kkty/compiler/parser"
	"github.com/kkty/compiler/stringset"
	"github.com/stretchr/testify/assert"
)

func TestCompileAndExec(t *testing.T) {
	for _, c := range []struct {
		file     string
		input    string
		expected string
	}{
		{"./ack.ml", "", "253"},
		{"./matmul.ml", "", "5864139154"},
		{"./fib.ml", "", "89"},
		{"./gcd.ml", "", "24"},
	} {
		t.Run(c.file, func(t *testing.T) {
			b, err := ioutil.ReadFile(c.file)
			if err != nil {
				t.Fatal(err)
			}
			program := string(b)
			astNode := parser.Parse(program)
			ast.AlphaTransform(astNode)
			types := ast.GetTypes(astNode)
			main, functions, globals, _ := ir.Generate(astNode, types)
			main, functions = ir.Inline(main, functions, 5, types, false)
			main = ir.RemoveRedundantVariables(main, functions)

			for _, function := range functions {
				assert.Equal(t, 0, len(function.FreeVariables()))
			}
			assert.Equal(t, 0, len(main.FreeVariables(stringset.New())))

			buf := bytes.Buffer{}
			ir.Execute(functions, main, globals, &buf, bytes.NewBufferString(c.input))
			assert.Equal(t, c.expected, buf.String())
		})
	}
}
