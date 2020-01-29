package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlphaTransform(t *testing.T) {
	n := &Assignment{
		"x",
		&FunctionAssignment{
			"x", []string{"y"},
			&Variable{"y"},
			&Assignment{
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
		n.Body.(*FunctionAssignment).Name,
		n.Body.(*FunctionAssignment).Next.(*Assignment).Body.(*Application).Function)
	assert.NotEqual(t, n.Name, n.Body.(*FunctionAssignment).Name)

	// y.
	assert.Equal(t,
		n.Body.(*FunctionAssignment).Args[0],
		n.Body.(*FunctionAssignment).Body.(*Variable).Name)
	assert.Equal(t,
		n.Body.(*FunctionAssignment).Next.(*Assignment).Name,
		n.Body.(*FunctionAssignment).Next.(*Assignment).Next.(*Variable).Name)
	assert.NotEqual(t,
		n.Body.(*FunctionAssignment).Args[0],
		n.Body.(*FunctionAssignment).Next.(*Assignment).Name)
}
