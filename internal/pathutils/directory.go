package pathutils

import (
	"os"
	"path/filepath"
)

// DetermineDirectory returns the directory path for the given path.
//
// If the path is a file, it returns the parent directory.
// If it's already a directory, it returns the path as is.
func DetermineDirectory(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return filepath.Dir(absPath), nil
	}

	return absPath, nil
}
