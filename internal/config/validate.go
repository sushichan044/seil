package config

import (
	"errors"
	"fmt"
	"strings"
)

func Validate(cfg *Config) error {
	issues := configSchema.Validate(cfg)
	var errs []error
	for _, issue := range issues {
		errs = append(errs, fmt.Errorf("%s: %s", formatPath(issue.Path), issue.Message))
	}
	return errors.Join(errs...)
}

func formatPath(path []string) string {
	return strings.Join(path, ".")
}
