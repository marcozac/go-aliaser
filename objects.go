package aliaser

import (
	"bytes"
	"go/types"
	"strings"

	"github.com/marcozac/go-aliaser/importer"
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
	tsig *types.Signature
}

// NewFunc returns a new [Func] with the given function. The importer is used to
// add the function package to the list of imports.
func NewFunc(fn *types.Func, imp *importer.Importer) *Func {
	nfn := &Func{fn, newObjectResolver(fn, imp), fn.Type().(*types.Signature)}
	return nfn
}

// Signature returns the signature of the function as a string. It is a wrapper
// around [types.WriteSignature] that uses a custom [types.Qualifier] to resolve
// the package aliases.
func (fn *Func) Signature() string {
	buf := new(bytes.Buffer)
	types.WriteSignature(buf, fn.tsig, fn.qualifier())
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
	params := fn.tsig.Params()
	l := params.Len()
	if l == 0 {
		return ""
	}
	names := make([]string, 0, l)
	WalkTupleVars(params, func(pv *types.Var) {
		names = append(names, pv.Name())
	})
	if fn.tsig.Variadic() {
		names[l-1] += "..."
	}
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
	orig types.Object
	imp  *importer.Importer
}

func newObjectResolver(obj types.Object, imp *importer.Importer) objectResolver {
	o := objectResolver{obj, imp}
	o.importType(obj.Type())
	return o
}

// PackageAlias returns the alias of the package as declared in the import
// statement.
func (o *objectResolver) PackageAlias() string {
	return o.imp.AliasOf(o.orig.Pkg().Path())
}

// TypeString returns the object type as a string, resolving the package
// names using the aliases declared in the import statements.
func (o *objectResolver) TypeString() string {
	return types.TypeString(o.orig.Type(), o.qualifier())
}

func (o *objectResolver) qualifier() types.Qualifier {
	return func(p *types.Package) string {
		return o.imp.AliasOf(p.Path())
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
		o.importType(typ.Params())
		o.importType(typ.Results())
	case *types.Tuple:
		WalkTupleVars(typ, o.varImporter())
	case *types.Struct:
		WalkStructFields(typ, o.varImporter())
	case *types.Interface:
		WalkInterfaceMethods(typ, o.funcImporter())
		WalkInterfaceEmbeddeds(typ, o.importType)
	}
}

func (o *objectResolver) varImporter() func(pv *types.Var) {
	return func(pv *types.Var) {
		o.imp.AddImport(pv.Pkg())
		o.importType(pv.Type())
	}
}

func (o *objectResolver) funcImporter() func(*types.Func) {
	return func(fn *types.Func) {
		o.imp.AddImport(fn.Pkg())
		o.importType(fn.Type())
	}
}

// Object extends the [types.Object] interface with methods to get the package
// alias and the resolved type as a string.
type Object interface {
	types.Object
	PackageAliaser
	TypeStringer
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

// WalkTupleVars calls the given function for each variable in the given tuple.
func WalkTupleVars(tp *types.Tuple, fn func(*types.Var)) {
	for i := 0; i < tp.Len(); i++ {
		fn(tp.At(i))
	}
}

// WalkStructFields calls the given function for each field in the given struct.
func WalkStructFields(st *types.Struct, fn func(*types.Var)) {
	for i := 0; i < st.NumFields(); i++ {
		fn(st.Field(i))
	}
}

// WalkInterfaceMethods calls the given function for each method in the given
// interface.
func WalkInterfaceMethods(iface *types.Interface, fn func(*types.Func)) {
	for i := 0; i < iface.NumMethods(); i++ {
		fn(iface.Method(i))
	}
}

// WalkInterfaceEmbeddeds calls the given function for each embedded type in the
// given interface.
func WalkInterfaceEmbeddeds(iface *types.Interface, fn func(types.Type)) {
	for i := 0; i < iface.NumEmbeddeds(); i++ {
		fn(iface.EmbeddedType(i))
	}
}
