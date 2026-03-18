package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/config"
)

// resolveSymlinks resolves symlinks on a path (e.g. macOS /var → /private/var).
func resolveSymlinks(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	require.NoError(t, err)
	return resolved
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, nil, 0o644))
}

func TestNewWorkspacePath(t *testing.T) {
	t.Run("absolute filePath stays absolute", func(t *testing.T) {
		root := t.TempDir()
		abs := filepath.Join(root, "sub", "file.go")
		mustWriteFile(t, abs)

		p, err := config.NewWorkspacePath(root, abs)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(resolveSymlinks(t, root), "sub", "file.go"), p.Abs())
	})

	t.Run("relative filePath is interpreted as relative to root", func(t *testing.T) {
		root := t.TempDir()
		mustWriteFile(t, filepath.Join(root, "main.go"))

		p, err := config.NewWorkspacePath(root, "main.go")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(resolveSymlinks(t, root), "main.go"), p.Abs())
	})

	t.Run("Rel returns path relative to root", func(t *testing.T) {
		root := t.TempDir()
		abs := filepath.Join(root, "src", "foo.go")
		mustWriteFile(t, abs)

		p, err := config.NewWorkspacePath(root, abs)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("src", "foo.go"), p.Rel())
	})

	t.Run("Rel for root-relative input equals the input", func(t *testing.T) {
		root := t.TempDir()
		mustWriteFile(t, filepath.Join(root, "main.go"))

		p, err := config.NewWorkspacePath(root, "main.go")
		require.NoError(t, err)
		assert.Equal(t, "main.go", p.Rel())
	})

	t.Run("returns error when path escapes workspace root", func(t *testing.T) {
		root := t.TempDir()
		_, err := config.NewWorkspacePath(root, "../outside.go")
		assert.Error(t, err)
	})

	t.Run("symlinked root: file inside workspace is accepted", func(t *testing.T) {
		rawDir := t.TempDir()
		symlinkDir := filepath.Join(t.TempDir(), "symlink-root")
		require.NoError(t, os.Symlink(rawDir, symlinkDir))

		filePath := filepath.Join(rawDir, "file.go")
		mustWriteFile(t, filePath)

		p, err := config.NewWorkspacePath(symlinkDir, filePath)
		require.NoError(t, err)
		assert.Equal(t, resolveSymlinks(t, filePath), p.Abs())
		assert.Equal(t, "file.go", p.Rel())
	})

	t.Run("symlinked filePath: file inside workspace via symlink is accepted", func(t *testing.T) {
		root := t.TempDir()
		realFile := filepath.Join(root, "file.go")
		mustWriteFile(t, realFile)

		symlinkFile := filepath.Join(t.TempDir(), "symlink-file.go")
		require.NoError(t, os.Symlink(realFile, symlinkFile))

		p, err := config.NewWorkspacePath(root, symlinkFile)
		require.NoError(t, err)
		assert.Equal(t, resolveSymlinks(t, realFile), p.Abs())
	})
}
