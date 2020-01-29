package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlphaTransform(t *testing.T) {
	n := &Assignment{
		"x",
		&FunctionBinding{
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
		n.Body.(*FunctionBinding).Name,
		n.Body.(*FunctionBinding).Next.(*Assignment).Body.(*Application).Function)
	assert.NotEqual(t, n.Name, n.Body.(*FunctionBinding).Name)

	// y.
	assert.Equal(t,
		n.Body.(*FunctionBinding).Args[0],
		n.Body.(*FunctionBinding).Body.(*Variable).Name)
	assert.Equal(t,
		n.Body.(*FunctionBinding).Next.(*Assignment).Name,
		n.Body.(*FunctionBinding).Next.(*Assignment).Next.(*Variable).Name)
	assert.NotEqual(t,
		n.Body.(*FunctionBinding).Args[0],
		n.Body.(*FunctionBinding).Next.(*Assignment).Name)
}
