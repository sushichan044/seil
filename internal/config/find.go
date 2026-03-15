package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

var ErrConfigNotFound = errors.New("config file not found")

func FindConfigFile(fs afero.Fs, fromPath string) (string, error) {
	startDir := filepath.Dir(fromPath)

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
