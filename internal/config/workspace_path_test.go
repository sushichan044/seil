package config_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/config"
)

func TestNewWorkspacePath(t *testing.T) {
	t.Run("absolute filePath stays absolute", func(t *testing.T) {
		root := t.TempDir()
		abs := filepath.Join(root, "sub", "file.go")
		p, err := config.NewWorkspacePath(root, abs)
		require.NoError(t, err)
		assert.Equal(t, abs, p.Abs())
	})

	t.Run("relative filePath is interpreted as relative to root", func(t *testing.T) {
		root := t.TempDir()
		p, err := config.NewWorkspacePath(root, "main.go")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(root, "main.go"), p.Abs())
	})

	t.Run("Rel returns path relative to root", func(t *testing.T) {
		root := t.TempDir()
		abs := filepath.Join(root, "src", "foo.go")
		p, err := config.NewWorkspacePath(root, abs)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("src", "foo.go"), p.Rel())
	})

	t.Run("Rel for root-relative input equals the input", func(t *testing.T) {
		root := t.TempDir()
		p, err := config.NewWorkspacePath(root, "main.go")
		require.NoError(t, err)
		assert.Equal(t, "main.go", p.Rel())
	})

	t.Run("returns error when path escapes workspace root", func(t *testing.T) {
		root := t.TempDir()
		_, err := config.NewWorkspacePath(root, "../outside.go")
		assert.Error(t, err)
	})
}
