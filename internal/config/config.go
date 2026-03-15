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
		Setup    Setup    `yaml:"setup"`
		Teardown Teardown `yaml:"teardown"`
	}

	ResolvedConfig struct {
		Config Config
		Path   string
		CWD    string
	}

	PostEdit struct {
		Hooks map[string]FilePatternHook `yaml:"hooks"`
	}

	FilePatternHook struct {
		Glob    string `yaml:"glob"`
		Command string `yaml:"command"`
	}

	SimpleHook struct {
		Command string `yaml:"command"`
	}

	Setup struct {
		Hooks map[string]SimpleHook `yaml:"hooks"`
	}

	Teardown struct {
		Hooks map[string]SimpleHook `yaml:"hooks"`
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

	config, err := newConfigFromBytes(data)
	if err != nil {
		return nil, err
	}

	return &ResolvedConfig{
		Config: *config,
		Path:   configPath,
		CWD:    filepath.Dir(configPath),
	}, nil
}

func newConfigFromBytes(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := Validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
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
