package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	assert.Panics(t, func() {
		Must(assert.AnError)
	})
	assert.Panics(t, func() {
		MustV(0, assert.AnError)
	})
}
