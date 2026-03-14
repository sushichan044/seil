package postedit_test

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/gitignore"
	"github.com/sushichan044/himo/internal/postedit"
)

func mustMatcher(t *testing.T, fs afero.Fs, root string) *gitignore.Matcher {
	t.Helper()
	m, err := gitignore.NewMatcherFromRoot(fs, root)
	require.NoError(t, err)
	return m
}

func TestRunner_Run(t *testing.T) {
	t.Run("skips hook when glob does not match", func(t *testing.T) {
		repoDir := t.TempDir()
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll(repoDir, 0o700))

		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.Hook{
						"fmt": {Glob: "**/*.ts", Command: "echo ts"},
					},
				},
			},
		}

		r := &postedit.Runner{
			Config:    cfg,
			Gitignore: mustMatcher(t, fs, repoDir),
			Fs:        fs,
		}

		results, err := r.Run(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, postedit.HookStatusSkipped, results[0].Status)
	})

	t.Run("skips hook when file is gitignored", func(t *testing.T) {
		repoDir := t.TempDir()
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll(repoDir, 0o700))
		require.NoError(t, afero.WriteFile(fs, repoDir+"/.gitignore", []byte("*.go\n"), 0o600))

		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.Hook{
						"fmt": {Glob: "**/*.go", Command: "echo go"},
					},
				},
			},
		}

		r := &postedit.Runner{
			Config:    cfg,
			Gitignore: mustMatcher(t, fs, repoDir),
			Fs:        fs,
		}

		results, err := r.Run(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, postedit.HookStatusSkipped, results[0].Status)
	})

	t.Run("executes hook and returns success on exit code 0", func(t *testing.T) {
		repoDir := t.TempDir()
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll(repoDir, 0o700))

		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.Hook{
						"echo": {Glob: "**/*.go", Command: "echo hello"},
					},
				},
			},
		}

		r := &postedit.Runner{
			Config:    cfg,
			Gitignore: mustMatcher(t, fs, repoDir),
			Fs:        fs,
		}

		results, err := r.Run(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "echo", results[0].Name)
		assert.Equal(t, postedit.HookStatusSuccess, results[0].Status)
		assert.Equal(t, 0, results[0].ExitCode)
		assert.Equal(t, "hello", results[0].Summary)
	})

	t.Run("returns failure when command exits non-zero", func(t *testing.T) {
		repoDir := t.TempDir()
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll(repoDir, 0o700))

		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.Hook{
						"fail": {Glob: "**/*.go", Command: "exit 1"},
					},
				},
			},
		}

		r := &postedit.Runner{
			Config:    cfg,
			Gitignore: mustMatcher(t, fs, repoDir),
			Fs:        fs,
		}

		results, err := r.Run(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, postedit.HookStatusFailure, results[0].Status)
		assert.Equal(t, 1, results[0].ExitCode)
	})

	t.Run("returns hooks sorted alphabetically", func(t *testing.T) {
		repoDir := t.TempDir()
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll(repoDir, 0o700))

		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.Hook{
						"zzz": {Glob: "**/*.go", Command: "echo zzz"},
						"aaa": {Glob: "**/*.go", Command: "echo aaa"},
						"mmm": {Glob: "**/*.go", Command: "echo mmm"},
					},
				},
			},
		}

		r := &postedit.Runner{
			Config:    cfg,
			Gitignore: mustMatcher(t, fs, repoDir),
			Fs:        fs,
		}

		results, err := r.Run(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, "aaa", results[0].Name)
		assert.Equal(t, "mmm", results[1].Name)
		assert.Equal(t, "zzz", results[2].Name)
	})

	t.Run("template variables are expanded in command", func(t *testing.T) {
		repoDir := t.TempDir()
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll(repoDir, 0o700))

		cfg := &config.ResolvedConfig{
			CWD: repoDir,
			Config: config.Config{
				PostEdit: config.PostEdit{
					Hooks: map[string]config.Hook{
						"echo": {Glob: "**/*.go", Command: `echo {{.Files | join " "}}`},
					},
				},
			},
		}

		r := &postedit.Runner{
			Config:    cfg,
			Gitignore: mustMatcher(t, fs, repoDir),
			Fs:        fs,
		}

		results, err := r.Run(context.Background(), "main.go")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, postedit.HookStatusSuccess, results[0].Status)
		assert.Equal(t, "main.go", results[0].Summary)
	})
}
