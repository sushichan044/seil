package himo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	himo "github.com/sushichan044/himo"
	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/runner"
)

func TestWorkspace_RunSetupHooks(t *testing.T) {
	t.Run("executes hook and returns success on exit code 0", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				Setup: config.Setup{
					Jobs: []config.SimpleHook{
						{Name: "install", Run: "echo installed"},
					},
				},
			},
		}

		ws, err := himo.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunSetupHooks(context.Background())
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "install", results[0].Name)
		assert.Equal(t, runner.HookStatusSuccess, results[0].Status)
		assert.Equal(t, 0, results[0].ExitCode)
		assert.Equal(t, "installed", results[0].Summary)
	})

	t.Run("returns failure when command exits non-zero", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				Setup: config.Setup{
					Jobs: []config.SimpleHook{
						{Name: "fail", Run: "exit 1"},
					},
				},
			},
		}

		ws, err := himo.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunSetupHooks(context.Background())
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, runner.HookStatusFailure, results[0].Status)
		assert.Equal(t, 1, results[0].ExitCode)
	})

	t.Run("returns hooks in definition order", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				Setup: config.Setup{
					Jobs: []config.SimpleHook{
						{Name: "zzz", Run: "echo zzz"},
						{Name: "aaa", Run: "echo aaa"},
						{Name: "mmm", Run: "echo mmm"},
					},
				},
			},
		}

		ws, err := himo.NewWorkspace(cfg)
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
		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				Setup: config.Setup{
					Jobs: []config.SimpleHook{},
				},
			},
		}

		ws, err := himo.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunSetupHooks(context.Background())
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}
