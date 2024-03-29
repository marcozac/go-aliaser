package aliaser

import (
	"errors"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	// ErrNilConfig is returned when the given config is nil.
	ErrNilConfig = errors.New("nil config")

	// ErrEmptyTarget is returned when the given target is empty.
	ErrEmptyTarget = errors.New("empty target")

	// ErrEmptyPattern is returned when the given pattern is empty.
	ErrEmptyPattern = errors.New("empty pattern")
)

// PackagesErrors is a slice of [packages.Error] as returned by
// [packages.Load] that implements the error interface.
//
// The error message is a concatenation of the messages of the underlying
// errors, separated by a semicolon and a space.
type PackagesErrors []packages.Error

func (e PackagesErrors) Error() string {
	msgs := make([]string, 0, len(e))
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}
