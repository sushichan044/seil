package seil

import (
	"context"
	"fmt"

	"github.com/spf13/afero"

	"github.com/sushichan044/seil/internal/config"
	"github.com/sushichan044/seil/internal/gitignore"
	"github.com/sushichan044/seil/internal/run"
)

func runPostEditHooks(
	ctx context.Context,
	cfg *config.ResolvedConfig,
	gitignoreMatcher *gitignore.Matcher,
	fs afero.Fs,
	wsPath config.WorkspacePath,
) ([]run.Result, error) {
	jobs := cfg.Config.PostEdit.Jobs
	results := make([]run.Result, len(jobs))
	var toRun []config.GlobJob
	var toRunIdx []int

	for i, job := range jobs {
		if !job.Matches(wsPath) {
			results[i] = run.Skipped(job.DisplayName(), run.SkipReason{
				Code:    run.SkipReasonGlobNoMatch,
				Message: fmt.Sprintf("glob pattern %q did not match", job.Glob),
			})
			continue
		}
		if gitignoreMatcher != nil && gitignoreMatcher.IsIgnored(wsPath.Rel()) {
			results[i] = run.Skipped(job.DisplayName(), run.SkipReason{
				Code:    run.SkipReasonGitignored,
				Message: fmt.Sprintf("file %q is gitignored", wsPath.Rel()),
			})
			continue
		}
		toRun = append(toRun, job)
		toRunIdx = append(toRunIdx, i)
	}

	if len(toRun) == 0 {
		return results, nil
	}

	r, err := run.Prepare(fs, cfg)
	if err != nil {
		return nil, err
	}

	runResults, err := r.RunPostEdit(ctx, wsPath, toRun)
	if err != nil {
		return nil, err
	}

	for k, idx := range toRunIdx {
		results[idx] = runResults[k]
	}
	return results, nil
}
