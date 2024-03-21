package aliaser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

// File is an interface that extends [fs.File] with the [io.WriteSeeker] and
// Truncate methods.
type File interface {
	fs.File
	io.WriteSeeker
	Truncate(size int64) error
}

// OpenFileWithReset opens a file for reading and writing, creating it if it
// does not exist. It also returns a function to reset the file to its original
// state. If the file does not exist, the reset function will remove it,
// otherwise, it is truncate and the reset function will rewrite the original
// content. The reset function does not close the file, so it is the caller's
// responsibility to do so.
//
// NOTE: This function does not create the parent directories of the file.
func OpenFileWithReset(name string) (File, func() error, error) {
	exists, err := fileExists(name)
	if err != nil {
		return nil, nil, fmt.Errorf("exists: %w", err)
	}
	f, err := openFile(name)
	if err != nil {
		return nil, nil, fmt.Errorf("open: %w", err)
	}
	if !exists {
		return f, func() error { return os.Remove(name) }, nil
	}
	var bak bytes.Buffer
	if _, err := bak.ReadFrom(f); err != nil {
		return nil, nil, fmt.Errorf("read: %w", err)
	}
	if err := seekAndTruncate(f); err != nil {
		return nil, nil, err
	}
	return f, func() error {
		if err := seekAndTruncate(f); err != nil {
			return fmt.Errorf("restore: %w", err)
		}
		if _, err := bak.WriteTo(f); err != nil {
			return fmt.Errorf("restore: %w", err)
		}
		return nil
	}, nil
}

// openFile is a helper function to open a file for reading and writing, creating
// it if it does not exist. It can be mocked in tests.
var openFile = func(name string) (File, error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	return f, nil
}

// seekAndTruncate is a helper function to seek to the beginning of a file and
// truncate it.
func seekAndTruncate(f File) error {
	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("seek file: %w", err)
	}
	if err := f.Truncate(0); err != nil {
		return fmt.Errorf("truncate file: %w", err)
	}
	return nil
}

// fileExists is a helper function to check if a file exists by calling
// [os.Lstat].
func fileExists(name string) (bool, error) {
	if _, err := os.Lstat(name); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
