package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindUpConfig(t *testing.T) {
	t.Run("finds himo.yml in ancestor directory within git repo", func(t *testing.T) {
		repoRoot := t.TempDir()
		mustMkdirAll(t, filepath.Join(repoRoot, ".git"))

		configPath := filepath.Join(repoRoot, "nested", defaultConfigFileName)
		mustMkdirAll(t, filepath.Dir(configPath))
		mustWriteFile(t, configPath, []byte("post_edit:\n  hooks: []\n"))

		startDir := filepath.Join(repoRoot, "nested", "deeper")
		mustMkdirAll(t, startDir)

		got, err := FindUpAndLoad(startDir)
		require.NoError(t, err)

		assert.Equal(t, configPath, got.Path)
		assert.Equal(t, filepath.Dir(configPath), got.CWD)
	})

	t.Run("returns not found when config does not exist within git repo", func(t *testing.T) {
		parent := t.TempDir()
		repoRoot := filepath.Join(parent, "repo")
		startDir := filepath.Join(repoRoot, "nested")

		mustMkdirAll(t, filepath.Join(repoRoot, ".git"))
		mustMkdirAll(t, startDir)
		mustWriteFile(t, filepath.Join(parent, defaultConfigFileName), []byte("post_edit:\n  hooks: []\n"))

		_, err := FindUpAndLoad(startDir)
		assert.Error(t, err)
	})

	t.Run("starts from parent when given a file path", func(t *testing.T) {
		repoRoot := t.TempDir()
		mustMkdirAll(t, filepath.Join(repoRoot, ".git"))

		configPath := filepath.Join(repoRoot, defaultConfigFileName)
		mustWriteFile(t, configPath, []byte("post_edit:\n  hooks: []\n"))

		sourceFilePath := filepath.Join(repoRoot, "nested", "file.go")
		mustMkdirAll(t, filepath.Dir(sourceFilePath))
		mustWriteFile(t, sourceFilePath, []byte("package nested\n"))

		got, err := FindUpAndLoad(sourceFilePath)
		require.NoError(t, err)
		assert.Equal(t, configPath, got.Path)
	})

	t.Run("returns not found when outside git repo", func(t *testing.T) {
		startDir := t.TempDir()
		mustWriteFile(t, filepath.Join(startDir, defaultConfigFileName), []byte("post_edit:\n  hooks: []\n"))

		_, err := FindUpAndLoad(startDir)
		assert.Error(t, err)
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("uses explicit config path and sets cwd from it", func(t *testing.T) {
		configDir := filepath.Join(t.TempDir(), "config")
		configPath := filepath.Join(configDir, defaultConfigFileName)
		mustMkdirAll(t, configDir)
		mustWriteFile(t, configPath, []byte("post_edit:\n  hooks:\n    - glob: \"**/*.go\"\n      command: go fmt ./...\n"))

		got, err := Load(configPath)
		require.NoError(t, err)

		assert.Equal(t, configPath, got.Path)
		assert.Equal(t, configDir, got.CWD)
		assert.Len(t, got.Config.PostEdit.Hooks, 1)
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
