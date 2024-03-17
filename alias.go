package aliaser

import (
	"fmt"
	"go/types"
	"slices"
	"sync"
)

// Alias is the type used as the data for the template execution. It contains
// the configuration used to define the target package and the list of
// exported constants, variables, functions, and types in the loaded package.
type Alias struct {
	*Config

	// Constants is the list of exported constants in the loaded package.
	Constants []*types.Const

	// Variables is the list of exported variables in the loaded package.
	Variables []*types.Var

	// Functions is the list of exported functions in the loaded package.
	Functions []*types.Func

	// Types is the list of exported types in the loaded package.
	Types []*types.TypeName

	// imports is the map of package imports formatted as "path:Package"
	imports map[string]*types.Package

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
	a.Constants = append(a.Constants, c)
	a.addImport(c.Pkg())
}

// AddVariables adds the given variables to the list of the variables to
// generate aliases for.
func (a *Alias) AddVariables(cs ...*types.Var) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, c := range cs {
		a.addVariable(c)
	}
}

func (a *Alias) addVariable(c *types.Var) {
	a.Variables = append(a.Variables, c)
	a.addImport(c.Pkg())
}

// AddFunctions adds the given functions to the list of the functions to
// generate aliases for.
func (a *Alias) AddFunctions(cs ...*types.Func) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, c := range cs {
		a.addFunction(c)
	}
}

func (a *Alias) addFunction(c *types.Func) {
	a.Functions = append(a.Functions, c)
	a.addImport(c.Pkg())
}

// AddTypes adds the given types to the list of the types to generate aliases
// for.
func (a *Alias) AddTypes(cs ...*types.TypeName) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, c := range cs {
		a.addType(c)
	}
}

func (a *Alias) addType(c *types.TypeName) {
	a.Types = append(a.Types, c)
	a.addImport(c.Pkg())
}

// AddImport adds the given package to the list of imports.
// If the package is nil, it does nothing.
func (a *Alias) AddImport(p *types.Package) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.addImport(p)
}

func (a *Alias) addImport(p *types.Package) {
	if p == nil {
		return
	}
	path := p.Path()
	if _, ok := a.imports[path]; !ok {
		a.imports[path] = p
	}
}

// Imports returns the map of package imports formatted as "path:alias".
// All the imports are aliased to avoid conflicts on same-name packages.
// Moreover, the imports are sorted by path before aliasing them ensuring
// deterministic results and avoiding, for example, false positives in tests
// comparing the generated code.
//
// NOTE:
// Calling this method will also modify the package names of the imported
// packages. [types.Package.Name] will then return the alias instead of the
// original name. For this reason and because the aliased name (and the package
// name itself) would be different if, between two calls to this method, is
// added another package with the same name but a path that sorts before, this
// method should be called only once, after all imports have been added.
//
// Example:
//
//	alias := &aliaser.Alias{Config: &aliaser.Config{
//		TargetPackage: "foo",
//		Header:        "Code generated by aliaser. DO NOT EDIT.",
//	}}
//
//	alias.AddImport(types.NewPackage("github.com/marcozac/go-aliaser/fake1", "fake"))
//	alias.AddImport(types.NewPackage("github.com/marcozac/go-aliaser/fake3", "fake"))
//	imports := alias.Imports()
//	// "github.com/marcozac/go-aliaser/fake1" => "fake"
//	// "github.com/marcozac/go-aliaser/fake3" => "fake_2"
//
//	// Bad!
//	alias.AddImport(types.NewPackage("github.com/marcozac/go-aliaser/fake2", "fake"))
//	imports = alias.Imports()
//	// "github.com/marcozac/go-aliaser/fake1" => "fake"
//	// "github.com/marcozac/go-aliaser/fake2" => "fake_2"
//	// "github.com/marcozac/go-aliaser/fake3" => "fake_3"
func (a *Alias) Imports() map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.aliasImports()
}

func (a *Alias) aliasImports() map[string]string {
	l := len(a.imports)
	paths := make([]string, 0, l)
	for path := range a.imports {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	orderedImports := make(map[string]string, l)
	for _, path := range paths {
		pkg := a.imports[path]
		name := pkg.Name()
		alias := name
		i := 2
	aliasLoop:
		for _, aliased := range orderedImports {
			if alias == aliased {
				alias = fmt.Sprintf("%s_%d", name, i)
				i++
				goto aliasLoop
			}
		}
		orderedImports[path] = alias
		pkg.SetName(alias)
	}
	return orderedImports
}
