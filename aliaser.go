package aliaser

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"text/template"

	"github.com/marcozac/go-aliaser/importer"
	"github.com/marcozac/go-aliaser/util/set"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

//go:embed template/*
var tmplFS embed.FS

// Aliaser is the primary type of this package. It is used to generate the
// aliases for the loaded package.
type Aliaser struct {
	*Config
	*importer.Importer

	// constants is the list of exported constants in the loaded package.
	constants []*Const

	// Variables is the list of exported variables in the loaded package.
	variables []*Var

	// Functions is the list of exported functions in the loaded package.
	functions []*Func

	// Types is the list of exported types in the loaded package.
	types []*TypeName

	names *set.Set[string, objectId]
	mu    sync.RWMutex
}

// New returns a new [Aliaser] with the given configuration.
// The configuration is required and must have a valid target package and
// pattern. Otherwise, a [ErrNilConfig], [ErrEmptyTarget] or [ErrEmptyPattern]
// will be returned. See [Config] for more details.
//
// New may also return an error in these cases:
//   - Package loading fails
//   - The package has errors
//   - More or less than one package is loaded with the given pattern
//   - The loaded package has an unexpected object type
//
// Example:
//
//	a, err := aliaser.New(&aliaser.Config{
//		TargetPackage: "foo",
//		Pattern: "github.com/example/package",
//	})
//	if err != nil {
//		// ...
//	}
//	a.GenerateFile("mypkg/alias.go")
//
//	// mypkg/alias.go
//	// Code generated by aliaser. DO NOT EDIT.
//
//	package foo
//
//	import "github.com/marcozac/go-aliaser/internal/testing/pkg"
//
//	const (
//		// ...
//	)
func New(c *Config, opts ...Option) (*Aliaser, error) {
	switch {
	case c == nil:
		return nil, ErrNilConfig
	case c.TargetPackage == "":
		return nil, ErrEmptyTarget
	case c.Pattern == "":
		return nil, ErrEmptyPattern
	}
	a := &Aliaser{
		Config:   c.setDefaults().applyOptions(opts...),
		Importer: importer.New(),
		names:    set.New[string, objectId](),
	}
	if err := a.load(); err != nil {
		return nil, err
	}
	return a, nil
}

const loadMode = packages.NeedName | packages.NeedTypes

func (a *Aliaser) load() error {
	pkgs, err := packages.Load(&packages.Config{Mode: loadMode, Context: a.ctx}, a.Pattern)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}
	if len(pkgs) != 1 {
		return fmt.Errorf("expected one package, got %d", len(pkgs))
	}
	pkg := pkgs[0]
	if errs := pkg.Errors; len(errs) > 0 {
		return fmt.Errorf("package errors: %w", PackagesErrors(errs))
	}
	return a.addPkgObjects(pkg)
}

func (a *Aliaser) addPkgObjects(pkg *packages.Package) error {
	a.AddImport(pkg.Types)
	scope := pkg.Types.Scope()
	for _, name := range pkg.Types.Scope().Names() {
		o := scope.Lookup(name)
		if !o.Exported() {
			continue
		}
		if _, ok := a.excludedNames[o.Name()]; ok {
			continue
		}
		switch o := o.(type) {
		case *types.Const:
			if !a.excludeConstants {
				a.AddConstants(o)
			}
		case *types.Var:
			if !a.excludeVariables {
				a.AddVariables(o)
			}
		case *types.Func:
			if !a.excludeFunctions {
				a.AddFunctions(o)
			}
		case *types.TypeName:
			if !a.excludeTypes {
				a.AddTypes(o)
			}
		default: // should never happen
			return fmt.Errorf("unexpected object type for %s: %T", o.Name(), o)
		}
	}
	return nil
}

// Constants returns the list of the constants loaded for aliasing.
func (a *Aliaser) Constants() []*Const {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.constants
}

// AddConstants adds the given constants to the list of the constants to
// generate aliases for.
//
// NOTE:
// Currently, the Aliaser does not perform any check to avoid adding the same
// constant or any other object with the same name more than once. Adding a
// constant with the same name of another object will result in a non-buildable
// generated code.
func (a *Aliaser) AddConstants(cs ...*types.Const) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, c := range cs {
		a.addConstant(c)
	}
}

func (a *Aliaser) addConstant(c *types.Const) {
	if !a.addObjectName(c, constantId) {
		a.constants = append(a.constants, NewConst(c, a.Importer))
		a.AddImport(c.Pkg())
	}
}

