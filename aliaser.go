package aliaser

import (
	"context"
	"errors"
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// Load loads the package at the given path and returns a [Src] with the
// package's exported constants, variables, functions, and types.
//
// The path must be in Go format, e.g. "github.com/marcozac/go-aliaser/internal/testdata".
func Load(from string, opts ...Option) (*Src, error) {
	if from == "" {
		return nil, errors.New("empty package path")
	}
	c := applyOptions(nil, opts...)
	pkgs, err := packages.Load(&packages.Config{
		Mode:    packages.NeedName | packages.NeedTypes,
		Context: c.Context,
	}, from)
	if err != nil {
		return nil, fmt.Errorf("load package: %w", err)
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected one package, got %d", len(pkgs))
	}
	lp := pkgs[0]
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
		switch o := o.(type) {
		case *types.Const:
			p.Constants = append(p.Constants, o)
		case *types.Var:
			p.Variables = append(p.Variables, o)
		case *types.Func:
			p.Functions = append(p.Functions, o)
		case *types.TypeName:
			p.Types = append(p.Types, o)
		default:
			return nil, fmt.Errorf("unexpected object type for %s: %T", o.Name(), o)
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
	Context context.Context
}

func WithContext(ctx context.Context) Option {
	return func(c *config) {
		c.Context = ctx
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
		Context: context.Background(),
	}
}
