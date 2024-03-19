package sequence

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type sequencedSlice []string

func (s sequencedSlice) Len() int {
	return len(s)
}

func (s sequencedSlice) At(i int) string {
	return s[i]
}

func TestSequence(t *testing.T) {
	ss := sequencedSlice{"foo", "bar", "baz"}
	seq := New(ss.Len, ss.At)
	t.Run("Len", func(t *testing.T) {
		assert.Equal(t, 3, seq.Len())
	})
	t.Run("At", func(t *testing.T) {
		assert.Equal(t, "foo", seq.At(0))
		assert.Equal(t, "bar", seq.At(1))
		assert.Equal(t, "baz", seq.At(2))
	})
	t.Run("Slice", func(t *testing.T) {
		assert.Equal(t, []string{"foo", "bar", "baz"}, seq.Slice())
	})
	t.Run("SliceFunc", func(t *testing.T) {
		assert.Equal(t, []string{"FOO", "BAR", "BAZ"}, seq.SliceFunc(func(s string) string {
			return strings.ToUpper(s)
		}))
	})
	t.Run("SliceFuncFilter", func(t *testing.T) {
		assert.Equal(t, []string{"foo", "baz"}, seq.SliceFuncFilter(func(s string) (string, bool) {
			return s, s != "bar"
		}))
	})
}
