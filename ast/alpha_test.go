package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlphaTransform(t *testing.T) {
	n := &ValueBinding{
		"x",
		&FunctionBinding{
			"x", []string{"y"},
			&Variable{"y"},
			&ValueBinding{
				"y", &Application{"x", []Node{&Int{0}}},
				&Variable{"y"},
			},
		},
		&Add{&Variable{"x"}, &Variable{"x"}},
	}

	AlphaTransform(n)

	// x.
	assert.Equal(t, n.Name, n.Next.(*Add).Left.(*Variable).Name)
	assert.Equal(t,
		n.Body.(*FunctionBinding).Name,
		n.Body.(*FunctionBinding).Next.(*ValueBinding).Body.(*Application).Function)
	assert.NotEqual(t, n.Name, n.Body.(*FunctionBinding).Name)

	// y.
	assert.Equal(t,
		n.Body.(*FunctionBinding).Args[0],
		n.Body.(*FunctionBinding).Body.(*Variable).Name)
	assert.Equal(t,
		n.Body.(*FunctionBinding).Next.(*ValueBinding).Name,
		n.Body.(*FunctionBinding).Next.(*ValueBinding).Next.(*Variable).Name)
	assert.NotEqual(t,
		n.Body.(*FunctionBinding).Args[0],
		n.Body.(*FunctionBinding).Next.(*ValueBinding).Name)
}
