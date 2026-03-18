package config

import (
	"path/filepath"

	"github.com/spf13/afero"
)

const defaultConfigFileName = ".seil.yml"

func Load(fs afero.Fs, path string) (*ResolvedConfig, error) {
	configPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	data, err := afero.ReadFile(fs, configPath)
	if err != nil {
		return nil, err
	}

	config, err := ParseConfigYAML(data)
	if err != nil {
		return nil, err
	}

	configRoot := filepath.Dir(configPath)
	filename := filepath.Base(configPath)

	return &ResolvedConfig{
		Config:   *config,
		basename: filename,
		rootDir:  configRoot,
	}, nil
}
