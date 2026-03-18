package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func Validate(cfg *Config) error {
	issues := configSchema.Validate(cfg)
	var errs []error
	for _, issue := range issues {
		errs = append(errs, fmt.Errorf("%s: %s", formatPath(issue.Path), issue.Message))
	}

	if cfg.LogDir != "" {
		if err := validateLogDir(cfg.LogDir); err != nil {
			errs = append(errs, fmt.Errorf("log_dir: %w", err))
		}
	}

	return errors.Join(errs...)
}

func validateLogDir(logDir string) error {
	if filepath.IsAbs(logDir) {
		return errors.New("must be a relative path")
	}
	cleaned := filepath.Clean(logDir)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return errors.New("must not escape config root")
	}
	return nil
}

func formatPath(path []string) string {
	return strings.Join(path, ".")
}
