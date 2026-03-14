package config

import (
	"errors"
	"fmt"
	"strings"

	z "github.com/Oudwins/zog"
)

var hookSchema = z.Struct(z.Shape{ //nolint:gochecknoglobals // zog schema initialized at package level
	"glob":    z.String().Required().Min(1),
	"command": z.String().Required().Min(1),
})

var hooksSchema = z.EXPERIMENTAL_MAP[string, Hook]( //nolint:gochecknoglobals // zog schema initialized at package level
	z.String(),
	hookSchema,
)

// Validate validates a Config using zog schemas and returns a combined error.
func Validate(cfg *Config) error {
	issues := hooksSchema.Validate(&cfg.PostEdit.Hooks)
	if len(issues) == 0 {
		return nil
	}

	errs := make([]error, 0, len(issues))
	for _, issue := range issues {
		errs = append(errs, fmt.Errorf("hooks%s: %s", formatPath(issue.Path), issue.Message))
	}
	return errors.Join(errs...)
}

func formatPath(path []string) string {
	return strings.Join(path, "")
}
