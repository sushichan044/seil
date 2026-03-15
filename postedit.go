package himo

import (
	"context"
	"path/filepath"
	"sort"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/afero"

	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/gitignore"
	"github.com/sushichan044/himo/internal/runner"
	"github.com/sushichan044/himo/internal/template"
)

func runPostEditHooks(
	ctx context.Context,
	cfg *config.ResolvedConfig,
	gitignoreMatcher *gitignore.Matcher,
	fs afero.Fs,
	filePath string,
) ([]runner.HookResult, error) {
	hooks := cfg.Config.PostEdit.Hooks
	names := sortedKeys(hooks)

	var toRun []runner.Hook
	skipped := map[string]bool{}

	for _, name := range names {
		hook := hooks[name]
		if hook.Glob != "" {
			normalized := filepath.ToSlash(filePath)
			matched, err := doublestar.Match(hook.Glob, normalized)
			if err != nil || !matched {
				skipped[name] = true
				continue
			}
		}
		if gitignoreMatcher.IsIgnored(filePath) {
			skipped[name] = true
			continue
		}
		cmd, err := template.EvalCommand(hook.Command, template.CommandVars{Files: []string{filePath}})
		if err != nil {
			return nil, err
		}
		toRun = append(toRun, runner.Hook{Name: name, Command: cmd})
	}

	var runResultMap map[string]runner.HookResult
	if len(toRun) > 0 {
		r := &runner.Runner{WorkDir: cfg.CWD, Fs: fs}
		runResults, err := r.Run(ctx, toRun)
		if err != nil {
			return nil, err
		}
		runResultMap = make(map[string]runner.HookResult, len(runResults))
		for _, res := range runResults {
			runResultMap[res.Name] = res
		}
	}

	results := make([]runner.HookResult, 0, len(names))
	for _, name := range names {
		if skipped[name] {
			results = append(results, runner.HookResult{Name: name, Status: runner.HookStatusSkipped})
		} else {
			results = append(results, runResultMap[name])
		}
	}
	return results, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
