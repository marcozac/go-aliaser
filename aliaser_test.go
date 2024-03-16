package aliaser

import (
	"context"
	"embed"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestAliaserOptions(t *testing.T) {
	t.Run("WithContext", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := New(TestTarget, TestPattern, WithContext(ctx))
		assert.Error(t, err)
	})
	t.Run("ExcludeConstants", func(t *testing.T) {
		AliaserTest(t, func(t *testing.T, a *Aliaser) {
			assert.Empty(t, a.alias.Constants)
		}, ExcludeConstants())
	})
	t.Run("ExcludeVariables", func(t *testing.T) {
		AliaserTest(t, func(t *testing.T, a *Aliaser) {
			assert.Empty(t, a.alias.Variables)
		}, ExcludeVariables())
	})
	t.Run("ExcludeFunctions", func(t *testing.T) {
		AliaserTest(t, func(t *testing.T, a *Aliaser) {
			assert.Empty(t, a.alias.Functions)
		}, ExcludeFunctions())
	})
	t.Run("ExcludeTypes", func(t *testing.T) {
		AliaserTest(t, func(t *testing.T, a *Aliaser) {
			assert.Empty(t, a.alias.Types)
		}, ExcludeTypes())
	})
	t.Run("ExcludeNames", func(t *testing.T) {
		AliaserTest(t, func(t *testing.T, a *Aliaser) {
			ObjectBatchTest(t, a.alias.Constants, func(t *testing.T, o types.Object) {
				assert.NotEqual(t, "A", o.Name())
				assert.NotEqual(t, "D", o.Name())
			})
			ObjectBatchTest(t, a.alias.Variables, func(t *testing.T, o types.Object) {
				assert.NotEqual(t, "A", o.Name())
				assert.NotEqual(t, "D", o.Name())
			})
			ObjectBatchTest(t, a.alias.Functions, func(t *testing.T, o types.Object) {
				assert.NotEqual(t, "A", o.Name())
				assert.NotEqual(t, "D", o.Name())
			})
			ObjectBatchTest(t, a.alias.Types, func(t *testing.T, o types.Object) {
				assert.NotEqual(t, "A", o.Name())
				assert.NotEqual(t, "D", o.Name())
			})
		}, ExcludeNames("A", "D"))
	})
}

func TestAliaserError(t *testing.T) {
	t.Run("InvalidPattern", func(t *testing.T) {
		_, err := New(TestTarget, "golang.org/x/tools/go/*")
		assert.Error(t, err)
	})
	t.Run("TooManyPackages", func(t *testing.T) {
		_, err := New(TestTarget, "golang.org/x/tools/go/...")
		assert.Error(t, err)
	})
	t.Run("Load", func(t *testing.T) {
		t.Setenv("GOPACKAGESDRIVER", "fakedriver")
		_, err := New(TestTarget, TestPattern)
		assert.Error(t, err)
	})
	t.Run("ParseTemplate", func(t *testing.T) {
		oldFS := tmplFS
		defer func() { tmplFS = oldFS }()
		tmplFS = embed.FS{}
		a := &Aliaser{}
		assert.Error(t, a.Generate(io.Discard)) // empty a.alias
	})
	t.Run("ExecuteTemplate", func(t *testing.T) {
		a := &Aliaser{}
		assert.Error(t, a.Generate(io.Discard)) // empty a.alias
	})
	t.Run("Format", func(t *testing.T) {
		AliaserTest(t, func(t *testing.T, a *Aliaser) {
			c0 := a.alias.Constants[0] // change the first constant to have an invalid name
			a.alias.Constants[0] = types.NewConst(c0.Pos(), c0.Pkg(), c0.Name()+".", c0.Type(), c0.Val())
			assert.Error(t, a.Generate(io.Discard))
		})
	})
	t.Run("Write", func(t *testing.T) {
		AliaserTest(t, func(t *testing.T, a *Aliaser) {
			assert.ErrorIs(t, a.Generate(WriterE{}), assert.AnError)
		})
	})
	t.Run("File", func(t *testing.T) {
		tempDir := t.TempDir()
		nwDir := filepath.Join(tempDir, "non-writable-dir")
		require.NoError(t, os.Mkdir(nwDir, 0o555))
		AliaserTest(t, func(t *testing.T, a *Aliaser) {
			t.Run("Mkdir", func(t *testing.T) {
				assert.Error(t, a.GenerateFile(filepath.Join(nwDir, "out/alias.go")))
			})
			t.Run("OpenFile", func(t *testing.T) {
				assert.Error(t, a.GenerateFile(filepath.Join(nwDir, "alias.go")))
			})
		})
	})
}

func TestSetAlias(t *testing.T) {
	t.Run("NotExportedObject", func(t *testing.T) {
		AliaserTest(t, func(t *testing.T, a *Aliaser) {
			LoadedPackageTest(t, func(t *testing.T, p *packages.Package) {
				p.Types.Scope().Insert(
					types.NewConst(0, p.Types, "notExported", types.Typ[types.Uint8], nil),
				)
				assert.NoError(t, a.setAlias(a.alias.Config, p)) // cover not exported object
			})
		})
	})
	t.Run("ObjectTypeError", func(t *testing.T) {
		AliaserTest(t, func(t *testing.T, a *Aliaser) {
			LoadedPackageTest(t, func(t *testing.T, p *packages.Package) {
				p.Types.Scope().Insert(
					types.NewLabel(0, p.Types, "MyLabel"),
				)
				assert.Error(t, a.setAlias(a.alias.Config, p))
			})
		})
	})
}

// AliaserTest is a helper function that creates a new valid Aliaser with
// [TestTarget], [TestPattern] and the options, then, it calls the given function.
func AliaserTest(t *testing.T, fn func(*testing.T, *Aliaser), opts ...Option) {
	t.Helper()
	a, err := New(TestTarget, TestPattern, opts...)
	assert.NoError(t, err)
	require.NotNil(t, a.alias)
	fn(t, a)
}

func LoadedPackageTest(t *testing.T, fn func(*testing.T, *packages.Package)) {
	t.Helper()
	pkgs, err := packages.Load(&packages.Config{Mode: loadMode}, TestPattern)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	p := pkgs[0]
	fn(t, p)
}

func ObjectBatchTest[TO types.Object](t *testing.T, objs []TO, fn func(*testing.T, types.Object)) {
	t.Helper()
	for _, obj := range objs {
		fn(t, obj)
	}
}
