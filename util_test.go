package aliaser

import "github.com/stretchr/testify/assert"

const (
	// TestPattern is a valid pattern for testing.
	TestPattern = "github.com/marcozac/go-aliaser/internal/testing/pkg"

	// testTarget is the expected target for testing.
	TestTarget = "out"
)

// WriterE is a writer that always returns an error.
type WriterE struct{}

func (WriterE) Write(p []byte) (n int, err error) {
	return 0, assert.AnError
}
