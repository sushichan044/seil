package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

var ErrConfigNotFound = errors.New("config file not found")

func FindConfigFile(fs afero.Fs, fromPath string) (string, error) {
	startDir, err := filepath.Abs(fromPath)
	if err != nil {
		return "", err
	}

	info, err := fs.Stat(startDir)
	switch {
	case err == nil && info.IsDir():
		// keep startDir as is
	case err == nil || os.IsNotExist(err):
		startDir = filepath.Dir(startDir)
	default:
		return "", err
	}

	for {
		configPath := filepath.Join(startDir, defaultConfigFileName)
		fi, statErr := fs.Stat(configPath)
		if statErr == nil {
			mode := fi.Mode()
			if mode.IsRegular() {
				return configPath, nil
			}

			return "", fmt.Errorf(
				"found config at %s but it is not a regular file (mode: %s)",
				configPath,
				mode.String(),
			)
		} else if !os.IsNotExist(statErr) {
			return "", statErr
		}

		parent := filepath.Dir(startDir)
		// now we are at the root of the filesystem, stop searching
		if parent == startDir {
			return "", ErrConfigNotFound
		}

		startDir = parent
	}
}
