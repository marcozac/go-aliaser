//go:build !codeanalysis

package pkg

import "context"

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

func J(ctx context.Context, d *D, v any) (context.Context, any, *D, error) {
	select {
	case <-ctx.Done():
		return nil, nil, nil, ctx.Err()
	default:
		return ctx, v, d, nil
	}
}
