package seil

import (
	"context"

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
	filePath string,
) ([]run.Result, error) {
	jobs := cfg.Config.PostEdit.Jobs
	results := make([]run.Result, len(jobs))
	var toRun []config.GlobJob
	var toRunIdx []int

	for i, job := range jobs {
		if !job.Matches(filePath, cfg.RootDir()) || gitignoreMatcher.IsIgnored(filePath) {
			results[i] = run.Skipped(job.DisplayName())
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

	runResults, err := r.RunPostEdit(ctx, filePath, toRun)
	if err != nil {
		return nil, err
	}

	for k, idx := range toRunIdx {
		results[idx] = runResults[k]
	}
	return results, nil
}
