package git

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/sushichan044/seil/internal/pathutils"
)

var ErrNotInGitRepo = errors.New("not in a git repository")

func FindRepoRootFrom(path string) (string, error) {
	startDir, err := pathutils.DetermineDirectory(path)
	if err != nil {
		return "", err
	}

	for dir := startDir; ; dir = filepath.Dir(dir) {
		gitPath := filepath.Join(dir, ".git")
		if _, statErr := os.Stat(gitPath); statErr == nil {
			return dir, nil
		} else if !os.IsNotExist(statErr) {
			return "", statErr
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotInGitRepo
		}
	}
}
