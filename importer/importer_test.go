package importer

import (
	"go/types"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImporter(t *testing.T) {
	imp := New()
	const (
		pkgName  = "fake"
		pkgPath1 = "github.com/marcozac/go-aliaser/fake1"
		pkgPath2 = "github.com/marcozac/go-aliaser/fake2"
		pkgPath3 = "github.com/marcozac/go-aliaser/fake3"
	)
	pm := map[string]string{
		pkgPath1: pkgName,
		pkgPath2: pkgName,
		pkgPath3: pkgName,
	}
	for path, name := range pm {
		pkg := types.NewPackage(path, name)
		imp.AddImport(pkg)
	}
	l := len(imp.Imports())
	t.Run("AddNil", func(t *testing.T) {
		// cover nil case
		assert.NotPanics(t, func() { imp.AddImport(nil) })
		assert.Equal(t, l, len(imp.Imports()))
	})
	t.Run("AliasedImports", func(t *testing.T) {
		imports := imp.AliasedImports()
		assert.Equal(t, l, 3)
		require.Contains(t, imports, pkgPath1)
		require.Contains(t, imports, pkgPath2)
		require.Contains(t, imports, pkgPath3)
		alias1 := imports[pkgPath1]
		alias2 := imports[pkgPath2]
		alias3 := imports[pkgPath3]
		assert.Equal(t, pkgName, alias1)
		assert.Equal(t, alias1, imp.imports[pkgPath1].Name())
		assert.Equal(t, pkgName+"_2", alias2)
		assert.Equal(t, alias2, imp.imports[pkgPath2].Name())
		assert.Equal(t, pkgName+"_3", alias3)
		assert.Equal(t, alias3, imp.imports[pkgPath3].Name())
	})
	t.Run("AliasOf", func(t *testing.T) {
		imp.toAlias = true // force re-aliasing
		assert.Equal(t, pkgName, imp.AliasOf(pkgPath1))
		assert.Equal(t, pkgName+"_2", imp.AliasOf(pkgPath2))
		assert.Equal(t, pkgName+"_3", imp.AliasOf(pkgPath3))
		assert.Empty(t, imp.AliasOf("unknown"))
	})
	t.Run("Concurrent", func(t *testing.T) {
		imp2 := New()
		imp2.AddImport(types.NewPackage(pkgPath1, pkgName)) // already exists
		const n = 150
		var wg sync.WaitGroup
		wg.Add(n)
		assert.NotPanics(t, func() {
			for i := 0; i < n; i++ {
				switch {
				case i < 50:
					go func() {
						defer wg.Done()
						imp.AddImport(types.NewPackage(pkgPath1, pkgName)) // already exists
					}()
				case i >= 50 && i < 100:
					go func() {
						defer wg.Done()
						_ = imp.AliasedImports()
					}()
				case i >= 100 && i < 150:
					go func() {
						defer wg.Done()
						_ = imp.Imports()
					}()
				}
			}
		})
		wg.Wait()
	})
}
