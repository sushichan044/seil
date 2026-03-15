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
