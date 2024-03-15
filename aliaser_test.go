//go:build !testout

package aliaser

import (
	"context"
	"embed"
	"go/types"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		src, err := Load("github.com/marcozac/go-aliaser/internal/testdata")
		assert.NoError(t, err)
		require.NotNil(t, src)

		assert.Greater(t, len(src.Constants), 0)
		assert.Greater(t, len(src.Variables), 0)
		assert.Greater(t, len(src.Functions), 0)
		assert.Greater(t, len(src.Types), 0)
	})
	t.Run("Options", func(t *testing.T) {
		t.Run("WithContext", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			c := applyOptions(nil, WithContext(ctx))
			_, err := load(c, "github.com/marcozac/go-aliaser/internal/testdata")
			assert.Error(t, err, "context canceled")
			t.Logf("context canceled error: %v", err)
		})
		t.Run("ExcludeConstants", func(t *testing.T) {
			src, err := Load("github.com/marcozac/go-aliaser/internal/testdata", ExcludeConstants())
			assert.NoError(t, err)
			require.NotNil(t, src)
			assert.Empty(t, src.Constants)
		})
		t.Run("ExcludeVariables", func(t *testing.T) {
			src, err := Load("github.com/marcozac/go-aliaser/internal/testdata", ExcludeVariables())
			assert.NoError(t, err)
			require.NotNil(t, src)
			assert.Empty(t, src.Variables)
		})
		t.Run("ExcludeFunctions", func(t *testing.T) {
			src, err := Load("github.com/marcozac/go-aliaser/internal/testdata", ExcludeFunctions())
			assert.NoError(t, err)
			require.NotNil(t, src)
			assert.Empty(t, src.Functions)
		})
		t.Run("ExcludeTypes", func(t *testing.T) {
			src, err := Load("github.com/marcozac/go-aliaser/internal/testdata", ExcludeTypes())
			assert.NoError(t, err)
			require.NotNil(t, src)
			assert.Empty(t, src.Types)
		})
		t.Run("ExcludeNames", func(t *testing.T) {
			src, err := Load("github.com/marcozac/go-aliaser/internal/testdata", ExcludeNames("A", "D"))
			assert.NoError(t, err)
			require.NotNil(t, src)
			for _, c := range src.Constants {
				assert.NotEqual(t, "A", c.Name())
			}
			for _, typ := range src.Types {
				assert.NotEqual(t, "D", typ.Name())
			}
		})
	})
	t.Run("Error", func(t *testing.T) {
		_, err := Load("")
		assert.Error(t, err, "empty path")
		t.Logf("empty path error: %v", err)

		_, err = Load("github.com/marcozac/go-aliaser/internal/doesnotexist")
		assert.Error(t, err, "non-existent path")
		t.Logf("non-existent path error: %v", err)

		_, err = Load("golang.org/x/tools/go/*")
		assert.Error(t, err, "invalid pattern")
		t.Logf("invalid pattern error: %v", err)

		_, err = Load("golang.org/x/tools/go/...")
		assert.Error(t, err, "too many packages")
		t.Logf("too many packages error: %v", err)

		t.Run("FakeDriver", func(t *testing.T) {
			// Run in a subtest to not pollute the environment
			t.Setenv("GOPACKAGESDRIVER", "fake")
			_, err = Load("github.com/marcozac/go-aliaser/internal/testdata")
			assert.Error(t, err, "fake driver")
			t.Logf("fake driver error: %v", err)
		})
	})
}

func TestGenerate(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		src, err := Load("github.com/marcozac/go-aliaser/internal/testdata")
		assert.NoError(t, err)
		require.NotNil(t, src)
		a := &Alias{
			PkgName: "testout",
			Out:     "internal/testout/alias.go",
			Src:     src,
			Header:  "//go:build testout",
		}
		assert.NoError(t, Generate(a))
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("ParseTemplate", func(t *testing.T) {
			exTmplFS := tmplFS
			tmplFS = embed.FS{}
			defer func() { tmplFS = exTmplFS }()
			err := Generate(&Alias{})
			assert.Error(t, err, "parse template")
			t.Logf("parse template error: %v", err)
		})

		err := Generate(&Alias{})
		assert.Error(t, err, "execute template")
		t.Logf("execute template error: %v", err)

		err = Generate(&Alias{
			PkgName: "testout",
			Out:     "non-existent-dir/alias.go",
			Src: &Src{
				PkgPath: "github.com/marcozac/go-aliaser/internal/testdata",
			},
		})
		assert.Error(t, err, "open file")
		t.Logf("open file error: %v", err)

		t.Run("ImportProcess", func(t *testing.T) {
			src, err := Load("github.com/marcozac/go-aliaser/internal/testdata")
			assert.NoError(t, err)
			require.NotNil(t, src)

			// Change the first constant to have an invalid name
			c0 := src.Constants[0]
			src.Constants[0] = types.NewConst(c0.Pos(), c0.Pkg(), c0.Name()+".", c0.Type(), c0.Val())

			dir := t.TempDir()
			f, err := os.CreateTemp(dir, "alias-*.go")
			require.NoError(t, err)
			f.Close()

			err = Generate(&Alias{
				PkgName: "testout",
				Out:     f.Name(),
				Src:     src,
			})
			assert.Error(t, err, "import process")
			t.Logf("import process error: %v", err)
		})
	})
}
