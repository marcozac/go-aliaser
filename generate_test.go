package aliaser

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateFile(t *testing.T) {
	t.Run("out/alias.go", func(t *testing.T) {
		const filename = "internal/testing/out/alias.go"
		if err := os.Remove(filename); err != nil {
			require.ErrorIs(t, err, fs.ErrNotExist)
		}
		assert.NoError(t, GenerateFile(TestTarget, TestPattern, filename,
			WithHeader(fmt.Sprintf("%s\n\n%s",
				"// Code generated by aliaser. DO NOT EDIT.",
				"//go:build testout",
			)),
		))
		assert.FileExists(t, filename)
	})
	t.Run("EmptyTarget", func(t *testing.T) {
		assert.ErrorIs(t, GenerateFile("", TestPattern, "alias.go"), ErrEmptyTarget)
	})
	t.Run("EmptyPattern", func(t *testing.T) {
		assert.ErrorIs(t, GenerateFile(TestTarget, "", "alias.go"), ErrEmptyPattern)
	})
}

func TestGenerate(t *testing.T) {
	assert.NoError(t, Generate(TestTarget, TestPattern, io.Discard))
	t.Run("EmptyTarget", func(t *testing.T) {
		assert.ErrorIs(t, Generate("", TestPattern, io.Discard), ErrEmptyTarget)
	})
	t.Run("EmptyPattern", func(t *testing.T) {
		assert.ErrorIs(t, Generate(TestTarget, "", io.Discard), ErrEmptyPattern)
	})
}
