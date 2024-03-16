package aliaser

import (
	"fmt"
	"io"
)

// Generate loads the package defined by the given pattern, generates the
// aliases for the target package name, and writes the result to the given
// writer.
//
// Under the hood, Generate creates a new [Aliaser] with the given parameters
// and calls [Aliaser.Generate] with the writer.
func Generate(target, pattern string, wr io.Writer, opts ...Option) error {
	a, err := New(target, pattern, opts...)
	if err != nil {
		return fmt.Errorf("aliaser: %w", err)
	}
	return a.Generate(wr)
}

// GenerateFile behaves like [Generate], but it writes the aliases to the file
// with the given name. It creates the necessary directories if they don't
// exist.
func GenerateFile(target, pattern, name string, opts ...Option) error {
	a, err := New(target, pattern, opts...)
	if err != nil {
		return fmt.Errorf("aliaser: %w", err)
	}
	return a.GenerateFile(name)
}
