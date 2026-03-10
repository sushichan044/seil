package config

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	"github.com/sushichan044/himo/internal/git"
	"github.com/sushichan044/himo/internal/pathutils"
)

type (
	Config struct {
		PostEdit PostEdit `yaml:"post_edit"`
	}

	ResolvedConfig struct {
		Config Config
		Path   string
		CWD    string
	}

	PostEdit struct {
		Hooks []EditHook `yaml:"hooks"`
	}

	EditHook struct {
		Glob    string `yaml:"glob"`
		Command string `yaml:"command"`
	}
)

const defaultConfigFileName = "himo.yml"

func Load(path string) (*ResolvedConfig, error) {
	configPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config, err := NewConfigFromBytes(data)
	if err != nil {
		return nil, err
	}

	return &ResolvedConfig{
		Config: *config,
		Path:   configPath,
		CWD:    filepath.Dir(configPath),
	}, nil
}

func NewConfigFromBytes(data []byte) (*Config, error) {
	var config Config
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func ResolveConfigFilePathFrom(path string) (string, error) {
	startDir, err := pathutils.DetermineDirectory(path)
	if err != nil {
		return "", err
	}

	repoRoot, err := git.FindRepoRootFrom(startDir)
	if err != nil {
		return "", err
	}

	for dir := startDir; ; dir = filepath.Dir(dir) {
		configPath := filepath.Join(dir, defaultConfigFileName)
		if _, statErr := os.Stat(configPath); statErr == nil {
			return configPath, nil
		} else if !os.IsNotExist(statErr) {
			return "", statErr
		}

		if dir == repoRoot {
			return "", nil
		}
	}
}
