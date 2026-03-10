package git_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/himo/internal/git"
)

func TestFindRepoRootFrom(t *testing.T) {
	t.Run("finds git root from nested directory", func(t *testing.T) {
		repoRoot := t.TempDir()
		mustMkdirAll(t, filepath.Join(repoRoot, ".git"))

		startDir := filepath.Join(repoRoot, "nested", "deeper")
		mustMkdirAll(t, startDir)

		got, err := git.FindRepoRootFrom(startDir)
		require.NoError(t, err)
		assert.Equal(t, repoRoot, got)
	})

	t.Run("finds git root from file path", func(t *testing.T) {
		repoRoot := t.TempDir()
		mustMkdirAll(t, filepath.Join(repoRoot, ".git"))

		sourceFile := filepath.Join(repoRoot, "nested", "file.go")
		mustMkdirAll(t, filepath.Dir(sourceFile))
		mustWriteFile(t, sourceFile, []byte("package nested\n"))

		got, err := git.FindRepoRootFrom(sourceFile)
		require.NoError(t, err)
		assert.Equal(t, repoRoot, got)
	})

	t.Run("returns not found outside git repo", func(t *testing.T) {
		startDir := t.TempDir()

		_, err := git.FindRepoRootFrom(startDir)
		require.ErrorIs(t, err, git.ErrNotInGitRepo)
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
