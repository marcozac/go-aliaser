package aliaser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObject(t *testing.T) {
	t.Run("TypeString", AliaserTest(func(t *testing.T, a *Aliaser) {
		var tn *TypeName
		for _, atn := range a.Types() {
			if atn.Name() == "M" {
				tn = atn
				break
			}
		}
		require.NotNil(t, tn)
		assert.Equal(t, "pkg.M", tn.TypeString())
	}))
}
