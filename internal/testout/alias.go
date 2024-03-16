//go:build testout

package testout

import "github.com/marcozac/go-aliaser/internal/testdata"

const (
	A = testdata.A
	H = testdata.H
)

var (
	B = testdata.B
	I = testdata.I
)

// Functions
var (
	C = testdata.C
)

type (
	D = testdata.D
	E = testdata.E
	F = testdata.F
	G = testdata.G
)
