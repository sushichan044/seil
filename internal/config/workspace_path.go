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
	abs := filePath
	if !filepath.IsAbs(filePath) {
		abs = filepath.Join(root, filePath)
		rel, err := filepath.Rel(root, abs)
		if err != nil || strings.HasPrefix(rel, "..") {
			return WorkspacePath{}, fmt.Errorf("path %q is outside workspace root %q", filePath, root)
		}
	}
	return WorkspacePath{root: root, abs: abs}, nil
}

// Abs returns the absolute path of the file.
func (p WorkspacePath) Abs() string { return p.abs }

// Rel returns the path relative to the workspace root.
// Never fails because both root and abs are guaranteed to be absolute.
func (p WorkspacePath) Rel() string {
	rel, _ := filepath.Rel(p.root, p.abs)
	return rel
}
