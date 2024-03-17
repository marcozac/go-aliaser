package aliaser

import (
	"go/types"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImporter(t *testing.T) {
	var imp *Importer
	AliaserTest(func(t *testing.T, a *Aliaser) {
		imp = a.alias.Importer
	})(t)
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
		assert.GreaterOrEqual(t, l, 3)
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
	t.Run("Merge", func(t *testing.T) {
		imp2 := NewImporter()
		imp2.AddImport(types.NewPackage(pkgPath1, pkgName)) // already exists
		imp2.AddImport(types.NewPackage(pkgPath2, pkgName)) // already exists
		imp2.AddImport(types.NewPackage("github.com/marcozac/go-aliaser/fake4", pkgName))
		imp2.AddImport(types.NewPackage("github.com/marcozac/go-aliaser/fake5", pkgName))
		imp.Merge(imp2)
		assert.Equal(t, l+2, len(imp.Imports()))
	})
	l = len(imp.Imports()) // update l after merge
	t.Run("Concurrent", func(t *testing.T) {
		imp2 := NewImporter()
		imp2.AddImport(types.NewPackage(pkgPath1, pkgName)) // already exists
		const n = 100
		var wg sync.WaitGroup
		wg.Add(n)
		assert.NotPanics(t, func() {
			for i := 0; i < n; i++ {
				switch {
				case i < 25:
					go func() {
						defer wg.Done()
						imp.AddImport(types.NewPackage(pkgPath1, pkgName)) // already exists
					}()
				case i >= 25 && i < 50:
					go func() {
						defer wg.Done()
						_ = imp.AliasedImports()
					}()
				case i >= 50 && i < 75:
					go func() {
						defer wg.Done()
						_ = imp.Imports()
					}()
				default:
					go func() {
						defer wg.Done()
						imp.Merge(imp2)
					}()
				}
			}
		})
		wg.Wait()
	})
}
