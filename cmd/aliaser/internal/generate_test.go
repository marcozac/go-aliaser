package internal

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestPattern is a valid pattern for testing.
const TestPattern = "github.com/marcozac/go-aliaser/internal/testing/pkg"

func TestGenerateCmd(t *testing.T) {
	t.Run("RequiredFlags", func(t *testing.T) {
		root, buf := NewTestRoot(t)
		root.SetArgs([]string{"generate"})
		assert.Error(t, root.Execute())
		assert.Contains(t, buf.String(), "required flag(s) \"pattern\", \"target\" not set")
	})
	t.Run("DryRun", func(t *testing.T) {
		root, buf := NewTestRoot(t)
		root.SetArgs([]string{
			"generate", "--dry-run",
			"--target", "foo",
			"--pattern", TestPattern,
		})
		assert.NoError(t, root.Execute())
		assert.Contains(t, buf.String(), "package foo")
	})
	t.Run("ExcludeAll", func(t *testing.T) {
		root, buf := NewTestRoot(t)
		root.SetArgs([]string{
			"generate", "--dry-run",
			"--target", "foo",
			"--pattern", TestPattern,
			"--exclude-constants",
			"--exclude-variables",
			"--exclude-functions",
			"--exclude-types",
		})
		assert.NoError(t, root.Execute())
		assert.NotContains(t, buf.String(), "const")
		assert.NotContains(t, buf.String(), "var")
		assert.NotContains(t, buf.String(), "type")
	})
	t.Run("ExcludeNames", func(t *testing.T) {
		root, buf := NewTestRoot(t)
		root.SetArgs([]string{
			"generate", "--dry-run",
			"--target", "foo",
			"--pattern", TestPattern,
			"--exclude-names", "A,C",
		})
		assert.NoError(t, root.Execute())
		assert.NotContains(t, buf.String(), "A = pkg.A")
		assert.NotContains(t, buf.String(), "C = pkg.C")
	})
	t.Run("Header", func(t *testing.T) {
		root, buf := NewTestRoot(t)
		root.SetArgs([]string{
			"generate", "--dry-run",
			"--target", "foo",
			"--pattern", TestPattern,
			"--header", "// my header",
		})
		assert.NoError(t, root.Execute())
		assert.Contains(t, buf.String(), "// my header")
	})
	t.Run("GenerateFile", func(t *testing.T) {
		tempDir := t.TempDir()
		filename := filepath.Join(tempDir, "alias.go")
		root, _ := NewTestRoot(t)
		root.SetArgs([]string{
			"generate",
			"--target", "foo",
			"--pattern", TestPattern,
			"--file", filename,
		})
		assert.NoError(t, root.Execute())
		assert.FileExists(t, filename)
	})
	t.Run("FileError", func(t *testing.T) {
		root, buf := NewTestRoot(t)
		root.SetArgs([]string{
			"generate",
			"--target", "foo",
			"--pattern", TestPattern,
		})
		assert.Error(t, root.Execute())
		assert.Contains(t, buf.String(), "at least one of the flags in the group [file dry-run] is required")
	})
	t.Run("AliaserError", func(t *testing.T) {
		root, buf := NewTestRoot(t)
		root.SetArgs([]string{
			"generate", "--dry-run",
			"--target", "foo",
			"--pattern", "golang.org/x/tools/go/*",
		})
		assert.Error(t, root.Execute())
		assert.Contains(t, buf.String(), "aliaser: package errors:")
	})
}

// NewTestRoot returns a new root command and a buffer to capture its output
// and error.
func NewTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	buf := new(bytes.Buffer)
	cmd := NewRoot()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	return cmd, buf
}

// NewTestGenerate returns a new generate command and a buffer to capture its
// output and error.
func NewTestGenerate(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	buf := new(bytes.Buffer)
	cmd := NewGenerate()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	return cmd, buf
}
