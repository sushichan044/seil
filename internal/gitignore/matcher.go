package gitignore

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	gogitignore "github.com/sabhiram/go-gitignore"
)

// scopedIgnore pairs a .gitignore with the directory it was found in (relative to root).
type scopedIgnore struct {
	// dir is the directory containing the .gitignore, relative to root (e.g. "", "sub", "a/b").
	dir      string
	compiled *gogitignore.GitIgnore
}

// Matcher checks whether a file path is gitignored.
type Matcher struct {
	scopes []scopedIgnore
}

// NewMatcherFromRoot collects all .gitignore files under root and compiles each one
// with its directory scope preserved. fs is used for file system access to enable testing.
func NewMatcherFromRoot(fs afero.Fs, root string) (*Matcher, error) {
	var scopes []scopedIgnore

	err := afero.Walk(fs, root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Base(path) != ".gitignore" {
			return nil
		}
		data, readErr := afero.ReadFile(fs, path)
		if readErr != nil {
			return readErr
		}
		lines := splitLines(string(data))
		compiled := gogitignore.CompileIgnoreLines(lines...)

		relDir, relErr := filepath.Rel(root, filepath.Dir(path))
		if relErr != nil {
			return relErr
		}
		// Normalize to forward slashes for consistent matching.
		relDir = filepath.ToSlash(relDir)
		if relDir == "." {
			relDir = ""
		}
		scopes = append(scopes, scopedIgnore{dir: relDir, compiled: compiled})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Matcher{scopes: scopes}, nil
}

// IsIgnored returns true if the given path is matched by any scoped gitignore.
// path should be relative to the workspace root.
func (m *Matcher) IsIgnored(path string) bool {
	path = filepath.ToSlash(path)
	for _, s := range m.scopes {
		rel := path
		if s.dir != "" {
			if !strings.HasPrefix(path, s.dir+"/") {
				continue
			}
			rel = path[len(s.dir)+1:]
		}
		if s.compiled.MatchesPath(rel) {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := range len(s) {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
