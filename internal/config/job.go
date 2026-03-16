package config

import (
	"path/filepath"
	"regexp"

	z "github.com/Oudwins/zog"
	"github.com/bmatcuk/doublestar/v4"
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

// Matches reports whether this job should run for the given filePath.
// filePath and configRoot should both be absolute paths.
// If Glob is empty, the job always matches.
func (j *GlobJob) Matches(filePath, configRoot string) bool {
	if j.Glob == "" {
		return true
	}
	rel, err := filepath.Rel(configRoot, filePath)
	if err != nil {
		rel = filePath
	}
	normalized := filepath.ToSlash(rel)
	matched, err := doublestar.Match(j.Glob, normalized)
	return err == nil && matched
}
