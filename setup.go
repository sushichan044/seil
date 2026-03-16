package seil

import (
	"context"

	"github.com/spf13/afero"

	"github.com/sushichan044/seil/internal/config"
	"github.com/sushichan044/seil/internal/run"
)

func runSetupHooks(
	ctx context.Context,
	cfg *config.ResolvedConfig,
	fs afero.Fs,
) ([]run.Result, error) {
	if len(cfg.Config.Setup.Jobs) == 0 {
		return []run.Result{}, nil
	}
	r, err := run.Prepare(fs, cfg)
	if err != nil {
		return nil, err
	}
	return r.RunSetup(ctx)
}
