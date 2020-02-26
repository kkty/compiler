package ir

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	for i, c := range []struct {
		functions []*Function
		main      Node
		input     string
		output    []byte
	}{
		{
			[]*Function{},
			&Assignment{
				"a", &ReadInt{},
				&Assignment{
					"b",
					&AddImmediate{"a", 10},
					&WriteByte{"b"},
				},
			},
			"1", []byte{11},
		},
		{
			[]*Function{
				&Function{"f", []string{"a", "b"}, &Add{"a", "b"}},
			},
			&Assignment{
				"x", &ReadInt{},
				&Assignment{
					"y", &ReadInt{},
					&Assignment{
						"z", &Application{"f", []string{"x", "y"}},
						&WriteByte{"z"},
					},
				},
			},
			"2 3", []byte{5},
		},
	} {
		t.Run(fmt.Sprintf("Case%d", i), func(t *testing.T) {
			buf := bytes.Buffer{}
			Execute(c.functions, c.main, nil, &buf, bytes.NewBufferString(c.input))
			assert.Equal(t, c.output, buf.Bytes())
		})
	}
}
