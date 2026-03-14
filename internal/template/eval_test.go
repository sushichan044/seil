package template_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmpl "github.com/sushichan044/himo/internal/template"
)

func TestEvalConfig(t *testing.T) {
	t.Run("returns input unchanged when no template directives", func(t *testing.T) {
		got, err := tmpl.EvalConfig("post_edit:\n  hooks: {}", tmpl.ConfigVars{})
		require.NoError(t, err)
		assert.Equal(t, "post_edit:\n  hooks: {}", got)
	})

	t.Run("returns error on invalid template syntax", func(t *testing.T) {
		_, err := tmpl.EvalConfig("{{.Invalid", tmpl.ConfigVars{})
		assert.Error(t, err)
	})
}

func TestEvalCommand(t *testing.T) {
	t.Run("substitutes .Files into command", func(t *testing.T) {
		got, err := tmpl.EvalCommand("gofmt -w {{.Files | join \" \"}}", tmpl.CommandVars{
			Files: []string{"a.go", "b.go"},
		})
		require.NoError(t, err)
		assert.Equal(t, "gofmt -w a.go b.go", got)
	})

	t.Run("handles single file", func(t *testing.T) {
		got, err := tmpl.EvalCommand("go vet {{.Files | join \" \"}}", tmpl.CommandVars{
			Files: []string{"main.go"},
		})
		require.NoError(t, err)
		assert.Equal(t, "go vet main.go", got)
	})

	t.Run("returns error on invalid template syntax", func(t *testing.T) {
		_, err := tmpl.EvalCommand("{{.Files | unknown}}", tmpl.CommandVars{
			Files: []string{"a.go"},
		})
		assert.Error(t, err)
	})
}
