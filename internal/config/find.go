package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

var ErrConfigNotFound = errors.New("config file not found")

func FindConfigFile(fs afero.Fs, fromPath string) (string, error) {
	startDir := filepath.Clean(fromPath)

	info, err := fs.Stat(startDir)
	switch {
	case err == nil && info.IsDir():
		// keep startDir as is
	case err == nil || os.IsNotExist(err):
		startDir = filepath.Dir(startDir)
	default:
		return "", err
	}

	for dir := startDir; ; dir = filepath.Dir(dir) {
		configPath := filepath.Join(dir, defaultConfigFileName)
		if _, statErr := fs.Stat(configPath); statErr == nil {
			return configPath, nil
		} else if !os.IsNotExist(statErr) {
			return "", statErr
		}

		parent := filepath.Dir(dir)
		// Now we are at the root. Stop searching.
		if parent == dir {
			return "", ErrConfigNotFound
		}
	}
}
