package config

import (
	"errors"
	"fmt"
	"strings"

	z "github.com/Oudwins/zog"
)

var filePatternHookSchema = z.Struct(z.Shape{ //nolint:gochecknoglobals // zog schema initialized at package level
	"glob":    z.String(),
	"command": z.String().Required().Min(1),
})

var filePatternHooksSchema = z.EXPERIMENTAL_MAP[string, FilePatternHook]( //nolint:gochecknoglobals // zog schema initialized at package level
	z.String(),
	filePatternHookSchema,
)

var simpleHookSchema = z.Struct(z.Shape{ //nolint:gochecknoglobals // zog schema initialized at package level
	"command": z.String().Required().Min(1),
})

var simpleHooksSchema = z.EXPERIMENTAL_MAP[string, SimpleHook]( //nolint:gochecknoglobals // zog schema initialized at package level
	z.String(),
	simpleHookSchema,
)

var configSchema = z.Struct(z.Shape{ //nolint:gochecknoglobals // zog schema initialized at package level
	"postEdit": z.Struct(z.Shape{
		"hooks": filePatternHooksSchema,
	}),
	"setup": z.Struct(z.Shape{
		"hooks": simpleHooksSchema,
	}),
	"teardown": z.Struct(z.Shape{
		"hooks": simpleHooksSchema,
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
