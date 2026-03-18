package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// OutsideWorkspaceError is returned when a file path is outside the workspace root.
type OutsideWorkspaceError struct {
	FilePath string
	Root     string
}

func (e *OutsideWorkspaceError) Error() string {
	return fmt.Sprintf("path %q is outside workspace root %q", e.FilePath, e.Root)
}

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
// Both root and filePath are resolved through symlinks so that paths differing only
// by symlink chains (e.g. macOS /tmp → /private/tmp) are handled correctly.
func NewWorkspacePath(root, filePath string) (WorkspacePath, error) {
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return WorkspacePath{}, fmt.Errorf("failed to resolve workspace root %q: %w", root, err)
	}

	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return WorkspacePath{}, fmt.Errorf("failed to resolve symlinks for workspace root %q: %w", root, err)
	}

	abs := filePath
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(realRoot, abs)
	}
	abs, err = filepath.Abs(filepath.Clean(abs))
	if err != nil {
		return WorkspacePath{}, fmt.Errorf("failed to resolve path %q: %w", filePath, err)
	}

	realAbs, err := filepath.EvalSymlinks(abs)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return WorkspacePath{}, fmt.Errorf("failed to resolve symlinks for path %q: %w", filePath, err)
		}
		realAbs = abs
	}

	rel, err := filepath.Rel(realRoot, realAbs)
	relSlash := filepath.ToSlash(rel)
	if err != nil || rel == ".." || strings.HasPrefix(relSlash, "../") {
		return WorkspacePath{}, &OutsideWorkspaceError{FilePath: filePath, Root: realRoot}
	}
	return WorkspacePath{root: realRoot, abs: realAbs}, nil
}

// Abs returns the absolute path of the file.
func (p WorkspacePath) Abs() string { return p.abs }

// Rel returns the path relative to the workspace root.
// Never fails because both root and abs are guaranteed to be absolute.
func (p WorkspacePath) Rel() string {
	rel, _ := filepath.Rel(p.root, p.abs)
	return rel
}
