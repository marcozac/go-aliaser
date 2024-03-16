//go:build testout

package out

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/marcozac/go-aliaser/internal/testing/pkg"
)

func TestOut(t *testing.T) {
	assert.Equal(t, pkg.A, A)
	assert.Equal(t, pkg.B, B)
	assert.NotPanics(t, C)
	assert.Equal(t, pkg.D(""), D(""))
	assert.Equal(t, pkg.E(0), E(0))
	assert.Equal(t, pkg.F(0), F(0))
	assert.Equal(t, pkg.G(""), G(""))
	assert.Equal(t, pkg.H, H)
	assert.Equal(t, pkg.I, I)
}
