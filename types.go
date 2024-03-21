package aliaser

import (
	"go/types"
	"slices"
	"sync"

	"github.com/marcozac/go-aliaser/importer"
	"github.com/marcozac/go-aliaser/util/sequence"
)

// Signature is the type used to represent a function signature in the loaded
// package. It contains the original signature and the importer used to generate
// the aliases. It also contains a wrapper signature that replaces the parameter
// and result names on conflict with the package aliases.
type Signature struct {
	*types.Signature
	imp     *importer.Importer
	wrapper *types.Signature
	once    sync.Once
}

// NewSignature returns a new [Signature] with the given signature. The importer
// is used to add the signature package to the list of imports.
func NewSignature(sig *types.Signature, imp *importer.Importer) *Signature {
	return &Signature{Signature: sig, imp: imp}
}

// Wrapper returns a [types.Signature] that wraps the original one, replacing
// the parameter and result names on conflict with the package aliases.
//
// Since it calls the [importer.Importer.AliasedImports] method for initializing
// the wrapper and further calls will return the same wrapper, it should be called
// only after the package has been fully loaded.
//
// Example:
//
//	import "github.com/google/uuid"
//
//	func Must(uuid_ uuid.UUID, err error) uuid.UUID {
//		return uuid.Must(uuid_, err)
//	}
func (s *Signature) Wrapper() *types.Signature {
	s.once.Do(func() {
		ai := s.imp.AliasedImports()
		aliases := make([]string, 0, len(ai))
		for _, alias := range ai {
			aliases = append(aliases, alias)
		}
		typeParams := make([]*types.TypeParam, s.TypeParams().Len())
		sequence.FromSequenceable(s.TypeParams()).
			ForEachIndex(func(tp *types.TypeParam, i int) {
				typeParams[i] = types.NewTypeParam(tp.Obj(), tp.Constraint())
			})
		s.wrapper = types.NewSignatureType(
			s.Recv(), // always nil
			nil,      // wrap funcs, not methods
			typeParams,
			NewAliasedTuple(aliases, s.Params()),
			NewAliasedTuple(aliases, s.Results()),
			s.Variadic(),
		)
	})
	return s.wrapper
}

// NewAliasedTuple returns a new tuple ensuring none of its variable names
// is in the given aliases list. If a variable name is in the list, it is
// suffixed with an underscore.
func NewAliasedTuple(aliases []string, tuple *types.Tuple) *types.Tuple {
	return types.NewTuple(sequence.FromSequenceable(tuple).SliceFunc(func(pv *types.Var) *types.Var {
		if slices.Contains(aliases, pv.Name()) {
			return types.NewVar(pv.Pos(), pv.Pkg(), pv.Name()+"_", pv.Type())
		}
		return pv
	})...)
}
