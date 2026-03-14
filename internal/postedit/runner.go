package postedit

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/afero"

	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/gitignore"
	"github.com/sushichan044/himo/internal/template"
)

const (
	maxSummaryLines = 20
	logDirPerm      = 0o700
)

// Runner executes post-edit hooks for a given file path.
type Runner struct {
	Config    *config.ResolvedConfig
	Gitignore *gitignore.Matcher
	Fs        afero.Fs
}

// Run executes all configured hooks for the given filePath and returns their results.
// Hooks are processed in alphabetical order by name.
func (r *Runner) Run(ctx context.Context, filePath string) ([]HookResult, error) {
	hooks := r.Config.Config.PostEdit.Hooks
	names := make([]string, 0, len(hooks))
	for name := range hooks {
		names = append(names, name)
	}
	sort.Strings(names)

	logDir, err := os.MkdirTemp("", "himo-")
	if err != nil {
		return nil, err
	}
	err = r.Fs.MkdirAll(logDir, logDirPerm)
	if err != nil {
		return nil, err
	}

	results := make([]HookResult, 0, len(names))
	for _, name := range names {
		hook := hooks[name]
		result, hookErr := r.runHook(ctx, name, hook, filePath, logDir)
		if hookErr != nil {
			return nil, hookErr
		}
		results = append(results, result)
	}
	return results, nil
}

func (r *Runner) runHook(
	ctx context.Context,
	name string,
	hook config.Hook,
	filePath, logDir string,
) (HookResult, error) {
	normalized := filepath.ToSlash(filePath)
	matched, err := doublestar.Match(hook.Glob, normalized)
	if err != nil {
		return HookResult{Name: name, Status: HookStatusSkipped}, nil
	}
	if !matched || r.Gitignore.IsIgnored(filePath) {
		return HookResult{Name: name, Status: HookStatusSkipped}, nil
	}

	cmd, err := template.EvalCommand(hook.Command, template.CommandVars{Files: []string{filePath}})
	if err != nil {
		return HookResult{}, err
	}

	logPath := filepath.Join(logDir, name+".log")
	logFile, err := r.Fs.Create(logPath)
	if err != nil {
		return HookResult{}, err
	}
	defer logFile.Close()

	exitCode, runErr := runCommand(ctx, cmd, r.Config.CWD, logFile)
	if runErr != nil && exitCode < 0 {
		return HookResult{}, runErr
	}

	summary, err := readFirstNLines(r.Fs, logPath, maxSummaryLines)
	if err != nil {
		return HookResult{}, err
	}

	status := HookStatusSuccess
	if exitCode != 0 {
		status = HookStatusFailure
	}

	return HookResult{
		Name:     name,
		Status:   status,
		ExitCode: exitCode,
		LogPath:  logPath,
		Summary:  summary,
	}, nil
}

// runCommand executes a shell command and writes combined stdout+stderr to w.
// Returns the exit code and any execution error.
func runCommand(ctx context.Context, cmd, cwd string, w afero.File) (int, error) {
	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	c.Dir = cwd
	c.Stdout = w
	c.Stderr = w

	if err := c.Run(); err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return -1, err
	}
	return 0, nil
}

func readFirstNLines(fs afero.Fs, path string, n int) (string, error) {
	f, err := fs.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() && len(lines) < n {
		lines = append(lines, scanner.Text())
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return "", scanErr
	}
	return strings.Join(lines, "\n"), nil
}
