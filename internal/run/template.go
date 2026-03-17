package run

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sushichan044/seil/internal/config"
)

//nolint:gochecknoglobals // funcMap is a static map of template functions
var funcMap = template.FuncMap{
	"dir":  filepath.Dir,
	"base": filepath.Base,
	"ext":  filepath.Ext,
}

// vars holds template variables available during hook command evaluation.
type vars struct {
	File string
}

type JobEvaluationParams struct {
	Filepath config.WorkspacePath
}

// EvalJob evaluates a Go template string with the given JobEvaluationParams.
func EvalJob(tmpl string, params JobEvaluationParams) (string, error) {
	v := vars{
		File: params.Filepath.Rel(),
	}

	t, err := template.New("").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err = t.Execute(&sb, v); err != nil {
		return "", err
	}
	return sb.String(), nil
}
