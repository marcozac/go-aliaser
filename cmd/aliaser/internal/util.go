package internal

// Must panics if err is not nil.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustV accepts a value and an error. If the error is not nil, it panics.
// Otherwise, it returns the value.
func MustV[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
