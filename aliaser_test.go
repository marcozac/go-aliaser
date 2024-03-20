package aliaser

import (
	"bytes"
	"context"
	"embed"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/marcozac/go-aliaser/util/sequence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestSetAlias(t *testing.T) {
	t.Run("NotExportedObject", AliaserTest(func(t *testing.T, a *Aliaser) {
		LoadedPackageHelper(t, func(t *testing.T, p *packages.Package) {
			p.Types.Scope().Insert(
				types.NewConst(0, p.Types, "notExported", types.Typ[types.Uint8], nil),
			)
			assert.NoError(t, a.addPkgObjects(p)) // cover not exported object
		})
	}))
}

func TestAliaserOptions(t *testing.T) {
	t.Run("WithContext", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := New(&Config{TargetPackage: TestTarget, Pattern: TestPattern}, WithContext(ctx))
		assert.Error(t, err)
	})
	t.Run("ExcludeConstants", AliaserTest(func(t *testing.T, a *Aliaser) {
		assert.Empty(t, a.Constants())
	}, ExcludeConstants()))
	t.Run("ExcludeVariables", AliaserTest(func(t *testing.T, a *Aliaser) {
		assert.Empty(t, a.Variables())
	}, ExcludeVariables()))
	t.Run("ExcludeFunctions", AliaserTest(func(t *testing.T, a *Aliaser) {
		assert.Empty(t, a.Functions())
	}, ExcludeFunctions()))
	t.Run("ExcludeTypes", AliaserTest(func(t *testing.T, a *Aliaser) {
		assert.Empty(t, a.Types())
	}, ExcludeTypes()))
	t.Run("ExcludeNames", AliaserTest(func(t *testing.T, a *Aliaser) {
		ObjectBatchHelper(t, a.Constants(), func(t *testing.T, o types.Object) {
			assert.NotEqual(t, "A", o.Name())
			assert.NotEqual(t, "D", o.Name())
		})
		ObjectBatchHelper(t, a.Variables(), func(t *testing.T, o types.Object) {
			assert.NotEqual(t, "A", o.Name())
			assert.NotEqual(t, "D", o.Name())
		})
		ObjectBatchHelper(t, a.Functions(), func(t *testing.T, o types.Object) {
			assert.NotEqual(t, "A", o.Name())
			assert.NotEqual(t, "D", o.Name())
		})
		ObjectBatchHelper(t, a.Types(), func(t *testing.T, o types.Object) {
			assert.NotEqual(t, "A", o.Name())
			assert.NotEqual(t, "D", o.Name())
		})
	}, ExcludeNames("A", "D")))
	t.Run("AssignFunctions", func(t *testing.T) {
		t.Run("True", AliaserTest(func(t *testing.T, a *Aliaser) {
			var buf bytes.Buffer
			require.NoError(t, a.Generate(&buf))
			assert.NotContains(t, buf.String(), "func J(")
		}, AssignFunctions(true)))
		t.Run("False", AliaserTest(func(t *testing.T, a *Aliaser) {
			var buf bytes.Buffer
			require.NoError(t, a.Generate(&buf))
			assert.Contains(t, buf.String(), "func J(")
		}, AssignFunctions(false)))
	})
	t.Run("OnDuplicate", func(t *testing.T) {
		t.Run("Skip", AliaserTest(func(t *testing.T, a *Aliaser) {
			v0 := a.variables[0]
			a.AddVariables(types.NewVar(0, v0.Pkg(), "A", types.Typ[types.Uint8]))
			assert.False(t, slices.ContainsFunc(a.variables, func(c *Var) bool { return c.Name() == "A" }))
		}, OnDuplicate(OnDuplicateSkip)))
		t.Run("Replace", AliaserTest(func(t *testing.T, a *Aliaser) {
			pkg := a.variables[0].Pkg()
			assert.True(t, slices.ContainsFunc(a.constants, func(c *Const) bool { return c.Name() == "A" }))

			a.AddVariables(types.NewVar(0, pkg, "A", types.Typ[types.Uint8]))
			assert.False(t, slices.ContainsFunc(a.constants, func(c *Const) bool { return c.Name() == "A" }))
			assert.True(t, slices.ContainsFunc(a.variables, func(c *Var) bool { return c.Name() == "A" }))

			a.AddConstants(types.NewConst(0, pkg, "A", types.Typ[types.Uint8], nil))
			assert.False(t, slices.ContainsFunc(a.variables, func(c *Var) bool { return c.Name() == "A" }))
			assert.True(t, slices.ContainsFunc(a.constants, func(c *Const) bool { return c.Name() == "A" }))

			f0 := a.functions[0]
			a.AddFunctions(types.NewFunc(0, pkg, "A", types.NewSignatureType(
				f0.tsig.Recv(),
				sequence.FromSequenceable(f0.tsig.RecvTypeParams()).Slice(),
				sequence.FromSequenceable(f0.tsig.TypeParams()).Slice(),
				f0.tsig.Params(),
				f0.tsig.Results(),
				f0.tsig.Variadic(),
			)))
			assert.False(t, slices.ContainsFunc(a.constants, func(c *Const) bool { return c.Name() == "A" }))
			assert.True(t, slices.ContainsFunc(a.functions, func(c *Func) bool { return c.Name() == "A" }))

			a.AddTypes(types.NewTypeName(0, pkg, "A", types.Typ[types.Uint8]))
			assert.False(t, slices.ContainsFunc(a.functions, func(c *Func) bool { return c.Name() == "A" }))
			assert.True(t, slices.ContainsFunc(a.types, func(c *TypeName) bool { return c.Name() == "A" }))

			// replace the type
			a.AddConstants(types.NewConst(0, pkg, "A", types.Typ[types.Uint8], nil))
			assert.False(t, slices.ContainsFunc(a.types, func(c *TypeName) bool { return c.Name() == "A" }))
			assert.True(t, slices.ContainsFunc(a.constants, func(c *Const) bool { return c.Name() == "A" }))
		}, OnDuplicate(OnDuplicateReplace)))
		t.Run("Panic", AliaserTest(func(t *testing.T, a *Aliaser) {
			v0 := a.variables[0]
			assert.Panics(t, func() { a.AddVariables(types.NewVar(0, v0.Pkg(), "A", types.Typ[types.Uint8])) })
		}, OnDuplicate(OnDuplicatePanic)))
	})
}

