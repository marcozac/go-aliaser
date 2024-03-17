package aliaser

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

//go:embed template/*
var tmplFS embed.FS

// New returns a new [Aliaser] with the given target and pattern.
//
// The target is the name of the package where the aliases will be generated
// and the pattern must be a valid package pattern in Go format.
//
// Example:
//
//	a, err := New("foo", "github.com/marcozac/go-aliaser/internal/testing/pkg")
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
func New(target, pattern string, opts ...Option) (*Aliaser, error) {
	switch "" {
	case target:
		return nil, ErrEmptyTarget
	case pattern:
		return nil, ErrEmptyPattern
	}
	c := defaultConfig(target, pattern)
	for _, o := range opts {
		o.set(c)
	}
	a := &Aliaser{}
	if err := a.load(c); err != nil {
		return nil, err
	}
	return a, nil
}

// Aliaser is the primary type of this package. It is used to generate the
// aliases for the loaded package.
type Aliaser struct {
	alias *Alias
}

// Generate writes the aliases to the given writer.
//
// Generate returns an error if fails to execute the template, format the
// generated code, or write the result to the writer.
func (a *Aliaser) Generate(wr io.Writer) error {
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

// GenerateFile behaves like [Aliaser.Generate], but it writes the aliases to
// the file with the given name. It creates the necessary directories if they
// don't exist. If the file already exists, it is truncated.
//
// GenerateFile returns an error in the same cases as [Aliaser.Generate] and
// if any of the directory creation or file writing operations fail.
func (a *Aliaser) GenerateFile(name string) error {
	// TODO: keep a backup and restore in case of failure
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	return a.Generate(f)
}

func (a *Aliaser) executeTemplate(buf *bytes.Buffer) error {
	tmpl, err := template.ParseFS(tmplFS, "template/*.tmpl")
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	if err := tmpl.ExecuteTemplate(buf, "alias", a.alias); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	return nil
}

const loadMode = packages.NeedName | packages.NeedTypes

func (a *Aliaser) load(c *Config) error {
	pkgs, err := packages.Load(&packages.Config{Mode: loadMode, Context: c.ctx}, c.pattern)
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
	return a.setAlias(c, pkg)
}

func (a *Aliaser) setAlias(c *Config, pkg *packages.Package) error {
	a.alias = &Alias{
		Config:   c,
		Importer: NewImporter(),
	}
	a.alias.AddImport(pkg.Types)
	scope := pkg.Types.Scope()
	for _, name := range pkg.Types.Scope().Names() {
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
				a.alias.AddConstants(o)
			}
		case *types.Var:
			if !c.excludeVariables {
				a.alias.AddVariables(o)
			}
		case *types.Func:
			if !c.excludeFunctions {
				a.alias.AddFunctions(o)
			}
		case *types.TypeName:
			if !c.excludeTypes {
				a.alias.AddTypes(o)
			}
		default: // should never happen
			return fmt.Errorf("unexpected object type for %s: %T", o.Name(), o)
		}
	}
	return nil
}

// Config is the configuration used to define the target package. It is
// embedded in the [Alias] type.
type Config struct {
	config

	// TargetPackage is the name of the package where the aliases will be
	// generated.
	TargetPackage string

	// Header is an optional header to be written at the top of the file.
	Header string
}

type config struct {
	pattern          string // required
	ctx              context.Context
	excludeConstants bool
	excludeVariables bool
	excludeFunctions bool
	excludeTypes     bool
	excludedNames    map[string]struct{}
}

func defaultConfig(target, pattern string) *Config {
	return &Config{
		TargetPackage: target,
		Header:        "// Code generated by aliaser. DO NOT EDIT.",
		config: config{
			pattern:       pattern,
			ctx:           context.Background(),
			excludedNames: make(map[string]struct{}),
		},
	}
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
func ExcludeConstants() Option {
	return option(func(c *Config) {
		c.excludeConstants = true
	})
}

// ExcludeVariables excludes the variables from the loaded package.
func ExcludeVariables() Option {
	return option(func(c *Config) {
		c.excludeVariables = true
	})
}

// ExcludeFunctions excludes the functions from the loaded package.
func ExcludeFunctions() Option {
	return option(func(c *Config) {
		c.excludeFunctions = true
	})
}

// ExcludeTypes excludes the types from the loaded package.
func ExcludeTypes() Option {
	return option(func(c *Config) {
		c.excludeTypes = true
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
