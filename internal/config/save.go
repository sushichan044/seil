package config

import (
	"github.com/spf13/afero"
)

func SaveDefault(fs afero.Fs, path string) error {
	return afero.WriteFile(fs, path, defaultConfigYml, 0o600)
}
