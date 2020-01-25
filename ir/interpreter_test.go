package ir

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	for i, c := range []struct {
		functions     []*Function
		main          Node
		input, output string
	}{
		{
			[]*Function{},
			&ValueBinding{
				"a", &ReadInt{},
				&ValueBinding{
					"b",
					&AddImmediate{"a", 10},
					&PrintInt{"b"},
				},
			},
			"1", "11",
		},
		{
			[]*Function{
				&Function{"f", []string{"a", "b"}, &Sub{"a", "b"}},
			},
			&ValueBinding{
				"x", &ReadInt{},
				&ValueBinding{
					"y", &ReadInt{},
					&ValueBinding{
						"z", &Application{"f", []string{"x", "y"}},
						&PrintInt{"z"},
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