func TestAliaserError(t *testing.T) {
	// EmptyTarget and EmptyPattern are covered in the TestGenerate* tests
	t.Run("NilConfig", func(t *testing.T) {
		_, err := New(nil)
		assert.ErrorIs(t, err, ErrNilConfig)
	})
	t.Run("InvalidPattern", func(t *testing.T) {
		_, err := New(&Config{TargetPackage: TestTarget, Pattern: "golang.org/x/tools/go/*"})
		assert.Error(t, err)
	})
	t.Run("TooManyPackages", func(t *testing.T) {
		_, err := New(&Config{TargetPackage: TestTarget, Pattern: "golang.org/x/tools/go/..."})
		assert.Error(t, err)
	})
	t.Run("Load", func(t *testing.T) {
		t.Setenv("GOPACKAGESDRIVER", "fakedriver")
		_, err := New(&Config{TargetPackage: TestTarget, Pattern: TestPattern})
		assert.Error(t, err)
	})
	t.Run("ObjectTypeError", AliaserTest(func(t *testing.T, a *Aliaser) {
		LoadedPackageHelper(t, func(t *testing.T, p *packages.Package) {
			p.Types.Scope().Insert(
				types.NewLabel(0, p.Types, "MyLabel"),
			)
			assert.Error(t, a.addPkgObjects(p))
		})
	}))
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
		c0 := a.Constants()[0] // change the first constant to have an invalid name
		a.constants[0] = NewConst(types.NewConst(c0.Pos(), c0.Pkg(), c0.Name()+".", c0.Type(), c0.Val()), a.Importer)
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
	t.Run("addObjectName", AliaserTest(func(t *testing.T, a *Aliaser) {
		a.onDuplicate = 10
		v := types.NewVar(0, a.variables[0].Pkg(), "A", types.Typ[types.Uint8])
		assert.Panics(t, func() { a.addObjectName(v, 10) })
	}))
	t.Run("deleteObject", AliaserTest(func(t *testing.T, a *Aliaser) {
		v := types.NewVar(0, a.variables[0].Pkg(), "A", types.Typ[types.Uint8])
		assert.Panics(t, func() { a.deleteObject(v, 10) })
	}))
}

// AliaserTest returns a subtest function, compatible with the [testing.T.Run]
// method, that creates a new valid [Aliaser] with [TestTarget], [TestPattern]
// and the options, then, it calls the given function.
func AliaserTest(fn func(*testing.T, *Aliaser), opts ...Option) func(t *testing.T) {
	return func(t *testing.T) {
		a, err := New(&Config{TargetPackage: TestTarget, Pattern: TestPattern}, opts...)
		assert.NoError(t, err)
		require.NotNil(t, a)
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
