package pathutils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/pathutils"
)

func TestDetermineDirectory(t *testing.T) {
	t.Run("returns directory as is", func(t *testing.T) {
		startDir := filepath.Join(t.TempDir(), "nested")
		mustMkdirAll(t, startDir)

		got, err := pathutils.DetermineDirectory(startDir)
		require.NoError(t, err)
		assert.Equal(t, startDir, got)
	})

	t.Run("returns parent directory for file path", func(t *testing.T) {
		startDir := t.TempDir()
		filePath := filepath.Join(startDir, "nested", "file.go")
		mustMkdirAll(t, filepath.Dir(filePath))
		mustWriteFile(t, filePath, []byte("package nested\n"))

		got, err := pathutils.DetermineDirectory(filePath)
		require.NoError(t, err)
		assert.Equal(t, filepath.Dir(filePath), got)
	})
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(path, 0o700))
}

func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, data, 0o600))
}
