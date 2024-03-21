package aliaser

import (
	"bytes"
	"go/types"
	"strings"

	"github.com/marcozac/go-aliaser/importer"
	"github.com/marcozac/go-aliaser/util/sequence"
)

var _ Object = (*Const)(nil)

// Const is the type used to represent a constant in the loaded package.
// It contains the original constant and the resolver used to generate the
// aliases.
type Const struct {
	*types.Const
	objectResolver
}

// NewConst returns a new [Const] with the given constant. The importer is used
// to add the constant package to the list of imports.
func NewConst(c *types.Const, imp *importer.Importer) *Const {
	return &Const{c, newObjectResolver(c, imp)}
}

var _ Object = (*Var)(nil)

// Var is the type used to represent a variable in the loaded package.
// It contains the original variable and the resolver used to generate the
// aliases.
type Var struct {
	*types.Var
	objectResolver
}

// NewVar returns a new [Var] with the given variable. The importer is used to
// add the variable package to the list of imports.
func NewVar(v *types.Var, imp *importer.Importer) *Var {
	return &Var{v, newObjectResolver(v, imp)}
}

var _ Object = (*Func)(nil)

// Func is the type used to represent a function in the loaded package.
// It contains the original function and the resolver used to generate the
// aliases.
type Func struct {
	*types.Func
	objectResolver
	tsig *Signature
}

// NewFunc returns a new [Func] with the given function. The importer is used to
// add the function package to the list of imports.
func NewFunc(fn *types.Func, imp *importer.Importer) *Func {
	return &Func{fn, newObjectResolver(fn, imp), NewSignature(fn.Type().(*types.Signature), imp)}
}

// WriteSignature returns the signature of the function as a string. It is a wrapper
// around [types.WriteSignature] that uses a custom [types.Qualifier] to resolve
// the package aliases. The signature is wrapped by [Signature.Wrapper] to replace
// the parameter and result names on conflict with the package aliases.
func (fn *Func) WriteSignature() string {
	buf := new(bytes.Buffer)
	types.WriteSignature(buf, fn.tsig.Wrapper(), fn.qualifier())
	return buf.String()
}

// CallArgs returns the arguments of the function as a string. The arguments are
// joined by a comma and the variadic argument is suffixed with "...".
// If the function has no arguments, it returns an empty string.
//
// Example:
//
//	func(a int, b bool, c ...string) // "a, b, c..."
func (fn *Func) CallArgs() string {
	params := fn.tsig.Wrapper().Params()
	l := params.Len()
	if l == 0 {
		return ""
	}
	names := sequence.New(params.Len, func(i int) string { return params.At(i).Name() }).
		SliceFuncIndex(func(s string, i int) string {
			if i == l-1 && fn.tsig.Variadic() {
				return s + "..."
			}
			return s
		})
	return strings.Join(names, ", ")
}

// Returns returns true if the function has results.
func (fn *Func) Returns() bool {
	return fn.tsig.Results().Len() > 0
}

var _ Object = (*TypeName)(nil)

// TypeName is the type used to represent a type in the loaded package.
// It contains the original type and the resolver used to generate the
// aliases.
type TypeName struct {
	*types.TypeName
	objectResolver
}

// NewTypeName returns a new [TypeName] with the given type. The importer is used
// to add the type package to the list of imports.
func NewTypeName(tn *types.TypeName, imp *importer.Importer) *TypeName {
	return &TypeName{tn, newObjectResolver(tn, imp)}
}

type objectResolver struct {
	orig       types.Object
	imp        *importer.Importer
	typeParams *sequence.Sequence[*types.TypeParam]
	typeArgs   *sequence.Sequence[types.Type]
}

func newObjectResolver(obj types.Object, imp *importer.Importer) objectResolver {
	o := objectResolver{orig: obj, imp: imp}
	o.importType(obj.Type())
	return o
}

// PackageAlias returns the alias of the package as declared in the import
// statement.
func (o *objectResolver) PackageAlias() string {
	return o.imp.AliasOf(o.orig.Pkg())
}

// TypeString returns the object type as a string, resolving the package
// names using the aliases declared in the import statements.
func (o *objectResolver) TypeString() string {
	return types.TypeString(o.orig.Type(), o.qualifier())
}

// Generic returns true if the object is generic, that is, if it has type
// parameters that are not resolved by type arguments.
//
// Example:
//
//	type A[T1, T2 any] struct{ Foo T1; Bar T2 } // true
//	type B[T1 any] A[T1, int] // false
//	type C A[int, string] // false
func (o *objectResolver) Generic() bool {
	return len(o.TypeParams()) > 0
}

// TypeParams returns the type parameters of the object as a slice.
func (o *objectResolver) TypeParams() []*types.TypeParam {
	if o.typeParams == nil {
		return nil
	}
	return o.typeParams.Slice()
}

func (o *objectResolver) qualifier() types.Qualifier {
	return func(p *types.Package) string {
		return o.imp.AliasOf(p)
	}
}

func (o *objectResolver) importType(typ types.Type) {
	switch typ := typ.(type) {
	case *types.Named:
		if tn := typ.Obj(); tn != nil {
			o.imp.AddImport(tn.Pkg())
		}
	case *types.Map:
		o.importType(typ.Key())
		o.importType(typ.Elem())
	case interface{ Elem() types.Type }: // *types.Array, *types.Slice, *types.Chan, *types.Pointer
		o.importType(typ.Elem())
	case *types.Signature:
		// do not call o.importType(typ.Params()) and o.importType(typ.Results())
		// to avoid unnecessaty type switches
		sequence.FromSequenceable(typ.Params()).
			ForEach(o.varImporter)
		sequence.FromSequenceable(typ.Results()).
			ForEach(o.varImporter)
	case *types.Struct:
		sequence.New(typ.NumFields, typ.Field).
			ForEach(o.varImporter)
	case *types.Interface:
		sequence.New(typ.NumMethods, typ.Method).
			ForEach(o.funcImporter)
		sequence.New(typ.NumEmbeddeds, typ.EmbeddedType).
			ForEach(o.importType)
	}
	if typ, ok := typ.(interface{ TypeParams() *types.TypeParamList }); ok {
		o.typeParams = sequence.FromSequenceable(typ.TypeParams()).
			ForEach(func(tp *types.TypeParam) {
				o.importType(tp.Constraint())
			})
	}
	if typ, ok := typ.(interface{ TypeArgs() *types.TypeList }); ok {
		o.typeArgs = sequence.FromSequenceable(typ.TypeArgs()).
			ForEach(o.importType)
	}
}

func (o *objectResolver) varImporter(pv *types.Var) {
	o.importObject(pv)
}

func (o *objectResolver) funcImporter(fn *types.Func) {
	o.importObject(fn)
}

func (o *objectResolver) importObject(pt PackageTyper) {
	o.imp.AddImport(pt.Pkg())
	o.importType(pt.Type())
}

// Object extends the [types.Object] interface with methods to get the package
// alias and the resolved type as a string.
type Object interface {
	types.Object
	PackageAliaser
	TypeStringer
}

// PackageTyper is a subset of the [types.Object] interface that provides
// methods to get the package and the type of the object.
type PackageTyper interface {
	Pkg() *types.Package
	Type() types.Type
}

// PackageAliaser is the interface implemented by types that can provide the
// alias of the package they belong to.
type PackageAliaser interface {
	// PackageAlias returns the alias of the package as declared in the
	// import statement.
	PackageAlias() string
}

// TypeStringer is the interface implemented by types that can return their type
// as a string.
type TypeStringer interface {
	// TypeString returns the type of the implementing type as a string.
	TypeString() string
}
