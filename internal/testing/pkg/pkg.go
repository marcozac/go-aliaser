//go:build !codeanalysis

package pkg

import (
	"bytes"
	"context"
	stdjson "encoding/json"
	"strings"
	"text/template"

	"github.com/marcozac/go-aliaser/internal/testing/pkg/json"
	"golang.org/x/tools/go/packages"
)

const (
	A = 1
	a = 1
)

var (
	B = "b"
	b = "b"
)

func C() {}

type (
	D string
	E any
	F = int
	G = D
)

const H G = "h"

var I = H

func J(
	p1 int,
	p2 *string,
	p3 []string,
	p4 []*bool,
	p5 [2]string,
	p6 map[string]int,
	p7 chan int,
	p8 func(int) string,
	p9 any,
	p10 struct {
		Foo string
		Bar context.Context
		Baz struct {
			Builder strings.Builder
		}
	},
	p11 interface {
		L
		Foo() string
		Bar() bytes.Buffer
	},
	p12 D,
	p13 *E,
	p14 []F,
	p15 [2]G,
	p16 map[D]E,
	p17 context.CancelCauseFunc,
	p18 stdjson.Marshaler,
	json json.Foo, // test same name of alias
	variadic ...*packages.Module,
) (int, any, *D, error) {
	return 0, nil, nil, nil
}

type K interface {
	Baz() string
}

type L interface {
	K
}

type M template.Template

type N[T any] struct {
	Foo T
}

type O N[string]

type P[T any, V ~string] struct {
	N N[T]
	V V
}

func (*P[T, V]) Foo() {}

type Q = P[string, D]

type R[T any] P[T, string]

func S[T any](t T) {}

func T[C context.Context, S ~string, T any](ctx C, s S, t T) (S, *P[T, S]) {
	return s, &P[T, S]{}
}

func U[T any]() T {
	var v T
	return v
}

func V[I ~string]() P[int, I] {
	return P[int, I]{}
}

func W() P[int, string] {
	return P[int, string]{}
}
