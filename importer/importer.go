package importer

import (
	"fmt"
	"go/types"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/marcozac/go-aliaser/util/maps"
)

// Importer is the type used to manage the package imports.
// It is used to ensure that the same package is imported only once and to
// alias the package names avoiding conflicts on same-name packages.
//
// All exported methods are safe for concurrent use if the underlying packages
// are not modified concurrently by other goroutines.
type Importer struct {
	// imports is the map of package imports formatted as "path:Package"
	imports        map[string]*types.Package
	toAlias        atomic.Bool
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
		imp.toAlias.Store(true)
		imp.imports[path] = p
	}
}

// Imports creates a new slice containing all the imported packages.
func (imp *Importer) Imports() []*types.Package {
	imp.mu.RLock()
	defer imp.mu.RUnlock()
	return imp.importsList()
}

func (imp *Importer) importsList() []*types.Package {
	return maps.Values(imp.imports)
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
	if imp.toAlias.Load() {
		imp.aliasImports()
	}
	return imp.aliasedImports
}

func (imp *Importer) aliasImports() {
	paths := maps.Keys(imp.imports)
	slices.Sort(paths)
	aliases := make([]string, 0, len(paths))
	imp.aliasedImports = make(map[string]string, len(paths))
	for _, path := range paths {
		name := imp.imports[path].Name()
		alias := name
		for i := 2; ; i++ {
			if !slices.Contains(aliases, alias) {
				aliases = append(aliases, alias)
				imp.aliasedImports[path] = alias
				break
			}
			alias = fmt.Sprintf("%s_%d", name, i)
		}
	}
}

// AliasOf returns the alias of the given package path. If the package is not
// imported, the function returns an empty string.
func (imp *Importer) AliasOf(p *types.Package) string {
	imp.mu.Lock()
	defer imp.mu.Unlock()
	if imp.toAlias.Load() {
		imp.aliasImports()
	}
	return imp.aliasOf(p)
}

func (imp *Importer) aliasOf(p *types.Package) string {
	return imp.aliasedImports[p.Path()]
}
