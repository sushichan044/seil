package config

import (
	"os"
	"path/filepath"
	"testing"

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
		if err != nil {
			t.Fatalf("FindUpConfig() error = %v", err)
		}
		if got.Path != configPath {
			t.Fatalf("FindUpConfig().Path = %q, want %q", got.Path, configPath)
		}
		if got.CWD != filepath.Dir(configPath) {
			t.Fatalf("FindUpConfig().CWD = %q, want %q", got.CWD, filepath.Dir(configPath))
		}
	})

	t.Run("returns not found when config does not exist within git repo", func(t *testing.T) {
		parent := t.TempDir()
		repoRoot := filepath.Join(parent, "repo")
		startDir := filepath.Join(repoRoot, "nested")

		mustMkdirAll(t, filepath.Join(repoRoot, ".git"))
		mustMkdirAll(t, startDir)
		mustWriteFile(t, filepath.Join(parent, defaultConfigFileName), []byte("post_edit:\n  hooks: []\n"))

		_, err := FindUpAndLoad(startDir)
		require.Error(t, err)
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
		if err != nil {
			t.Fatalf("FindUpConfig() error = %v", err)
		}
		if got.Path != configPath {
			t.Fatalf("FindUpConfig().Path = %q, want %q", got.Path, configPath)
		}
	})

	t.Run("returns not found when outside git repo", func(t *testing.T) {
		startDir := t.TempDir()
		mustWriteFile(t, filepath.Join(startDir, defaultConfigFileName), []byte("post_edit:\n  hooks: []\n"))

		_, err := FindUpAndLoad(startDir)
		require.Error(t, err)
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("uses explicit config path and sets cwd from it", func(t *testing.T) {
		configDir := filepath.Join(t.TempDir(), "config")
		configPath := filepath.Join(configDir, defaultConfigFileName)
		mustMkdirAll(t, configDir)
		mustWriteFile(t, configPath, []byte("post_edit:\n  hooks:\n    - glob: \"**/*.go\"\n      command: go fmt ./...\n"))

		got, err := Load(configPath)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if got.Path != configPath {
			t.Fatalf("LoadConfig().Path = %q, want %q", got.Path, configPath)
		}
		if got.CWD != configDir {
			t.Fatalf("LoadConfig().CWD = %q, want %q", got.CWD, configDir)
		}
		if len(got.Config.PostEdit.Hooks) != 1 {
			t.Fatalf("LoadConfig().Config.PostEdit.Hooks length = %d, want 1", len(got.Config.PostEdit.Hooks))
		}
	})
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
}
