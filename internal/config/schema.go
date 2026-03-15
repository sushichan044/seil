package config

import (
	"errors"
	"fmt"
	"strings"

	z "github.com/Oudwins/zog"
)

var filePatternHookSchema = z.Struct(z.Shape{ //nolint:gochecknoglobals // zog schema initialized at package level
	"name": z.String(),
	"glob": z.String(),
	"run":  z.String().Required().Min(1),
})

var filePatternHooksSchema = z.Slice( //nolint:gochecknoglobals // zog schema initialized at package level
	filePatternHookSchema,
)

var simpleHookSchema = z.Struct(z.Shape{ //nolint:gochecknoglobals // zog schema initialized at package level
	"name": z.String(),
	"run":  z.String().Required().Min(1),
})

var simpleHooksSchema = z.Slice(simpleHookSchema) //nolint:gochecknoglobals // zog schema initialized at package level

var configSchema = z.Struct(z.Shape{ //nolint:gochecknoglobals // zog schema initialized at package level
	"postEdit": z.Struct(z.Shape{
		"jobs": filePatternHooksSchema,
	}),
	"setup": z.Struct(z.Shape{
		"jobs": simpleHooksSchema,
	}),
	"teardown": z.Struct(z.Shape{
		"jobs": simpleHooksSchema,
	}),
})

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
