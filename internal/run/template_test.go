// internal/run/template_test.go
package run_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/config"
	"github.com/sushichan044/seil/internal/run"
)

func TestEvalJob(t *testing.T) {
	t.Run("substitutes .File", func(t *testing.T) {
		tmpDir := t.TempDir()
		wsPath, err := config.NewWorkspacePath(tmpDir, "main.go")
		require.NoError(t, err)

		got, err := run.EvalJob(
			"gofmt -w {{.File}}",
			run.JobEvaluationParams{Filepath: wsPath},
		)
		require.NoError(t, err)
		assert.Equal(t, "gofmt -w main.go", got)
	})

	t.Run("dir function returns directory of file", func(t *testing.T) {
		tmpDir := t.TempDir()
		wsPath, err := config.NewWorkspacePath(tmpDir, "pkg/foo/bar.go")
		require.NoError(t, err)

		got, err := run.EvalJob("go test {{dir .File}}", run.JobEvaluationParams{Filepath: wsPath})
		require.NoError(t, err)
		assert.Equal(t, "go test pkg/foo", got)
	})

	t.Run("base function returns base name", func(t *testing.T) {
		tmpDir := t.TempDir()
		wsPath, err := config.NewWorkspacePath(tmpDir, "pkg/foo/bar.go")
		require.NoError(t, err)

		got, err := run.EvalJob("echo {{base .File}}", run.JobEvaluationParams{Filepath: wsPath})
		require.NoError(t, err)
		assert.Equal(t, "echo bar.go", got)
	})

	t.Run("ext function returns extension", func(t *testing.T) {
		tmpDir := t.TempDir()
		wsPath, err := config.NewWorkspacePath(tmpDir, "main.go")
		require.NoError(t, err)

		got, err := run.EvalJob("echo {{ext .File}}", run.JobEvaluationParams{Filepath: wsPath})
		require.NoError(t, err)
		assert.Equal(t, "echo .go", got)
	})

	t.Run("no template directives returns input unchanged", func(t *testing.T) {
		tmpDir := t.TempDir()
		wsPath, err := config.NewWorkspacePath(tmpDir, "main.go")
		require.NoError(t, err)

		got, err := run.EvalJob("echo hello", run.JobEvaluationParams{Filepath: wsPath})
		require.NoError(t, err)
		assert.Equal(t, "echo hello", got)
	})

	t.Run("returns error on invalid template syntax", func(t *testing.T) {
		tmpDir := t.TempDir()
		wsPath, err := config.NewWorkspacePath(tmpDir, "main.go")
		require.NoError(t, err)

		_, err = run.EvalJob("{{.File | unknown}}", run.JobEvaluationParams{Filepath: wsPath})
		assert.Error(t, err)
	})

	t.Run("returns error on template execution failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		wsPath, err := config.NewWorkspacePath(tmpDir, "main.go")
		require.NoError(t, err)

		_, err = run.EvalJob("{{.NonExistent}}", run.JobEvaluationParams{Filepath: wsPath})
		assert.Error(t, err)
	})
}

func TestEvalJobRecipe(t *testing.T) {
	t.Run("We can use dir fn to get go package and run tests in it", func(t *testing.T) {
		tmpDir := t.TempDir()
		wsPath, err := config.NewWorkspacePath(tmpDir, "pkg/foo/bar.go")
		require.NoError(t, err)

		got, err := run.EvalJob("go test ./{{dir .File}}", run.JobEvaluationParams{Filepath: wsPath})
		require.NoError(t, err)
		assert.Equal(t, "go test ./pkg/foo", got)
	})
}
