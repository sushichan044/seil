package seil

import (
	"context"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/afero"

	"github.com/sushichan044/seil/internal/config"
	"github.com/sushichan044/seil/internal/gitignore"
	"github.com/sushichan044/seil/internal/runner"
	"github.com/sushichan044/seil/internal/template"
)

func runPostEditHooks(
	ctx context.Context,
	cfg *config.ResolvedConfig,
	gitignoreMatcher *gitignore.Matcher,
	fs afero.Fs,
	filePath string,
) ([]runner.HookResult, error) {
	jobs := cfg.Config.PostEdit.Jobs

	skipped := make([]bool, len(jobs))
	jobNames := make([]string, len(jobs))
	for i, job := range jobs {
		jobNames[i] = job.Name
	}
	var toRun []runner.Hook

	for i, job := range jobs {
		if isPostEditJobSkipped(job, filePath, gitignoreMatcher) {
			skipped[i] = true
			continue
		}
		cmd, err := template.EvalCommand(job.Run, template.CommandVars{Files: []string{filePath}})
		if err != nil {
			return nil, err
		}
		toRun = append(toRun, runner.Hook{Name: job.Name, Command: cmd})
	}

	runResultMap, err := executeHooks(ctx, cfg.CWD, fs, toRun)
	if err != nil {
		return nil, err
	}

	results := make([]runner.HookResult, 0, len(jobs))
	for i := range jobs {
		if skipped[i] {
			results = append(results, runner.HookResult{Name: jobNames[i], Status: runner.HookStatusSkipped})
		} else {
			results = append(results, runResultMap[jobNames[i]])
		}
	}
	return results, nil
}

func isPostEditJobSkipped(job config.FilePatternHook, filePath string, matcher *gitignore.Matcher) bool {
	if job.Glob != "" {
		normalized := filepath.ToSlash(filePath)
		matched, err := doublestar.Match(job.Glob, normalized)
		if err != nil || !matched {
			return true
		}
	}
	return matcher.IsIgnored(filePath)
}

func executeHooks(
	ctx context.Context,
	workDir string,
	fs afero.Fs,
	toRun []runner.Hook,
) (map[string]runner.HookResult, error) {
	if len(toRun) == 0 {
		return map[string]runner.HookResult{}, nil
	}
	r := &runner.Runner{WorkDir: workDir, Fs: fs}
	runResults, err := r.Run(ctx, toRun)
	if err != nil {
		return nil, err
	}
	resultMap := make(map[string]runner.HookResult, len(runResults))
	for _, res := range runResults {
		resultMap[res.Name] = res
	}
	return resultMap, nil
}
