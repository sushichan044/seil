package seil

import (
	"context"

	"github.com/spf13/afero"

	"github.com/sushichan044/seil/internal/config"
	"github.com/sushichan044/seil/internal/run"
)

func runTeardownHooks(
	ctx context.Context,
	cfg *config.ResolvedConfig,
	fs afero.Fs,
) ([]run.Result, error) {
	if len(cfg.Config.Teardown.Jobs) == 0 {
		return []run.Result{}, nil
	}
	r, err := run.Prepare(fs, cfg)
	if err != nil {
		return nil, err
	}
	return r.RunTeardown(ctx)
}
