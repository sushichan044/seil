package himo

import (
	"context"

	"github.com/spf13/afero"

	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/runner"
	"github.com/sushichan044/himo/internal/template"
)

func runSetupHooks(
	ctx context.Context,
	cfg *config.ResolvedConfig,
	fs afero.Fs,
) ([]runner.HookResult, error) {
	return runSimpleHooks(ctx, cfg.CWD, fs, cfg.Config.Setup.Jobs)
}

func runSimpleHooks(
	ctx context.Context,
	workDir string,
	fs afero.Fs,
	jobs []config.SimpleHook,
) ([]runner.HookResult, error) {
	toRun := make([]runner.Hook, 0, len(jobs))
	for _, job := range jobs {
		cmd, err := template.EvalCommand(job.Run, template.CommandVars{Files: []string{}})
		if err != nil {
			return nil, err
		}
		toRun = append(toRun, runner.Hook{Name: job.Name, Command: cmd})
	}

	if len(toRun) == 0 {
		return []runner.HookResult{}, nil
	}

	r := &runner.Runner{WorkDir: workDir, Fs: fs}
	return r.Run(ctx, toRun)
}
