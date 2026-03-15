package himo

import (
	"context"

	"github.com/spf13/afero"

	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/runner"
)

func runTeardownHooks(
	ctx context.Context,
	cfg *config.ResolvedConfig,
	fs afero.Fs,
) ([]runner.HookResult, error) {
	return runSimpleHooks(ctx, cfg.CWD, fs, cfg.Config.Teardown.Jobs)
}
