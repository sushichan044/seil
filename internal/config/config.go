package config

import (
	"path/filepath"

	z "github.com/Oudwins/zog"
	"github.com/goccy/go-yaml"
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

	// basename is the basename of the config file.
	//
	// If config file is exists, this is an actual basename.
	// If config file is not exists, this is a default name.
	basename string

	// rootDir is the root directory to resolve relative paths in the config file.
	rootDir string
}

func NewEmpty(rootDir string) *ResolvedConfig {
	return &ResolvedConfig{
		Config:   Config{},
		rootDir:  rootDir,
		basename: defaultConfigFileName,
	}
}

func (r *ResolvedConfig) RootDir() string {
	return r.rootDir
}

// ConfigPath returns the absolute path to the config file.
//
// Existence of the file is not guaranteed.
func (r *ResolvedConfig) ConfigPath() string {
	return filepath.Join(r.rootDir, r.basename)
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
