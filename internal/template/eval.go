package template

import (
	"bytes"
	"strings"
	"text/template"
)

// ConfigVars holds template variables available during config file evaluation.
// Reserved for future extension.
type ConfigVars struct{}

// CommandVars holds template variables available during hook command evaluation.
type CommandVars struct {
	Files []string
}

var (
	//nolint:gochecknoglobals // define as global for efficiency
	fnMap = template.FuncMap{
		// join takes (sep, elems) so it works with template pipe: {{.Files | join " "}}
		"join": func(sep string, elems []string) string {
			return strings.Join(elems, sep)
		},
	}
)

// EvalConfig evaluates a Go template string with ConfigVars.
func EvalConfig(tmpl string, vars ConfigVars) (string, error) {
	return render(tmpl, vars)
}

// EvalCommand evaluates a Go template string with CommandVars.
func EvalCommand(tmpl string, vars CommandVars) (string, error) {
	return render(tmpl, vars)
}

func render(tmpl string, data any) (string, error) {
	t, err := template.New("").Funcs(fnMap).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