// Variables returns the list of the variables loaded for aliasing.
func (a *Aliaser) Variables() []*Var {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.variables
}

// AddVariables adds the given variables to the list of the variables to
// generate aliases for.
func (a *Aliaser) AddVariables(vs ...*types.Var) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, v := range vs {
		a.addVariable(v)
	}
}

func (a *Aliaser) addVariable(v *types.Var) {
	if !a.addObjectName(v, variableId) {
		a.AddImport(v.Pkg())
		a.variables = append(a.variables, NewVar(v, a.Importer))
	}
}

// Functions returns the list of the functions loaded for aliasing.
func (a *Aliaser) Functions() []*Func {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.functions
}

// AddFunctions adds the given functions to the list of the functions to
// generate aliases for.
func (a *Aliaser) AddFunctions(fns ...*types.Func) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, fn := range fns {
		a.addFunction(fn)
	}
}

func (a *Aliaser) addFunction(fn *types.Func) {
	if !a.addObjectName(fn, functionId) {
		a.AddImport(fn.Pkg())
		a.functions = append(a.functions, NewFunc(fn, a.Importer))
	}
}

// Types returns the list of the types loaded for aliasing.
func (a *Aliaser) Types() []*TypeName {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.types
}

// AddTypes adds the given types to the list of the types to generate aliases
// for.
func (a *Aliaser) AddTypes(ts ...*types.TypeName) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, t := range ts {
		a.addType(t)
	}
}

func (a *Aliaser) addType(t *types.TypeName) {
	if !a.addObjectName(t, typeId) {
		a.AddImport(t.Pkg())
		a.types = append(a.types, NewTypeName(t, a.Importer))
	}
}

// objectId is the type used to identify the kind of object.
type objectId int

const (
	_ objectId = iota
	constantId
	variableId
	functionId
	typeId
)

func (a *Aliaser) addObjectName(o types.Object, id objectId) (skip bool) {
	if a.names.PutNX(o.Name(), id) {
		return
	}
	switch a.onDuplicate {
	case OnDuplicateSkip:
		return true
	case OnDuplicateReplace:
		oldID := a.names.Swap(o.Name(), id)
		a.deleteObject(o, oldID)
	case OnDuplicatePanic:
		panic(fmt.Errorf("duplicate object name: %s", o.Name()))
	default: // should never happen, trap for development
		panic(fmt.Errorf("unexpected OnDuplicate value: %d", a.onDuplicate))
	}
	return
}

func (a *Aliaser) deleteObject(o types.Object, id objectId) {
	switch id {
	case constantId:
		a.constants = slices.DeleteFunc(a.constants, newObjSliceDel[*Const](o))
	case variableId:
		a.variables = slices.DeleteFunc(a.variables, newObjSliceDel[*Var](o))
	case functionId:
		a.functions = slices.DeleteFunc(a.functions, newObjSliceDel[*Func](o))
	case typeId:
		a.types = slices.DeleteFunc(a.types, newObjSliceDel[*TypeName](o))
	default: // should never happen, trap for development
		panic(fmt.Errorf("unexpected object ID: %d", id))
	}
}

// newObjSliceDel returns a function, compatible with the signature of the
// [slices.DeleteFunc], that returns true if the given object (the one that
// will be deleted) has the same name of the one used to create the function.
func newObjSliceDel[O types.Object](obj types.Object) func(O) bool {
	return func(o O) bool {
		return obj.Name() == o.Name()
	}
}

// Generate writes the aliases to the given writer.
//
// Generate returns an error if fails to execute the template, format the
// generated code, or write the result to the writer.
func (a *Aliaser) Generate(wr io.Writer) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.generate(wr)
}

// GenerateFile behaves like [Aliaser.Generate], but it writes the aliases to
// the file with the given name creating the necessary directories. If the file
// already exists, it is truncated.
//
// GenerateFile returns an error in the same cases as [Aliaser.Generate] and
// if any of the directory creation or file writing operations fail. In this
// case, if the file did not exist before the operation, it is removed,
// otherwise, its content is reset to the original state.
func (a *Aliaser) GenerateFile(name string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	f, reset, err := OpenFileWithReset(name)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := a.generate(f); err != nil {
		err = fmt.Errorf("generate: %w", err)
		if rerr := reset(); rerr != nil {
			return errors.Join(err, fmt.Errorf("reset file: %w", rerr))
		}
		return err
	}
	return nil
}

