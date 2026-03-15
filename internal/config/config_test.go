package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/git"
)

func TestResolveConfigFilePathFrom(t *testing.T) {
	t.Run("finds himo.yml in ancestor directory within git repo", func(t *testing.T) {
		repoRoot := t.TempDir()
		mustMkdirAll(t, filepath.Join(repoRoot, ".git"))

		configPath := filepath.Join(repoRoot, "nested", "himo.yml")
		mustMkdirAll(t, filepath.Dir(configPath))
		mustWriteFile(t, configPath, []byte("post_edit:\n  hooks: {}\n"))

		startDir := filepath.Join(repoRoot, "nested", "deeper")
		mustMkdirAll(t, startDir)

		got, err := config.ResolveConfigFilePathFrom(startDir)
		require.NoError(t, err)
		assert.Equal(t, configPath, got)
	})

	t.Run("returns not found when config does not exist within git repo", func(t *testing.T) {
		parent := t.TempDir()
		repoRoot := filepath.Join(parent, "repo")
		startDir := filepath.Join(repoRoot, "nested")

		mustMkdirAll(t, filepath.Join(repoRoot, ".git"))
		mustMkdirAll(t, startDir)
		mustWriteFile(t, filepath.Join(parent, "himo.yml"), []byte("post_edit:\n  hooks: {}\n"))

		got, err := config.ResolveConfigFilePathFrom(startDir)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("starts from parent when given a file path", func(t *testing.T) {
		repoRoot := t.TempDir()
		mustMkdirAll(t, filepath.Join(repoRoot, ".git"))

		configPath := filepath.Join(repoRoot, "himo.yml")
		mustWriteFile(t, configPath, []byte("post_edit:\n  hooks: {}\n"))

		sourceFilePath := filepath.Join(repoRoot, "nested", "file.go")
		mustMkdirAll(t, filepath.Dir(sourceFilePath))
		mustWriteFile(t, sourceFilePath, []byte("package nested\n"))

		got, err := config.ResolveConfigFilePathFrom(sourceFilePath)
		require.NoError(t, err)
		assert.Equal(t, configPath, got)
	})

	t.Run("returns not found when outside git repo", func(t *testing.T) {
		startDir := t.TempDir()
		mustWriteFile(t, filepath.Join(startDir, "himo.yml"), []byte("post_edit:\n  hooks: {}\n"))

		_, err := config.ResolveConfigFilePathFrom(startDir)
		assert.ErrorIs(t, err, git.ErrNotInGitRepo)
	})
}

func TestLoad(t *testing.T) {
	t.Run("uses explicit config path and sets cwd from it", func(t *testing.T) {
		configDir := filepath.Join(t.TempDir(), "config")
		configPath := filepath.Join(configDir, "himo.yml")
		mustMkdirAll(t, configDir)
		mustWriteFile(
			t,
			configPath,
			[]byte("post_edit:\n  hooks:\n    fmt:\n      glob: \"**/*.go\"\n      command: \"go fmt ./...\"\n"),
		)

		got, err := config.Load(configPath)
		require.NoError(t, err)

		assert.Equal(t, configPath, got.Path)
		assert.Equal(t, configDir, got.CWD)
		assert.Len(t, got.Config.PostEdit.Hooks, 1)
		assert.Equal(t, "**/*.go", got.Config.PostEdit.Hooks["fmt"].Glob)
	})

	t.Run("loads setup and teardown hooks", func(t *testing.T) {
		configDir := t.TempDir()
		configPath := filepath.Join(configDir, "himo.yml")
		mustWriteFile(t, configPath, []byte(
			"setup:\n  hooks:\n    install:\n      command: 'mise install'\n"+
				"teardown:\n  hooks:\n    cleanup:\n      command: 'mise run cleanup'\n",
		))

		got, err := config.Load(configPath)
		require.NoError(t, err)

		assert.Len(t, got.Config.Setup.Hooks, 1)
		assert.Equal(t, "mise install", got.Config.Setup.Hooks["install"].Command)
		assert.Len(t, got.Config.Teardown.Hooks, 1)
		assert.Equal(t, "mise run cleanup", got.Config.Teardown.Hooks["cleanup"].Command)
	})

	t.Run("rejects setup hook with empty command", func(t *testing.T) {
		configDir := t.TempDir()
		configPath := filepath.Join(configDir, "himo.yml")
		mustWriteFile(t, configPath, []byte(
			"setup:\n  hooks:\n    install:\n      command: ''\n",
		))

		_, err := config.Load(configPath)
		assert.Error(t, err)
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
