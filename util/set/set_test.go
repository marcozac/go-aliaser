// Copyright (c) 2023 ISK SRL. All rights reserved.

package set

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	s := New[string, struct{}]()
	_, ok := s.Get("foo")
	assert.False(t, ok, "expected key not to be found")
	_, ok = s.Pop("foo")
	assert.False(t, ok, "expected key not to be found")
	assert.False(t, s.PutX("foo", struct{}{}), "expected key not to be found")

	assert.True(t, s.PutNX("foo", struct{}{}), "expected key not to be found")
	assert.False(t, s.PutNX("foo", struct{}{}), "expected key to be found")
	assert.True(t, s.PutX("foo", struct{}{}), "expected key to be found")

	_, ok = s.Get("foo")
	assert.True(t, ok, "expected key to be found")
	_, ok = s.Pop("foo")
	assert.True(t, ok, "expected key to be found")
	assert.False(t, s.Delete("foo"), "expected key not to be found")

	s.Put("foo", struct{}{})
	assert.True(t, s.Delete("foo"), "expected key to be found")
}
