package aliaser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	src, err := Load("github.com/marcozac/go-aliaser/internal/testdata")
	assert.NoError(t, err)
	require.NotNil(t, src)

	assert.Greater(t, len(src.Constants), 0)
	assert.Greater(t, len(src.Variables), 0)
	assert.Greater(t, len(src.Functions), 0)
	assert.Greater(t, len(src.Types), 0)

	_, err = Load("")
	assert.Error(t, err, "expected error for empty path")
}
