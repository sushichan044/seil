package config

import z "github.com/Oudwins/zog"

type (
	PostEditHook struct {
		Jobs []GlobJob `yaml:"jobs"`
	}

	SetupHook struct {
		Jobs []Job `yaml:"jobs"`
	}

	TeardownHook struct {
		Jobs []Job `yaml:"jobs"`
	}
)

//nolint:gochecknoglobals // zog schema initialized at package level
var postEditHookSchema = z.Struct(z.Shape{
	"jobs": z.Slice(globJobSchema),
})

//nolint:gochecknoglobals // zog schema initialized at package level
var setupHookSchema = z.Struct(z.Shape{
	"jobs": z.Slice(jobSchema),
})

//nolint:gochecknoglobals // zog schema initialized at package level
var teardownHookSchema = z.Struct(z.Shape{
	"jobs": z.Slice(jobSchema),
})
