package config

import (
	"regexp"

	z "github.com/Oudwins/zog"
)

type Job struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

//nolint:gochecknoglobals // zog schema initialized at package level
var jobSchema = z.Struct(z.Shape{
	"name": z.String(),
	"run":  z.String().Required().Min(1),
})

type GlobJob struct {
	Job `yaml:",inline"`

	Glob string `yaml:"glob"`
}

//nolint:gochecknoglobals // zog schema initialized at package level
var globJobSchema = jobSchema.Extend(z.Shape{
	"glob": z.String(),
})

var unsafePathRe = regexp.MustCompile(`[^0-9A-Za-z_.\-]+`)

func (job *Job) DisplayName() string {
	if job.Name != "" {
		return job.Name
	}

	return job.Run
}

func (job *Job) PathSafeName() string {
	var base string
	if job.Name != "" {
		base = job.Name
	} else {
		base = job.displayableRun()
	}

	return unsafePathRe.ReplaceAllString(base, "_")
}

func (job *Job) displayableRun() string {
	if job.Run == "" {
		return "<no command>"
	}
	return job.Run
}
