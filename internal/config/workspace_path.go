package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// WorkspacePath is a file path anchored to a workspace root.
// Relative paths are interpreted as relative to root (not CWD).
type WorkspacePath struct {
	root string
	abs  string
}

// NewWorkspacePath constructs a WorkspacePath.
// If filePath is relative, it is joined to root (not the current working directory).
// Returns an error if a relative filePath escapes the workspace root (e.g. "../../etc/passwd").
// Absolute paths are accepted regardless of whether they are inside the root.
func NewWorkspacePath(root, filePath string) (WorkspacePath, error) {
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return WorkspacePath{}, fmt.Errorf("failed to resolve workspace root %q: %w", root, err)
	}

	abs := filePath
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(absRoot, abs)
	}
	abs, err = filepath.Abs(filepath.Clean(abs))
	if err != nil {
		return WorkspacePath{}, fmt.Errorf("failed to resolve path %q: %w", filePath, err)
	}

	rel, err := filepath.Rel(absRoot, abs)
	relSlash := filepath.ToSlash(rel)
	if err != nil || rel == ".." || strings.HasPrefix(relSlash, "../") {
		return WorkspacePath{}, fmt.Errorf("path %q is outside workspace root %q", filePath, absRoot)
	}
	return WorkspacePath{root: absRoot, abs: abs}, nil
}

// Abs returns the absolute path of the file.
func (p WorkspacePath) Abs() string { return p.abs }

// Rel returns the path relative to the workspace root.
// Never fails because both root and abs are guaranteed to be absolute.
func (p WorkspacePath) Rel() string {
	rel, _ := filepath.Rel(p.root, p.abs)
	return rel
}
