package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
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

func FindUpAndLoad(startPath string) (*ResolvedConfig, error) {
	startPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, err
	}

	configPath, err := findConfigPath(startPath)
	if err != nil {
		return nil, err
	}
	if configPath == "" {
		return nil, errors.New("config not found")
	}

	return Load(configPath)
}

func NewConfigFromBytes(data []byte) (*Config, error) {
	var config Config
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func findConfigPath(startDir string) (string, error) {
	startDir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	startDir, err = normalizeSearchStart(startDir)
	if err != nil {
		return "", err
	}

	repoRoot, err := findGitRepoRoot(startDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}

	for dir := startDir; ; dir = filepath.Dir(dir) {
		configPath := filepath.Join(dir, defaultConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}

		if dir == repoRoot {
			return "", nil
		}
	}
}

func normalizeSearchStart(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return path, nil
	}
	return filepath.Dir(path), nil
}

func findGitRepoRoot(startDir string) (string, error) {
	for dir := startDir; ; dir = filepath.Dir(dir) {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
	}
}
