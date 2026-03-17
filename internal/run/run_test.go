package run_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/config"
	"github.com/sushichan044/seil/internal/run"
)

func loadCfg(t *testing.T, repoDir, yaml string) *config.ResolvedConfig {
	t.Helper()
	cfgPath := filepath.Join(repoDir, ".seil.yml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(yaml), 0o600))
	cfg, err := config.Load(afero.NewOsFs(), cfgPath)
	require.NoError(t, err)
	return cfg
}

func TestPrepare(t *testing.T) {
	repoDir := t.TempDir()
	cfg := loadCfg(t, repoDir, `
setup:
  jobs:
    - name: hello
      run: echo hello
`)
	r, err := run.Prepare(afero.NewOsFs(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, r)
}

func TestJobRunner_RunSetup(t *testing.T) {
	t.Run("returns empty slice when no setup jobs", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `{}`)
		r, err := run.Prepare(afero.NewOsFs(), cfg)
		require.NoError(t, err)

		results, err := r.RunSetup(context.Background())
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("returns success for zero-exit command", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
setup:
  jobs:
    - name: greet
      run: echo hello
`)
		r, err := run.Prepare(afero.NewOsFs(), cfg)
		require.NoError(t, err)

		results, err := r.RunSetup(context.Background())
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "greet", results[0].Name)
		assert.Equal(t, run.StatusSuccess, results[0].Status)
	})

	t.Run("returns failure for non-zero exit command", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
setup:
  jobs:
    - name: fail
      run: exit 1
`)
		r, err := run.Prepare(afero.NewOsFs(), cfg)
		require.NoError(t, err)

		results, err := r.RunSetup(context.Background())
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, run.StatusFailure, results[0].Status)
	})

	t.Run("returns failure when command template is invalid", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
setup:
  jobs:
    - name: bad-template
      run: '{{.Unknown}}'
`)
		r, err := run.Prepare(afero.NewOsFs(), cfg)
		require.NoError(t, err)

		results, err := r.RunSetup(context.Background())
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, run.StatusFailure, results[0].Status)
	})
}

func TestJobRunner_RunTeardown(t *testing.T) {
	t.Run("returns empty slice when no teardown jobs", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `{}`)
		r, err := run.Prepare(afero.NewOsFs(), cfg)
		require.NoError(t, err)

		results, err := r.RunTeardown(context.Background())
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("returns success for zero-exit command", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
teardown:
  jobs:
    - name: cleanup
      run: echo cleanup
`)
		r, err := run.Prepare(afero.NewOsFs(), cfg)
		require.NoError(t, err)

		results, err := r.RunTeardown(context.Background())
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "cleanup", results[0].Name)
		assert.Equal(t, run.StatusSuccess, results[0].Status)
	})
}

func TestJobRunner_RunPostEdit(t *testing.T) {
	t.Run("returns success for zero-exit command", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
post_edit:
  jobs:
    - name: fmt
      glob: "**/*.go"
      run: echo fmt
`)
		r, err := run.Prepare(afero.NewOsFs(), cfg)
		require.NoError(t, err)

		jobs := cfg.Config.PostEdit.Jobs
		wsPath, err := config.NewWorkspacePath(repoDir, "main.go")
		require.NoError(t, err)
		results, err := r.RunPostEdit(context.Background(), wsPath, jobs)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "fmt", results[0].Name)
		assert.Equal(t, run.StatusSuccess, results[0].Status)
	})

	t.Run("expands File template variable", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
post_edit:
  jobs:
    - name: echo-file
      glob: "**/*.go"
      run: 'echo {{.File}}'
`)
		r, err := run.Prepare(afero.NewOsFs(), cfg)
		require.NoError(t, err)

		jobs := cfg.Config.PostEdit.Jobs
		wsPath, err := config.NewWorkspacePath(repoDir, "main.go")
		require.NoError(t, err)
		results, err := r.RunPostEdit(context.Background(), wsPath, jobs)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, run.StatusSuccess, results[0].Status)
	})

	t.Run("returns failure for non-zero exit command", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
post_edit:
  jobs:
    - name: fail
      glob: "**/*.go"
      run: exit 1
`)
		r, err := run.Prepare(afero.NewOsFs(), cfg)
		require.NoError(t, err)

		jobs := cfg.Config.PostEdit.Jobs
		wsPath, err := config.NewWorkspacePath(repoDir, "main.go")
		require.NoError(t, err)
		results, err := r.RunPostEdit(context.Background(), wsPath, jobs)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, run.StatusFailure, results[0].Status)
	})
}
