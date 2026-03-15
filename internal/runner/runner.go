package runner

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

const (
	maxSummaryLines = 20
	logDirPerm      = 0o700
)

// Runner executes hooks in the given order.
type Runner struct {
	WorkDir string
	Fs      afero.Fs
}

// Hook represents a single hook to execute (name + resolved command string).
type Hook struct {
	Name    string
	Command string
}

// Run executes hooks in the given order and returns their results.
func (r *Runner) Run(ctx context.Context, hooks []Hook) ([]HookResult, error) {
	logDir, err := os.MkdirTemp("", "himo-")
	if err != nil {
		return nil, err
	}
	err = r.Fs.MkdirAll(logDir, logDirPerm)
	if err != nil {
		return nil, err
	}

	results := make([]HookResult, 0, len(hooks))
	for _, hook := range hooks {
		result, hookErr := r.runHook(ctx, hook, logDir)
		if hookErr != nil {
			return nil, hookErr
		}
		results = append(results, result)
	}
	return results, nil
}

func (r *Runner) runHook(ctx context.Context, hook Hook, logDir string) (HookResult, error) {
	logPath := filepath.Join(logDir, hook.Name+".log")
	logFile, err := r.Fs.Create(logPath)
	if err != nil {
		return HookResult{}, err
	}
	defer logFile.Close()

	exitCode, runErr := runCommand(ctx, hook.Command, r.WorkDir, logFile)
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
		Name:     hook.Name,
		Status:   status,
		ExitCode: exitCode,
		LogPath:  logPath,
		Summary:  summary,
	}, nil
}

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
