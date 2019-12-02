package stringmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	m := Map{"x": "foo"}
	restore := m.Join(Map{"x": "bar", "y": "baz"})
	assert.Equal(t, 2, len(m))
	assert.Equal(t, "bar", m["x"])
	assert.Equal(t, "baz", m["y"])
	restore(m)
	assert.Equal(t, 1, len(m))
	assert.Equal(t, "foo", m["x"])
}
