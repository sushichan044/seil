package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/config"
)

func TestJob_DisplayName(t *testing.T) {
	t.Run("returns Name when set", func(t *testing.T) {
		job := config.Job{Name: "my-job", Run: "echo hello"}
		assert.Equal(t, "my-job", job.DisplayName())
	})

	t.Run("returns Run when Name is empty", func(t *testing.T) {
		job := config.Job{Run: "echo hello"}
		assert.Equal(t, "echo hello", job.DisplayName())
	})
}

func TestJob_PathSafeName(t *testing.T) {
	t.Run("returns Name when set with safe characters", func(t *testing.T) {
		job := config.Job{Name: "my_job-1.0", Run: "echo hello"}
		assert.Equal(t, "my_job-1.0", job.PathSafeName())
	})

	t.Run("replaces unsafe characters in Name with underscore", func(t *testing.T) {
		job := config.Job{Name: "my job/name", Run: "echo hello"}
		assert.Equal(t, "my_job_name", job.PathSafeName())
	})

	t.Run("uses Run when Name is empty", func(t *testing.T) {
		job := config.Job{Run: "echo hello"}
		assert.Equal(t, "echo_hello", job.PathSafeName())
	})

	t.Run("uses placeholder when both Name and Run are empty", func(t *testing.T) {
		job := config.Job{}
		assert.Equal(t, "_no_command_", job.PathSafeName())
	})

	t.Run("replaces unsafe characters in Run", func(t *testing.T) {
		job := config.Job{Run: "go test ./..."}
		assert.Equal(t, "go_test_._...", job.PathSafeName())
	})

	t.Run("preserves alphanumeric, underscore, dot, hyphen", func(t *testing.T) {
		job := config.Job{Name: "abc_123-foo.bar"}
		assert.Equal(t, "abc_123-foo.bar", job.PathSafeName())
	})
}

func TestGlobJob_Matches(t *testing.T) {
	t.Run("returns true when glob is empty (always run)", func(t *testing.T) {
		job := config.GlobJob{Glob: ""}
		p, err := config.NewWorkspacePath("/project", "main.go")
		require.NoError(t, err)
		assert.True(t, job.Matches(p))
	})

	t.Run("returns true when file matches glob pattern", func(t *testing.T) {
		job := config.GlobJob{Glob: "**/*.go"}
		p, err := config.NewWorkspacePath("/project", "/project/main.go")
		require.NoError(t, err)
		assert.True(t, job.Matches(p))
	})

	t.Run("returns false when file does not match glob pattern", func(t *testing.T) {
		job := config.GlobJob{Glob: "**/*.ts"}
		p, err := config.NewWorkspacePath("/project", "/project/main.go")
		require.NoError(t, err)
		assert.False(t, job.Matches(p))
	})

	t.Run("matches relative to configRoot", func(t *testing.T) {
		job := config.GlobJob{Glob: "src/**/*.go"}
		p1, err := config.NewWorkspacePath("/project", "/project/src/foo/bar.go")
		require.NoError(t, err)
		assert.True(t, job.Matches(p1))
		p2, err := config.NewWorkspacePath("/project", "/project/main.go")
		require.NoError(t, err)
		assert.False(t, job.Matches(p2))
	})
}
