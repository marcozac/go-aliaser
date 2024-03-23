package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafe(t *testing.T) {
	sm := NewSafe[string, int](nil)

	assert.False(t, sm.Exist("foo"))
	assert.False(t, sm.PutX("foo", 0))
	assert.False(t, sm.DeleteX("foo"))

	v, ok := sm.SwapX("foo", 1)
	assert.Equal(t, 0, v)
	assert.False(t, ok)

	assert.True(t, sm.PutNX("foo", 1))
	assert.False(t, sm.PutNX("foo", 1))
	assert.True(t, sm.Exist("foo"))
	assert.True(t, sm.PutX("foo", 0))
	assert.True(t, sm.DeleteX("foo"))

	sm.Put("foo", 2)
	v, ok = sm.Swap("foo", 3)
	assert.Equal(t, 2, v)
	assert.True(t, ok)
	v, ok = sm.SwapX("foo", 4)
	assert.Equal(t, 3, v)
	assert.True(t, ok)

	v, ok = sm.Pop("foo")
	assert.Equal(t, 4, v)
	assert.True(t, ok)
	v, ok = sm.Pop("foo")
	assert.Equal(t, 0, v)
	assert.False(t, ok)

	sm.Put("foo", 5)
	sm.Put("bar", 6)
	sm.Put("baz", 7)

	assert.ElementsMatch(t, []string{"foo", "bar", "baz"}, sm.Keys())
	assert.ElementsMatch(t, []int{5, 6, 7}, sm.Values())

	sm.ForEach(func(k string, v int) {
		gv, ok := sm.Get(k)
		assert.True(t, ok)
		assert.Equal(t, v, gv)
	})

	sm.Delete("foo")
	assert.Equal(t, 2, sm.Len())

	sm.Clear()
	assert.Equal(t, 0, sm.Len())
}
