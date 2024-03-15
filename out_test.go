//go:build testout

package aliaser

import (
	"testing"

	"github.com/marcozac/go-aliaser/internal/testdata"
	"github.com/marcozac/go-aliaser/internal/testout"
	"github.com/stretchr/testify/assert"
)

func TestOut(t *testing.T) {
	assert.Equal(t, testdata.A, testout.A)
	assert.Equal(t, testdata.B, testout.B)
	assert.NotPanics(t, testout.C)
	assert.Equal(t, testdata.D(""), testout.D(""))
	assert.Equal(t, testdata.E(0), testout.E(0))
	assert.Equal(t, testdata.F(0), testout.F(0))
	assert.Equal(t, testdata.G(""), testout.G(""))
	assert.Equal(t, testdata.H, testout.H)
	assert.Equal(t, testdata.I, testout.I)
}
