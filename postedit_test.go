package seil_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil"
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

func TestWorkspace_RunPostEditHooks(t *testing.T) {
	t.Run("skips hook when glob does not match", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
post_edit:
  jobs:
    - name: fmt
      glob: "**/*.ts"
      run: echo ts
`)
		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, run.StatusSkipped, results[0].Status)
	})

	t.Run("skips hook when file is gitignored", func(t *testing.T) {
		repoDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(repoDir, ".gitignore"), []byte("*.go\n"), 0o600))
		cfg := loadCfg(t, repoDir, `
post_edit:
  jobs:
    - name: fmt
      glob: "**/*.go"
      run: echo go
`)
		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, run.StatusSkipped, results[0].Status)
	})

	t.Run("executes hook and returns success on exit code 0", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
post_edit:
  jobs:
    - name: echo
      glob: "**/*.go"
      run: echo hello
`)
		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "echo", results[0].Name)
		assert.Equal(t, run.StatusSuccess, results[0].Status)
	})

	t.Run("returns failure when command exits non-zero", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
post_edit:
  jobs:
    - name: fail
      glob: "**/*.go"
      run: exit 1
`)
		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, run.StatusFailure, results[0].Status)
	})

	t.Run("returns hooks in definition order", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
post_edit:
  jobs:
    - name: zzz
      glob: "**/*.go"
      run: echo zzz
    - name: aaa
      glob: "**/*.go"
      run: echo aaa
    - name: mmm
      glob: "**/*.go"
      run: echo mmm
`)
		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, "zzz", results[0].Name)
		assert.Equal(t, "aaa", results[1].Name)
		assert.Equal(t, "mmm", results[2].Name)
	})

	t.Run("executes hook when glob is empty (always run)", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
post_edit:
  jobs:
    - name: always
      run: echo always
`)
		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, run.StatusSuccess, results[0].Status)
	})

	t.Run("skips hook when glob is empty but file is gitignored", func(t *testing.T) {
		repoDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(repoDir, ".gitignore"), []byte("*.go\n"), 0o600))
		cfg := loadCfg(t, repoDir, `
post_edit:
  jobs:
    - name: always
      run: echo always
`)
		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, run.StatusSkipped, results[0].Status)
	})

	t.Run("template variables are expanded in command", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
post_edit:
  jobs:
    - name: echo
      glob: "**/*.go"
      run: 'echo {{.File}}'
`)
		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, run.StatusSuccess, results[0].Status)
	})
}
