package aliaser

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"go/types"
	"io"
	"os"
	"text/template"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

//go:embed template/*
var tmplFS embed.FS

// Generate generates the aliases for the given [Alias.Src].
// By default, it writes the aliases to the file at [Alias.Out], but it is
// possible to override this behavior by providing a custom writer using the
// [WithWriter] option. In this case, the file at [Alias.Out] will be ignored.
func Generate(a *Alias, opts ...Option) error {
	c := applyOptions(nil, opts...)
	var buf bytes.Buffer
	if err := generate(a, &buf); err != nil {
		return err
	}
	data, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("format file: %w", err)
	}
	if c.writer == nil {
		f, err := os.OpenFile(a.Out, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer f.Close()
		c.writer = f
	}
	if _, err := c.writer.Write(data); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func generate(a *Alias, w io.Writer) error {
	tmpl, err := template.ParseFS(tmplFS, "template/*.tmpl")
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	if err := tmpl.ExecuteTemplate(w, "alias", a); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	return nil
}

type Alias struct {
	// PkgName is the name of the package where the aliases will be
	// generated.
	PkgName string

	// Out is the file path where the aliases will be written.
	Out string

	// Src is the loaded package to generate aliases for.
	Src *Src

	// Header is an optional header to be written at the top of the file.
	Header string
}

// Load loads the package at the given path and returns a [Src] with the
// package's exported constants, variables, functions, and types.
//
// The path must be in Go format, e.g. "github.com/marcozac/go-aliaser/internal/testdata".
func Load(from string, opts ...Option) (*Src, error) {
	c := applyOptions(nil, opts...)
	return load(c, from)
}

func load(c *config, from string) (*Src, error) {
	if from == "" {
		return nil, ErrEmptyPath
	}
	pkgs, err := packages.Load(&packages.Config{
		Mode:    packages.NeedName | packages.NeedTypes,
		Context: c.ctx,
	}, from)
	if err != nil {
		return nil, fmt.Errorf("load package: %w", err)
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected one package, got %d", len(pkgs))
	}
	lp := pkgs[0]
	if errs := lp.Errors; len(errs) > 0 {
		return nil, fmt.Errorf("load package errors: %w", PackagesErrors(errs))
	}
	p := &Src{
		PkgName: lp.Name,
		PkgPath: lp.PkgPath,
	}
	scope := lp.Types.Scope()
	for _, name := range lp.Types.Scope().Names() {
		o := scope.Lookup(name)
		if !o.Exported() {
			continue
		}
		if _, ok := c.excludedNames[o.Name()]; ok {
			continue
		}
		switch o := o.(type) {
		case *types.Const:
			if !c.excludeConstants {
				p.Constants = append(p.Constants, o)
			}
		case *types.Var:
			if !c.excludeVariables {
				p.Variables = append(p.Variables, o)
			}
		case *types.Func:
			if !c.excludeFunctions {
				p.Functions = append(p.Functions, o)
			}
		case *types.TypeName:
			if !c.excludeTypes {
				p.Types = append(p.Types, o)
			}
		default: // should never happen
			// return nil, fmt.Errorf("unexpected object type for %s: %T", o.Name(), o)
		}
	}
	return p, nil
}

// Src represents a loaded package.
type Src struct {
	// PkgName is the name of the loaded package.
	//
	// Example: "testdata"
	PkgName string

	// PkgPath is the loaded package path in Go format.
	//
	// Example: "github.com/marcozac/go-aliaser/internal/testdata"
	PkgPath string

	// Constants is the list of exported constants in the loaded package.
	Constants []*types.Const

	// Variables is the list of exported variables in the loaded package.
	Variables []*types.Var

	// Functions is the list of exported functions in the loaded package.
	Functions []*types.Func

	// Types is the list of exported types in the loaded package.
	Types []*types.TypeName
}

type Option func(*config)

type config struct {
	ctx              context.Context
	writer           io.Writer
	excludeConstants bool
	excludeVariables bool
	excludeFunctions bool
	excludeTypes     bool
	excludedNames    map[string]struct{}
}

// WithContext sets the context to be used when loading the package.
func WithContext(ctx context.Context) Option {
	return func(c *config) {
		c.ctx = ctx
	}
}

// WithWriter sets the writer to be used when generating the aliases.
func WithWriter(w io.Writer) Option {
	return func(c *config) {
		c.writer = w
	}
}

// ExcludeConstants excludes the constants from the loaded package.
func ExcludeConstants() Option {
	return func(c *config) {
		c.excludeConstants = true
	}
}

// ExcludeVariables excludes the variables from the loaded package.
func ExcludeVariables() Option {
	return func(c *config) {
		c.excludeVariables = true
	}
}

// ExcludeFunctions excludes the functions from the loaded package.
func ExcludeFunctions() Option {
	return func(c *config) {
		c.excludeFunctions = true
	}
}

// ExcludeTypes excludes the types from the loaded package.
func ExcludeTypes() Option {
	return func(c *config) {
		c.excludeTypes = true
	}
}

// ExcludeNames excludes the given names from the loaded package. Each name
// is valid for all kinds of objects (constants, variables, functions, and
// types).
//
// For example, if the loaded package has a constant named "A" and a type
// named "B", ExcludeNames("A", "B") will exclude both "A" and "B".
func ExcludeNames(names ...string) Option {
	return func(c *config) {
		for _, n := range names {
			c.excludedNames[n] = struct{}{}
		}
	}
}

func applyOptions(c *config, opts ...Option) *config {
	if c == nil {
		c = defaultconfig()
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func defaultconfig() *config {
	return &config{
		ctx:           context.Background(),
		excludedNames: make(map[string]struct{}),
	}
}
