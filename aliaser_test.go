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
	t.Run("ExcludeConstants", AliaserTest(func(t *testing.T, a *Aliaser) {
		assert.Empty(t, a.alias.Constants)
	}, ExcludeConstants()))
	t.Run("ExcludeVariables", AliaserTest(func(t *testing.T, a *Aliaser) {
		assert.Empty(t, a.alias.Variables)
	}, ExcludeVariables()))
	t.Run("ExcludeFunctions", AliaserTest(func(t *testing.T, a *Aliaser) {
		assert.Empty(t, a.alias.Functions)
	}, ExcludeFunctions()))
	t.Run("ExcludeTypes", AliaserTest(func(t *testing.T, a *Aliaser) {
		assert.Empty(t, a.alias.Types)
	}, ExcludeTypes()))
	t.Run("ExcludeNames", AliaserTest(func(t *testing.T, a *Aliaser) {
		ObjectBatchHelper(t, a.alias.Constants, func(t *testing.T, o types.Object) {
			assert.NotEqual(t, "A", o.Name())
			assert.NotEqual(t, "D", o.Name())
		})
		ObjectBatchHelper(t, a.alias.Variables, func(t *testing.T, o types.Object) {
			assert.NotEqual(t, "A", o.Name())
			assert.NotEqual(t, "D", o.Name())
		})
		ObjectBatchHelper(t, a.alias.Functions, func(t *testing.T, o types.Object) {
			assert.NotEqual(t, "A", o.Name())
			assert.NotEqual(t, "D", o.Name())
		})
		ObjectBatchHelper(t, a.alias.Types, func(t *testing.T, o types.Object) {
			assert.NotEqual(t, "A", o.Name())
			assert.NotEqual(t, "D", o.Name())
		})
	}, ExcludeNames("A", "D")))
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
	t.Run("Format", AliaserTest(func(t *testing.T, a *Aliaser) {
		c0 := a.alias.Constants[0] // change the first constant to have an invalid name
		a.alias.Constants[0] = types.NewConst(c0.Pos(), c0.Pkg(), c0.Name()+".", c0.Type(), c0.Val())
		assert.Error(t, a.Generate(io.Discard))
	}))
	t.Run("Write", AliaserTest(func(t *testing.T, a *Aliaser) {
		assert.ErrorIs(t, a.Generate(WriterE{}), assert.AnError)
	}))
	t.Run("File", AliaserTest(func(t *testing.T, a *Aliaser) {
		tempDir := t.TempDir()
		nwDir := filepath.Join(tempDir, "non-writable-dir")
		require.NoError(t, os.Mkdir(nwDir, 0o555))
		t.Run("Mkdir", func(t *testing.T) {
			assert.Error(t, a.GenerateFile(filepath.Join(nwDir, "out/alias.go")))
		})
		t.Run("OpenFile", func(t *testing.T) {
			assert.Error(t, a.GenerateFile(filepath.Join(nwDir, "alias.go")))
		})
	}))
}

func TestSetAlias(t *testing.T) {
	t.Run("NotExportedObject", AliaserTest(func(t *testing.T, a *Aliaser) {
		LoadedPackageHelper(t, func(t *testing.T, p *packages.Package) {
			p.Types.Scope().Insert(
				types.NewConst(0, p.Types, "notExported", types.Typ[types.Uint8], nil),
			)
			assert.NoError(t, a.setAlias(a.alias.Config, p)) // cover not exported object
		})
	}))
	t.Run("ObjectTypeError", AliaserTest(func(t *testing.T, a *Aliaser) {
		LoadedPackageHelper(t, func(t *testing.T, p *packages.Package) {
			p.Types.Scope().Insert(
				types.NewLabel(0, p.Types, "MyLabel"),
			)
			assert.Error(t, a.setAlias(a.alias.Config, p))
		})
	}))
}

// AliaserTest returns a subtest function, compatible with the [testing.T.Run]
// method, that creates a new valid [Aliaser] with [TestTarget], [TestPattern]
// and the options, then, it calls the given function.
func AliaserTest(fn func(*testing.T, *Aliaser), opts ...Option) func(t *testing.T) {
	return func(t *testing.T) {
		a, err := New(TestTarget, TestPattern, opts...)
		assert.NoError(t, err)
		require.NotNil(t, a.alias)
		fn(t, a)
	}
}

// LoadedPackageHelper is a helper function that load a valid [packages.Package]
// with [TestPattern] and calls the given function with t and the package.
func LoadedPackageHelper(t *testing.T, fn func(*testing.T, *packages.Package)) {
	t.Helper()
	pkgs, err := packages.Load(&packages.Config{Mode: loadMode}, TestPattern)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	p := pkgs[0]
	fn(t, p)
}

// ObjectBatchHelper is a helper function that iterates over the given objects
// and calls the given function with t and the object.
func ObjectBatchHelper[TO types.Object](t *testing.T, objs []TO, fn func(*testing.T, types.Object)) {
	t.Helper()
	for _, obj := range objs {
		fn(t, obj)
	}
}
