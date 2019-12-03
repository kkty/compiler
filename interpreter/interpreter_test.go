package interpreter

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/kkty/compiler/ir"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	for i, c := range []struct {
		functions     []*ir.Function
		main          ir.Node
		input, output string
	}{
		{
			[]*ir.Function{},
			&ir.ValueBinding{
				"a", &ir.ReadInt{},
				&ir.ValueBinding{
					"b",
					&ir.AddImmediate{"a", 10},
					&ir.PrintInt{"b"},
				},
			},
			"1", "11",
		},
		{
			[]*ir.Function{
				&ir.Function{"f", []string{"a", "b"}, &ir.Sub{"a", "b"}},
			},
			&ir.ValueBinding{
				"x", &ir.ReadInt{},
				&ir.ValueBinding{
					"y", &ir.ReadInt{},
					&ir.ValueBinding{
						"z", &ir.Application{"f", []string{"x", "y"}},
						&ir.PrintInt{"z"},
					},
				},
			},
			"2 3", "-1",
		},
	} {
		t.Run(fmt.Sprintf("Case%d", i), func(t *testing.T) {
			buf := bytes.Buffer{}
			Execute(c.functions, c.main, &buf, bytes.NewBufferString(c.input))
			assert.Equal(t, c.output, buf.String())
		})
	}
}
