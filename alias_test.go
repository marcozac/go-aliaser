package aliaser

import (
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlias(t *testing.T) {
	t.Run("Imports", AliaserTest(func(t *testing.T, a *Aliaser) {
		alias := a.alias
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
			alias.AddImport(pkg)
		}
		imports := alias.Imports()
		l := len(imports)
		assert.GreaterOrEqual(t, l, 3)
		require.Contains(t, imports, pkgPath1)
		require.Contains(t, imports, pkgPath2)
		require.Contains(t, imports, pkgPath3)
		alias1 := imports[pkgPath1]
		alias2 := imports[pkgPath2]
		alias3 := imports[pkgPath3]
		assert.Equal(t, pkgName, alias1)
		assert.Equal(t, alias1, alias.imports[pkgPath1].Name())
		assert.Equal(t, pkgName+"_2", alias2)
		assert.Equal(t, alias2, alias.imports[pkgPath2].Name())
		assert.Equal(t, pkgName+"_3", alias3)
		assert.Equal(t, alias3, alias.imports[pkgPath3].Name())
		// cover nil case
		assert.NotPanics(t, func() { alias.AddImport(nil) })
		assert.Equal(t, l, len(alias.imports))
	}))
}
