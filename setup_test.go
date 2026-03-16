package seil_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil"
	"github.com/sushichan044/seil/internal/run"
)

func TestWorkspace_RunSetupHooks(t *testing.T) {
	t.Run("executes hook and returns success on exit code 0", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
setup:
  jobs:
    - name: install
      run: echo installed
`)
		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunSetupHooks(context.Background())
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "install", results[0].Name)
		assert.Equal(t, run.StatusSuccess, results[0].Status)
	})

	t.Run("returns failure when command exits non-zero", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
setup:
  jobs:
    - name: fail
      run: exit 1
`)
		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunSetupHooks(context.Background())
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, run.StatusFailure, results[0].Status)
	})

	t.Run("returns hooks in definition order", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `
setup:
  jobs:
    - name: zzz
      run: echo zzz
    - name: aaa
      run: echo aaa
    - name: mmm
      run: echo mmm
`)
		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunSetupHooks(context.Background())
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, "zzz", results[0].Name)
		assert.Equal(t, "aaa", results[1].Name)
		assert.Equal(t, "mmm", results[2].Name)
	})

	t.Run("returns empty slice when no hooks configured", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := loadCfg(t, repoDir, `{}`)

		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunSetupHooks(context.Background())
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}
