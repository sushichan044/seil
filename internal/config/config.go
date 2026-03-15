package config

import (
	"os"
	"path/filepath"

	z "github.com/Oudwins/zog"
	"github.com/goccy/go-yaml"

	"github.com/sushichan044/seil/internal/git"
	"github.com/sushichan044/seil/internal/pathutils"
)

type Config struct {
	PostEdit PostEditHook `yaml:"post_edit"`
	Setup    SetupHook    `yaml:"setup"`
	Teardown TeardownHook `yaml:"teardown"`
}

//nolint:gochecknoglobals // zog schema initialized at package level
var configSchema = z.Struct(z.Shape{
	"postEdit": postEditHookSchema,
	"setup":    setupHookSchema,
	"teardown": teardownHookSchema,
})

type ResolvedConfig struct {
	Config Config
	path   string
}

func (r *ResolvedConfig) CWD() string {
	return filepath.Dir(r.path)
}

// ParseConfigYAML parses bytes and returns a validated Config.
func ParseConfigYAML(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := Validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func FindConfigFile(fromPath string) (string, error) {
	startDir, err := pathutils.DetermineDirectory(fromPath)
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
