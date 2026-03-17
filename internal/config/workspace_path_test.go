package config_test

import (
	"path/filepath"
	"testing"

	"github.com/sushichan044/seil/internal/config"
)

func TestNewWorkspacePath(t *testing.T) {
	t.Run("absolute filePath stays absolute", func(t *testing.T) {
		root := t.TempDir()
		abs := filepath.Join(root, "sub", "file.go")
		p := config.NewWorkspacePath(root, abs)
		if got := p.Abs(); got != abs {
			t.Errorf("Abs() = %q; want %q", got, abs)
		}
	})

	t.Run("relative filePath is interpreted as relative to root", func(t *testing.T) {
		root := t.TempDir()
		p := config.NewWorkspacePath(root, "main.go")
		want := filepath.Join(root, "main.go")
		if got := p.Abs(); got != want {
			t.Errorf("Abs() = %q; want %q", got, want)
		}
	})

	t.Run("Rel returns path relative to root", func(t *testing.T) {
		root := t.TempDir()
		abs := filepath.Join(root, "src", "foo.go")
		p := config.NewWorkspacePath(root, abs)
		want := filepath.Join("src", "foo.go")
		if got := p.Rel(); got != want {
			t.Errorf("Rel() = %q; want %q", got, want)
		}
	})

	t.Run("Rel for root-relative input equals the input", func(t *testing.T) {
		root := t.TempDir()
		p := config.NewWorkspacePath(root, "main.go")
		if got := p.Rel(); got != "main.go" {
			t.Errorf("Rel() = %q; want %q", got, "main.go")
		}
	})
}
