package testdata

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
