package aliaser

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ File = (*fileErrorer)(nil)

type fileErrorer struct {
	fileErrorerConfig
	f *os.File
}

func openFileErrorer(name string) (File, error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	return &fileErrorer{f: f}, nil
}

func newOpenFileErrorer(c fileErrorerConfig) func(name string) (File, error) {
	return func(name string) (File, error) {
		f, err := openFileErrorer(name)
		if err != nil {
			return nil, fmt.Errorf("open: %w", err)
		}
		f.(*fileErrorer).fileErrorerConfig = c
		return f, nil
	}
}

type fileErrorerConfig struct {
	noStatErr  bool
	noReadErr  bool
	noSeekErr  bool
	noWriteErr bool
	noTruncErr bool
}

func (fe *fileErrorer) Stat() (fs.FileInfo, error) {
	if fe.noStatErr {
		return fe.f.Stat()
	}
	return nil, assert.AnError
}

func (fe *fileErrorer) Close() error {
	return fe.f.Close()
}

func (fe *fileErrorer) Read(b []byte) (int, error) {
	if fe.noReadErr {
		return fe.f.Read(b)
	}
	return 0, assert.AnError
}

func (fe *fileErrorer) Seek(int64, int) (int64, error) {
	if fe.noSeekErr {
		return fe.f.Seek(0, 0)
	}
	return 0, assert.AnError
}

func (fe *fileErrorer) Write(b []byte) (int, error) {
	if fe.noWriteErr {
		return fe.f.Write(b)
	}
	return len(b) - 1, assert.AnError
}

func (fe *fileErrorer) Truncate(int64) error {
	if fe.noTruncErr {
		return fe.f.Truncate(0)
	}
	return assert.AnError
}

func TestOpenFileWithReset(t *testing.T) {
	dir := t.TempDir()
	t.Run("NotReadable", func(t *testing.T) {
		privateDir := filepath.Join(dir, "private")
		require.NoError(t, os.Mkdir(privateDir, 0o755))
		filename := filepath.Join(privateDir, "private.go")
		_, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o000)
		require.NoError(t, err)
		require.NoError(t, os.Chmod(privateDir, 0o000)) // revoke all permissions
		_, _, err = OpenFileWithReset(filename)
		assert.Error(t, err)
		require.NoError(t, os.Chmod(privateDir, 0o755)) // restore permissions to clean up
	})
	t.Run("Reset", func(t *testing.T) {
		t.Run("NotExists", func(t *testing.T) {
			filename := filepath.Join(dir, "not-exists.go")
			f, reset, err := OpenFileWithReset(filename)
			require.NoError(t, err)
			require.NoError(t, f.Close())
			assert.FileExists(t, filename)
			assert.NoError(t, reset())
			assert.NoFileExists(t, filename)
		})
		t.Run("Exists", func(t *testing.T) {
			tf, err := os.CreateTemp(dir, "exists.go")
			require.NoError(t, err)
			data := []byte("package foo\n")
			_, err = tf.Write(data)
			require.NoError(t, err)
			require.NoError(t, tf.Close())
			f, reset, err := OpenFileWithReset(tf.Name())
			require.NoError(t, err)
			_, err = f.Write([]byte("func main() {}\n"))
			require.NoError(t, err)
			assert.NoError(t, reset())
			require.NoError(t, f.Close())
			assert.FileExists(t, tf.Name())
			content, err := os.ReadFile(tf.Name())
			require.NoError(t, err)
			assert.Equal(t, data, content)
		})
	})
	t.Run("Error", func(t *testing.T) {
		opener := openFile
		defer func() { openFile = opener }()
		t.Run("BakRead", func(t *testing.T) {
			defer func() { openFile = opener }()
			openFile = openFileErrorer
			tf, err := os.CreateTemp(dir, "bak-read-*.go")
			require.NoError(t, err)
			require.NoError(t, tf.Close())
			f, reset, err := OpenFileWithReset(tf.Name())
			assert.Error(t, err)
			assert.Nil(t, f)
			assert.Nil(t, reset)
		})
		t.Run("Seek", func(t *testing.T) {
			defer func() { openFile = opener }()
			openFile = newOpenFileErrorer(fileErrorerConfig{noReadErr: true})
			tf, err := os.CreateTemp(dir, "seek-*.go")
			require.NoError(t, err)
			require.NoError(t, tf.Close())
			f, reset, err := OpenFileWithReset(tf.Name())
			assert.Error(t, err)
			assert.Nil(t, f)
			assert.Nil(t, reset)
		})
		t.Run("Truncate", func(t *testing.T) {
			defer func() { openFile = opener }()
			openFile = newOpenFileErrorer(fileErrorerConfig{
				noReadErr: true,
				noSeekErr: true,
			})
			tf, err := os.CreateTemp(dir, "truncate-*.go")
			require.NoError(t, err)
			require.NoError(t, tf.Close())
			f, reset, err := OpenFileWithReset(tf.Name())
			assert.Error(t, err)
			assert.Nil(t, f)
			assert.Nil(t, reset)
		})
		t.Run("Restore", func(t *testing.T) {
			defer func() { openFile = opener }()
			openFile = newOpenFileErrorer(fileErrorerConfig{
				noReadErr:  true,
				noSeekErr:  true,
				noTruncErr: true,
			})
			tf, err := os.CreateTemp(dir, "restore-*.go")
			require.NoError(t, err)
			_, err = tf.WriteString("package foo\n")
			require.NoError(t, err)
			require.NoError(t, tf.Close())
			f, reset, err := OpenFileWithReset(tf.Name())
			require.NoError(t, err)
			defer f.Close()
			require.NotNil(t, f)
			require.NotNil(t, reset)
			f.(*fileErrorer).fileErrorerConfig.noSeekErr = false
			assert.Error(t, reset(), "seek error")
			f.(*fileErrorer).fileErrorerConfig.noSeekErr = true
			f.(*fileErrorer).fileErrorerConfig.noTruncErr = false
			assert.Error(t, reset(), "truncate error")
			f.(*fileErrorer).fileErrorerConfig.noTruncErr = true
			assert.Error(t, reset(), "bak write error")
		})
	})
}
