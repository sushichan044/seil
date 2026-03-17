// internal/run/template_test.go
package run_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/run"
)

func TestEvalJob(t *testing.T) {
	t.Run("substitutes .File", func(t *testing.T) {
		got, err := run.EvalJob("gofmt -w {{.File}}", run.Vars{File: "main.go"})
		require.NoError(t, err)
		assert.Equal(t, "gofmt -w main.go", got)
	})

	t.Run("dir function returns directory of file", func(t *testing.T) {
		got, err := run.EvalJob("go test {{.File | dir}}", run.Vars{File: "pkg/foo/bar.go"})
		require.NoError(t, err)
		assert.Equal(t, "go test pkg/foo", got)
	})

	t.Run("base function returns base name", func(t *testing.T) {
		got, err := run.EvalJob("echo {{.File | base}}", run.Vars{File: "pkg/foo/bar.go"})
		require.NoError(t, err)
		assert.Equal(t, "echo bar.go", got)
	})

	t.Run("ext function returns extension", func(t *testing.T) {
		got, err := run.EvalJob("echo {{.File | ext}}", run.Vars{File: "main.go"})
		require.NoError(t, err)
		assert.Equal(t, "echo .go", got)
	})

	t.Run("no template directives returns input unchanged", func(t *testing.T) {
		got, err := run.EvalJob("echo hello", run.Vars{File: "main.go"})
		require.NoError(t, err)
		assert.Equal(t, "echo hello", got)
	})

	t.Run("returns error on invalid template syntax", func(t *testing.T) {
		_, err := run.EvalJob("{{.File | unknown}}", run.Vars{File: "main.go"})
		assert.Error(t, err)
	})

	t.Run("returns error on template execution failure", func(t *testing.T) {
		_, err := run.EvalJob("{{.NonExistent}}", run.Vars{File: "main.go"})
		assert.Error(t, err)
	})
}
