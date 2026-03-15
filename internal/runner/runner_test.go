package runner_test

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/himo/internal/runner"
)

func TestRunner_Run(t *testing.T) {
	t.Run("executes hook and returns success on exit code 0", func(t *testing.T) {
		repoDir := t.TempDir()
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll(repoDir, 0o700))

		r := &runner.Runner{WorkDir: repoDir, Fs: fs}
		results, err := r.Run(context.Background(), []runner.Hook{
			{Name: "echo", Command: "echo hello"},
		})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "echo", results[0].Name)
		assert.Equal(t, runner.HookStatusSuccess, results[0].Status)
		assert.Equal(t, 0, results[0].ExitCode)
		assert.Equal(t, "hello", results[0].Summary)
	})

	t.Run("returns failure when command exits non-zero", func(t *testing.T) {
		repoDir := t.TempDir()
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll(repoDir, 0o700))

		r := &runner.Runner{WorkDir: repoDir, Fs: fs}
		results, err := r.Run(context.Background(), []runner.Hook{
			{Name: "fail", Command: "exit 1"},
		})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, runner.HookStatusFailure, results[0].Status)
		assert.Equal(t, 1, results[0].ExitCode)
	})

	t.Run("returns multiple hooks in given order", func(t *testing.T) {
		repoDir := t.TempDir()
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll(repoDir, 0o700))

		r := &runner.Runner{WorkDir: repoDir, Fs: fs}
		results, err := r.Run(context.Background(), []runner.Hook{
			{Name: "zzz", Command: "echo zzz"},
			{Name: "aaa", Command: "echo aaa"},
			{Name: "mmm", Command: "echo mmm"},
		})

		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, "zzz", results[0].Name)
		assert.Equal(t, "aaa", results[1].Name)
		assert.Equal(t, "mmm", results[2].Name)
	})

	t.Run("returns empty slice for empty hooks", func(t *testing.T) {
		repoDir := t.TempDir()
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll(repoDir, 0o700))

		r := &runner.Runner{WorkDir: repoDir, Fs: fs}
		results, err := r.Run(context.Background(), []runner.Hook{})

		require.NoError(t, err)
		assert.Empty(t, results)
	})
}
