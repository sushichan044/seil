package run

import (
	"path/filepath"
	"strings"
	"text/template"
)

//nolint:gochecknoglobals // funcMap is a static map of template functions
var funcMap = template.FuncMap{
	"dir":  filepath.Dir,
	"base": filepath.Base,
	"ext":  filepath.Ext,
}

// Vars holds template variables available during hook command evaluation.
type Vars struct {
	File string
}

// EvalJob evaluates a Go template string with the given Vars.
func EvalJob(tmpl string, vars Vars) (string, error) {
	t, err := template.New("").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err = t.Execute(&sb, vars); err != nil {
		return "", err
	}
	return sb.String(), nil
}
