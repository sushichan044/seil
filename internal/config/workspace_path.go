package config

import "path/filepath"

// WorkspacePath is a file path anchored to a workspace root.
// Relative paths are interpreted as relative to root (not CWD).
type WorkspacePath struct {
	root string
	abs  string
}

// NewWorkspacePath constructs a WorkspacePath.
// If filePath is relative, it is joined to root (not the current working directory).
func NewWorkspacePath(root, filePath string) WorkspacePath {
	abs := filePath
	if !filepath.IsAbs(filePath) {
		abs = filepath.Join(root, filePath)
	}
	return WorkspacePath{root: root, abs: abs}
}

// Abs returns the absolute path of the file.
func (p WorkspacePath) Abs() string { return p.abs }

// Rel returns the path relative to the workspace root.
// Never fails because both root and abs are guaranteed to be absolute.
func (p WorkspacePath) Rel() string {
	rel, _ := filepath.Rel(p.root, p.abs)
	return rel
}
