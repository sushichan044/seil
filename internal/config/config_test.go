package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/config"
	"github.com/sushichan044/seil/internal/git"
)

func TestResolveConfigFilePathFrom(t *testing.T) {
	t.Run("finds seil.yml in ancestor directory within git repo", func(t *testing.T) {
		repoRoot := t.TempDir()
		mustMkdirAll(t, filepath.Join(repoRoot, ".git"))

		configPath := filepath.Join(repoRoot, "nested", "seil.yml")
		mustMkdirAll(t, filepath.Dir(configPath))
		mustWriteFile(t, configPath, []byte("post_edit:\n  jobs: []\n"))

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
		mustWriteFile(t, filepath.Join(parent, "seil.yml"), []byte("post_edit:\n  jobs: []\n"))

		got, err := config.ResolveConfigFilePathFrom(startDir)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("starts from parent when given a file path", func(t *testing.T) {
		repoRoot := t.TempDir()
		mustMkdirAll(t, filepath.Join(repoRoot, ".git"))

		configPath := filepath.Join(repoRoot, "seil.yml")
		mustWriteFile(t, configPath, []byte("post_edit:\n  jobs: []\n"))

		sourceFilePath := filepath.Join(repoRoot, "nested", "file.go")
		mustMkdirAll(t, filepath.Dir(sourceFilePath))
		mustWriteFile(t, sourceFilePath, []byte("package nested\n"))

		got, err := config.ResolveConfigFilePathFrom(sourceFilePath)
		require.NoError(t, err)
		assert.Equal(t, configPath, got)
	})

	t.Run("returns not found when outside git repo", func(t *testing.T) {
		startDir := t.TempDir()
		mustWriteFile(t, filepath.Join(startDir, "seil.yml"), []byte("post_edit:\n  jobs: []\n"))

		_, err := config.ResolveConfigFilePathFrom(startDir)
		assert.ErrorIs(t, err, git.ErrNotInGitRepo)
	})
}

func TestParseConfig(t *testing.T) {
	t.Run("normalizes spaces in FilePatternHook name to hyphens", func(t *testing.T) {
		data := []byte(`
post_edit:
  jobs:
    - name: 'go fmt'
      glob: '**/*.go'
      run: 'gofmt -w'
`)

		got, err := config.ParseConfig(data)
		require.NoError(t, err)

		assert.Equal(t, "go-fmt", got.PostEdit.Jobs[0].Name)
	})

	t.Run("normalizes spaces in SimpleHook name to hyphens", func(t *testing.T) {
		data := []byte(`
setup:
  jobs:
    - name: 'go tidy'
      run: 'go mod tidy'
`)

		got, err := config.ParseConfig(data)
		require.NoError(t, err)

		assert.Equal(t, "go-tidy", got.Setup.Jobs[0].Name)
	})

	t.Run("uses truncated run as SimpleHook name when name is omitted", func(t *testing.T) {
		data := []byte(`
setup:
  jobs:
    - run: 'echo installed'
`)

		got, err := config.ParseConfig(data)
		require.NoError(t, err)

		assert.Equal(t, "echo-installed", got.Setup.Jobs[0].Name)
	})

	t.Run("uses truncated run as Teardown SimpleHook name when name is omitted", func(t *testing.T) {
		data := []byte(`
teardown:
  jobs:
    - run: 'echo cleaned'
`)

		got, err := config.ParseConfig(data)
		require.NoError(t, err)

		assert.Equal(t, "echo-cleaned", got.Teardown.Jobs[0].Name)
	})

	t.Run("sanitizes slashes and special chars in fallback name", func(t *testing.T) {
		data := []byte(`
setup:
  jobs:
    - run: 'go test ./...'
`)

		got, err := config.ParseConfig(data)
		require.NoError(t, err)

		assert.Equal(t, "go-test-.-...", got.Setup.Jobs[0].Name)
	})
}

func TestLoad(t *testing.T) {
	t.Run("uses explicit config path and sets cwd from it", func(t *testing.T) {
		configDir := filepath.Join(t.TempDir(), "config")
		configPath := filepath.Join(configDir, "seil.yml")
		mustMkdirAll(t, configDir)
		mustWriteFile(t, configPath, []byte(`
post_edit:
  jobs:
    - name: fmt
      glob: '**/*.go'
      run: 'go fmt ./...'
`))

		got, err := config.Load(configPath)
		require.NoError(t, err)

		assert.Equal(t, configPath, got.Path)
		assert.Equal(t, configDir, got.CWD)
		assert.Len(t, got.Config.PostEdit.Jobs, 1)
		assert.Equal(t, "**/*.go", got.Config.PostEdit.Jobs[0].Glob)
	})

	t.Run("loads setup and teardown hooks", func(t *testing.T) {
		configDir := t.TempDir()
		configPath := filepath.Join(configDir, "seil.yml")
		mustWriteFile(t, configPath, []byte(`
setup:
  jobs:
    - name: install
      run: 'mise install'
teardown:
  jobs:
    - name: cleanup
      run: 'mise run cleanup'
`))

		got, err := config.Load(configPath)
		require.NoError(t, err)

		assert.Len(t, got.Config.Setup.Jobs, 1)
		assert.Equal(t, "mise install", got.Config.Setup.Jobs[0].Run)
		assert.Len(t, got.Config.Teardown.Jobs, 1)
		assert.Equal(t, "mise run cleanup", got.Config.Teardown.Jobs[0].Run)
	})

	t.Run("rejects setup hook with empty command", func(t *testing.T) {
		configDir := t.TempDir()
		configPath := filepath.Join(configDir, "seil.yml")
		mustWriteFile(t, configPath, []byte(`
setup:
  jobs:
    - name: install
      run: ''
`))

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
