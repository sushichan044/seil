package himo_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	himo "github.com/sushichan044/himo"
	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/runner"
)

func TestWorkspace_RunPostEditHooks(t *testing.T) {
	t.Run("skips hook when glob does not match", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.FilePatternHook{
						"fmt": {Glob: "**/*.ts", Command: "echo ts"},
					},
				},
			},
		}

		ws, err := himo.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, runner.HookStatusSkipped, results[0].Status)
	})

	t.Run("skips hook when file is gitignored", func(t *testing.T) {
		repoDir := t.TempDir()
		require.NoError(t, os.WriteFile(repoDir+"/.gitignore", []byte("*.go\n"), 0o600))

		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.FilePatternHook{
						"fmt": {Glob: "**/*.go", Command: "echo go"},
					},
				},
			},
		}

		ws, err := himo.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, runner.HookStatusSkipped, results[0].Status)
	})

	t.Run("executes hook and returns success on exit code 0", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.FilePatternHook{
						"echo": {Glob: "**/*.go", Command: "echo hello"},
					},
				},
			},
		}

		ws, err := himo.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "echo", results[0].Name)
		assert.Equal(t, runner.HookStatusSuccess, results[0].Status)
		assert.Equal(t, 0, results[0].ExitCode)
		assert.Equal(t, "hello", results[0].Summary)
	})

	t.Run("returns failure when command exits non-zero", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.FilePatternHook{
						"fail": {Glob: "**/*.go", Command: "exit 1"},
					},
				},
			},
		}

		ws, err := himo.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, runner.HookStatusFailure, results[0].Status)
		assert.Equal(t, 1, results[0].ExitCode)
	})

	t.Run("returns hooks sorted alphabetically", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.FilePatternHook{
						"zzz": {Glob: "**/*.go", Command: "echo zzz"},
						"aaa": {Glob: "**/*.go", Command: "echo aaa"},
						"mmm": {Glob: "**/*.go", Command: "echo mmm"},
					},
				},
			},
		}

		ws, err := himo.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, "aaa", results[0].Name)
		assert.Equal(t, "mmm", results[1].Name)
		assert.Equal(t, "zzz", results[2].Name)
	})

	t.Run("executes hook when glob is empty (always run)", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.FilePatternHook{
						"always": {Glob: "", Command: "echo always"},
					},
				},
			},
		}

		ws, err := himo.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, runner.HookStatusSuccess, results[0].Status)
	})

	t.Run("skips hook when glob is empty but file is gitignored", func(t *testing.T) {
		repoDir := t.TempDir()
		require.NoError(t, os.WriteFile(repoDir+"/.gitignore", []byte("*.go\n"), 0o600))

		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.FilePatternHook{
						"always": {Glob: "", Command: "echo always"},
					},
				},
			},
		}

		ws, err := himo.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, runner.HookStatusSkipped, results[0].Status)
	})

	t.Run("template variables are expanded in command", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.FilePatternHook{
						"echo": {Glob: "**/*.go", Command: `echo {{.Files | join " "}}`},
					},
				},
			},
		}

		ws, err := himo.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, runner.HookStatusSuccess, results[0].Status)
		assert.Equal(t, "main.go", results[0].Summary)
	})
}
