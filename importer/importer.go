package importer

import (
	"fmt"
	"go/types"
	"slices"
	"sync"
)

// Importer is the type used to manage the package imports.
// It is used to ensure that the same package is imported only once and to
// alias the package names avoiding conflicts on same-name packages.
//
// All exported methods are safe for concurrent use if the underlying package
// is not modified concurrently by other goroutines.
type Importer struct {
	// imports is the map of package imports formatted as "path:Package"
	imports        map[string]*types.Package
	toAlias        bool
	aliasedImports map[string]string
	mu             sync.RWMutex
}

// New returns a new [Importer].
func New() *Importer {
	return &Importer{imports: make(map[string]*types.Package)}
}

// AddImport adds the given package to the list of imports ensuring that the
// same package is imported only once. The comparison is based on the package
// path. If the given package is nil, the function returns without performing
// any operation.
func (imp *Importer) AddImport(p *types.Package) {
	imp.mu.Lock()
	defer imp.mu.Unlock()
	imp.addImport(p)
}

func (imp *Importer) addImport(p *types.Package) {
	if p == nil {
		return
	}
	path := p.Path()
	if _, ok := imp.imports[path]; !ok {
		imp.toAlias = true
		imp.imports[path] = p
	}
}

// Imports returns the list of the imported packages.
func (imp *Importer) Imports() []*types.Package {
	imp.mu.RLock()
	defer imp.mu.RUnlock()
	return imp.importsList()
}

func (imp *Importer) importsList() []*types.Package {
	imports := make([]*types.Package, 0, len(imp.imports))
	for _, pkg := range imp.imports {
		imports = append(imports, pkg)
	}
	return imports
}

// AliasedImports returns the map of package imports formatted as "path:alias".
// All the imports are aliased to avoid conflicts on same-name packages.
// Moreover, the imports are sorted by path before aliasing them, ensuring
// deterministic results and avoiding, for example, false positives in tests
// comparing the generated code.
//
// NOTE:
// The method should be called only once, after all imports have been added, or
// it may produce inconsistent results. For example, if a package (B) is added
// after the first call and has the same name of another one (A) but a path
// that sorts before, A alias will be different from the first result. See the
// example below for more details.
//
// Example:
//
//	a, err := aliaser.New(&aliaser.Config{TargetPackage: "mypkg", Pattern: "github.com/example/package"})
//	if err != nil {
//	  // ...
//	}
//
//	a.AddImport(types.NewPackage("github.com/marcozac/go-aliaser/fake1", "fake"))
//	a.AddImport(types.NewPackage("github.com/marcozac/go-aliaser/fake3", "fake"))
//	imports := a.AliasedImports()
//	// "github.com/marcozac/go-aliaser/fake1": "fake"
//	// "github.com/marcozac/go-aliaser/fake3": "fake_2"
//
//	// Bad!
//	a.AddImport(types.NewPackage("github.com/marcozac/go-aliaser/fake2", "fake"))
//	imports = a.AliasedImports()
//	// "github.com/marcozac/go-aliaser/fake1": "fake"
//	// "github.com/marcozac/go-aliaser/fake2": "fake_2"
//	// "github.com/marcozac/go-aliaser/fake3": "fake_3" // different alias!
func (imp *Importer) AliasedImports() map[string]string {
	imp.mu.Lock()
	defer imp.mu.Unlock()
	return imp.aliasImports()
}

func (imp *Importer) aliasImports() map[string]string {
	if !imp.toAlias {
		return imp.aliasedImports
	}
	l := len(imp.imports)
	paths := make([]string, 0, l)
	for path := range imp.imports {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	orderedImports := make(map[string]string, l)
	for _, path := range paths {
		pkg := imp.imports[path]
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
	}
	imp.aliasedImports = orderedImports // cache
	imp.toAlias = false
	return orderedImports
}

// AliasOf returns the alias of the given package path. If the package is not
// imported, the function returns an empty string.
func (imp *Importer) AliasOf(path string) string {
	imp.mu.Lock()
	defer imp.mu.Unlock()
	return imp.aliasOf(path)
}

func (imp *Importer) aliasOf(path string) string {
	if imp.toAlias {
		imp.aliasImports()
	}
	return imp.aliasedImports[path]
}
