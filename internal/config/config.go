package config

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/goccy/go-yaml"

	"github.com/sushichan044/himo/internal/git"
	"github.com/sushichan044/himo/internal/pathutils"
)

const hookNameTruncateLen = 32

var unsafeNameRe = regexp.MustCompile(`[^0-9A-Za-z_.\-]+`)

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
		Jobs []FilePatternHook `yaml:"jobs"`
	}

	FilePatternHook struct {
		Name string `yaml:"name"`
		Glob string `yaml:"glob"`
		Run  string `yaml:"run"`
	}

	SimpleHook struct {
		Name string `yaml:"name"`
		Run  string `yaml:"run"`
	}

	Setup struct {
		Jobs []SimpleHook `yaml:"jobs"`
	}

	Teardown struct {
		Jobs []SimpleHook `yaml:"jobs"`
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

	config, err := ParseConfig(data)
	if err != nil {
		return nil, err
	}

	return &ResolvedConfig{
		Config: *config,
		Path:   configPath,
		CWD:    filepath.Dir(configPath),
	}, nil
}

// ParseConfig parses YAML bytes and returns a normalized, validated Config.
func ParseConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := Validate(&cfg); err != nil {
		return nil, err
	}

	normalizeConfig(&cfg)
	return &cfg, nil
}

func normalizeConfig(cfg *Config) {
	for i := range cfg.PostEdit.Jobs {
		cfg.PostEdit.Jobs[i].Name = normalizeHookName(cfg.PostEdit.Jobs[i].Name, cfg.PostEdit.Jobs[i].Run)
	}
	for i := range cfg.Setup.Jobs {
		cfg.Setup.Jobs[i].Name = normalizeHookName(cfg.Setup.Jobs[i].Name, cfg.Setup.Jobs[i].Run)
	}
	for i := range cfg.Teardown.Jobs {
		cfg.Teardown.Jobs[i].Name = normalizeHookName(cfg.Teardown.Jobs[i].Name, cfg.Teardown.Jobs[i].Run)
	}
}

func normalizeHookName(name, run string) string {
	if name == "" {
		r := []rune(run)
		if len(r) > hookNameTruncateLen {
			name = string(r[:hookNameTruncateLen])
		} else {
			name = run
		}
	}
	return unsafeNameRe.ReplaceAllString(name, "-")
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
