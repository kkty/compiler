package parser

import (
	"testing"

	"github.com/kkty/mincaml-go/ast"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	for _, c := range []struct {
		program  string
		expected ast.Node
	}{
		{
			"if x = y then 1 else 0",
			&ast.If{
				&ast.Equal{&ast.Variable{"x"}, &ast.Variable{"y"}},
				&ast.Int{1},
				&ast.Int{0},
			},
		},
		{
			"let rec f x y = x + y in f 1 2",
			&ast.FunctionBinding{
				"f", []string{"x", "y"},
				&ast.Add{&ast.Variable{"x"}, &ast.Variable{"y"}},
				&ast.Application{"f", []ast.Node{&ast.Int{1}, &ast.Int{2}}},
			},
		},
	} {
		assert.Equal(t, c.expected, Parse(c.program))
	}
}
