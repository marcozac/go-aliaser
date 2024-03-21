package aliaser

import (
	"go/types"
	"testing"

	"github.com/marcozac/go-aliaser/importer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestTypeParam(t *testing.T) {
	LoadedPackageHelper(t, func(t *testing.T, p *packages.Package) {
		obj := types.NewTypeName(0, p.Types, "T", nil)
		tp := NewTypeParam(types.NewTypeParam(obj, types.Typ[types.Int]), importer.New())
		assert.NotNil(t, tp.Underlying())
		assert.Equal(t, "T", tp.String())
		bound := tp.Constraint()
		require.NotNil(t, bound)
		assert.NotNil(t, bound.Underlying())
		assert.Equal(t, "int", bound.String())
	})
}
