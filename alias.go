package aliaser

import (
	"go/types"
	"sync"

	"github.com/marcozac/go-aliaser/importer"
)

// Alias is the type used as the data for the template execution. It contains
// the configuration used to define the target package and the list of
// exported constants, variables, functions, and types in the loaded package.
type Alias struct {
	*Config
	*importer.Importer

	// Constants is the list of exported constants in the loaded package.
	Constants []*Const

	// Variables is the list of exported variables in the loaded package.
	Variables []*Var

	// Functions is the list of exported functions in the loaded package.
	Functions []*Func

	// Types is the list of exported types in the loaded package.
	Types []*TypeName

	mu sync.RWMutex
}

// AddConstants adds the given constants to the list of the constants to
// generate aliases for.
func (a *Alias) AddConstants(cs ...*types.Const) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, c := range cs {
		a.addConstant(c)
	}
}

func (a *Alias) addConstant(c *types.Const) {
	a.AddImport(c.Pkg())
	a.Constants = append(a.Constants, NewConst(c, a.Importer))
}

// AddVariables adds the given variables to the list of the variables to
// generate aliases for.
func (a *Alias) AddVariables(vs ...*types.Var) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, v := range vs {
		a.addVariable(v)
	}
}

func (a *Alias) addVariable(v *types.Var) {
	a.AddImport(v.Pkg())
	a.Variables = append(a.Variables, NewVar(v, a.Importer))
}

// AddFunctions adds the given functions to the list of the functions to
// generate aliases for.
func (a *Alias) AddFunctions(fns ...*types.Func) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, fn := range fns {
		a.addFunction(fn)
	}
}

func (a *Alias) addFunction(fn *types.Func) {
	a.AddImport(fn.Pkg())
	a.Functions = append(a.Functions, NewFunc(fn, a.Importer))
}

// AddTypes adds the given types to the list of the types to generate aliases
// for.
func (a *Alias) AddTypes(ts ...*types.TypeName) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, t := range ts {
		a.addType(t)
	}
}

func (a *Alias) addType(t *types.TypeName) {
	a.AddImport(t.Pkg())
	a.Types = append(a.Types, NewTypeName(t, a.Importer))
}
