// Code generated by aliaser. DO NOT EDIT.

//go:build testout

package out

import (
	bytes "bytes"
	context "context"
	json "encoding/json"
	strings "strings"

	pkg "github.com/marcozac/go-aliaser/internal/testing/pkg"
	json_2 "github.com/marcozac/go-aliaser/internal/testing/pkg/json"
	packages "golang.org/x/tools/go/packages"
)

const (
	A = pkg.A
	H = pkg.H
)

var (
	B = pkg.B
	I = pkg.I
)

func C() {
	pkg.C()
}

func J(p1 int, p2 *string, p3 []string, p4 []*bool, p5 [2]string, p6 map[string]int, p7 chan int, p8 func(int) string, p9 any, p10 struct {
	Foo string
	Bar context.Context
	Baz struct{ Builder strings.Builder }
}, p11 interface {
	Bar() bytes.Buffer
	Foo() string
	pkg.L
}, p12 pkg.D, p13 *pkg.E, p14 []int, p15 [2]pkg.D, p16 map[pkg.D]pkg.E, p17 context.CancelCauseFunc, p18 json.Marshaler, json_ json_2.Foo, variadic ...*packages.Module) (int, any, *pkg.D, error) {
	return pkg.J(p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12, p13, p14, p15, p16, p17, p18, json_, variadic...)
}

type (
	D = pkg.D
	E = pkg.E
	F = pkg.F
	G = pkg.G
	K = pkg.K
	L = pkg.L
	M = pkg.M
)
