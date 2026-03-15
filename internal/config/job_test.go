package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

	t.Run("replaces unsafe characters in Run", func(t *testing.T) {
		job := config.Job{Run: "go test ./..."}
		assert.Equal(t, "go_test_._...", job.PathSafeName())
	})

	t.Run("preserves alphanumeric, underscore, dot, hyphen", func(t *testing.T) {
		job := config.Job{Name: "abc_123-foo.bar"}
		assert.Equal(t, "abc_123-foo.bar", job.PathSafeName())
	})
}
