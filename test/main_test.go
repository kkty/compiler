package test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/kkty/mincaml-go/ast"
	"github.com/kkty/mincaml-go/interpreter"
	"github.com/kkty/mincaml-go/ir"
	"github.com/kkty/mincaml-go/knormalize"
	"github.com/kkty/mincaml-go/lifting"
	"github.com/kkty/mincaml-go/parser"
	"github.com/kkty/mincaml-go/typing"
	"github.com/stretchr/testify/assert"
)

func TestCompileAndExecution(t *testing.T) {
	for _, c := range []struct {
		file     string
		input    string
		expected string
	}{
		{"./ack.ml", "", "253"},
		{"./matmul.ml", "", "5864139154"},
		{"./fib.ml", "", "89"},
	} {
		b, err := ioutil.ReadFile(c.file)
		if err != nil {
			t.Fatal(err)
		}
		program := string(b)
		astNode := parser.Parse(program)
		ast.AlphaTransform(astNode)
		mirNode := knormalize.KNormalize(astNode)
		types := typing.GetTypes(mirNode)
		main, functions, _ := lifting.Lift(mirNode, types)
		main, functions = ir.Inline(main, functions, 10, types)
		buf := bytes.Buffer{}
		interpreter.Execute(functions, main, &buf, bytes.NewBufferString(c.input))
		assert.Equal(t, c.expected, buf.String())
	}
}

func TestCompile(t *testing.T) {
	b, err := ioutil.ReadFile("./min-rt.ml")
	if err != nil {
		t.Fatal(err)
	}
	program := string(b)
	astNode := parser.Parse(program)
	ast.AlphaTransform(astNode)
	mirNode := knormalize.KNormalize(astNode)
	types := typing.GetTypes(mirNode)
	main, functions, _ := lifting.Lift(mirNode, types)
	ir.Inline(main, functions, 10, types)
}