func (a *Aliaser) generate(wr io.Writer) error {
	buf := new(bytes.Buffer)
	if err := a.executeTemplate(buf); err != nil {
		return err
	}
	data, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("format: %w", err)
	}
	if _, err := wr.Write(data); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

func (a *Aliaser) executeTemplate(buf *bytes.Buffer) error {
	tmpl, err := template.ParseFS(tmplFS, "template/*.tmpl")
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if err := tmpl.ExecuteTemplate(buf, "alias", a); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}

// Config is the configuration used to define the target package.
type Config struct {
	config

	// [REQUIRED]
	// TargetPackage is the name of the package where the aliases will be
	// generated. For example, if the package path is
	// "github.com/marcozac/go-aliaser/pkg-that-needs-aliases/foo", the target
	// package is "foo".
	TargetPackage string

	// [REQUIRED]
	// Pattern is the package pattern in Go format to be loaded.
	//
	// Example:
	//
	//	"github.com/marcozac/go-aliaser/pkg-that-will-be-aliased"
	Pattern string

	// Header is an optional header to be written at the top of the file.
	//
	// Default: "// Code generated by aliaser. DO NOT EDIT."
	Header string

	// AssignFunctions sets whether the aliases for the functions should be
	// assigned to a variable instead of being wrapped.
	AssignFunctions bool
}

func (c *Config) setDefaults() *Config {
	c.excludedNames = make(map[string]struct{})
	if c.Header == "" {
		c.Header = "// Code generated by aliaser. DO NOT EDIT."
	}
	if c.ctx == nil {
		c.ctx = context.Background()
	}
	return c
}

func (c *Config) applyOptions(opts ...Option) *Config {
	for _, o := range opts {
		o.set(c)
	}
	return c
}

type config struct {
	ctx              context.Context
	excludeConstants bool
	excludeVariables bool
	excludeFunctions bool
	excludeTypes     bool
	excludedNames    map[string]struct{}
	onDuplicate      int
}

// Option is the interface implemented by all options.
type Option interface{ set(*Config) }

type option func(*Config)

func (o option) set(c *Config) {
	o(c)
}

// WithContext sets the context to be used when loading the package.
func WithContext(ctx context.Context) Option {
	return option(func(c *Config) {
		c.ctx = ctx
	})
}

// WithHeader sets an optional header to be written at the top of the file.
func WithHeader(header string) Option {
	return option(func(c *Config) {
		c.Header = header
	})
}

// ExcludeConstants excludes the constants from the loaded package.
func ExcludeConstants(v bool) Option {
	return option(func(c *Config) {
		c.excludeConstants = v
	})
}

// ExcludeVariables excludes the variables from the loaded package.
func ExcludeVariables(v bool) Option {
	return option(func(c *Config) {
		c.excludeVariables = v
	})
}

// ExcludeFunctions excludes the functions from the loaded package.
func ExcludeFunctions(v bool) Option {
	return option(func(c *Config) {
		c.excludeFunctions = v
	})
}

// ExcludeTypes excludes the types from the loaded package.
func ExcludeTypes(v bool) Option {
	return option(func(c *Config) {
		c.excludeTypes = v
	})
}

// ExcludeNames excludes the given names from the loaded package. Each name
// is valid for all kinds of objects (constants, variables, functions, and
// types).
//
// For example, if the loaded package has a constant named "A" and a type
// named "B", ExcludeNames("A", "B") will exclude both "A" and "B".
func ExcludeNames(names ...string) Option {
	return option(func(c *Config) {
		for _, n := range names {
			c.excludedNames[n] = struct{}{}
		}
	})
}

// AssignFunctions sets whether the aliases for the functions should be
// assigned to a variable instead of being wrapped.
func AssignFunctions(v bool) Option {
	return option(func(c *Config) {
		c.AssignFunctions = v
	})
}

const (
	// OnDuplicateSkip is the default behavior when a duplicate object name is
	// found. It skips the object and does not generate an alias for it.
	OnDuplicateSkip = iota

	// OnDuplicateReplace replaces the old object (even if it is a different
	// kind of object) with the new one.
	OnDuplicateReplace

	// OnDuplicatePanic panics when a duplicate object name is found.
	OnDuplicatePanic
)

// OnDuplicate sets the behavior when a duplicate object name is found.
func OnDuplicate(v int) Option {
	return option(func(c *Config) {
		c.onDuplicate = v
	})
}
