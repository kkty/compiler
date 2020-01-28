package stringset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	s1 := NewFromSlice([]string{"foo"})
	s2 := NewFromSlice([]string{"foo", "bar"})

	restore := s1.Join(s2)

	assert.Equal(t, true, s1.Has("foo"))
	assert.Equal(t, true, s1.Has("bar"))

	restore(s1)

	assert.Equal(t, true, s1.Has("foo"))
	assert.Equal(t, false, s1.Has("bar"))
}
