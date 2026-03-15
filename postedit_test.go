package seil_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil"
	"github.com/sushichan044/seil/internal/config"
	"github.com/sushichan044/seil/internal/runner"
)

func TestWorkspace_RunPostEditHooks(t *testing.T) {
	t.Run("skips hook when glob does not match", func(t *testing.T) {
		repoDir := t.TempDir()
		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Jobs: []config.FilePatternHook{
						{Name: "fmt", Glob: "**/*.ts", Run: "echo ts"},
					},
				},
			},
		}

		ws, err := seil.NewWorkspace(cfg)
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
					Jobs: []config.FilePatternHook{
						{Name: "fmt", Glob: "**/*.go", Run: "echo go"},
					},
				},
			},
		}

		ws, err := seil.NewWorkspace(cfg)
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
					Jobs: []config.FilePatternHook{
						{Name: "echo", Glob: "**/*.go", Run: "echo hello"},
					},
				},
			},
		}

		ws, err := seil.NewWorkspace(cfg)
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
					Jobs: []config.FilePatternHook{
						{Name: "fail", Glob: "**/*.go", Run: "exit 1"},
					},
				},
			},
		}

		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
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
				PostEdit: config.PostEdit{
					Jobs: []config.FilePatternHook{
						{Name: "zzz", Glob: "**/*.go", Run: "echo zzz"},
						{Name: "aaa", Glob: "**/*.go", Run: "echo aaa"},
						{Name: "mmm", Glob: "**/*.go", Run: "echo mmm"},
					},
				},
			},
		}

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
		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Jobs: []config.FilePatternHook{
						{Name: "always", Glob: "", Run: "echo always"},
					},
				},
			},
		}

		ws, err := seil.NewWorkspace(cfg)
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
					Jobs: []config.FilePatternHook{
						{Name: "always", Glob: "", Run: "echo always"},
					},
				},
			},
		}

		ws, err := seil.NewWorkspace(cfg)
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
					Jobs: []config.FilePatternHook{
						{Name: "echo", Glob: "**/*.go", Run: `echo {{.Files | join " "}}`},
					},
				},
			},
		}

		ws, err := seil.NewWorkspace(cfg)
		require.NoError(t, err)

		results, err := ws.RunPostEditHooks(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, runner.HookStatusSuccess, results[0].Status)
		assert.Equal(t, "main.go", results[0].Summary)
	})
}
