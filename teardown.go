package himo

import (
	"context"

	"github.com/spf13/afero"

	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/runner"
	"github.com/sushichan044/himo/internal/template"
)

func runTeardownHooks(
	ctx context.Context,
	cfg *config.ResolvedConfig,
	fs afero.Fs,
) ([]runner.HookResult, error) {
	hooks := cfg.Config.Teardown.Hooks
	names := sortedKeys(hooks)

	toRun := make([]runner.Hook, 0, len(names))
	for _, name := range names {
		hook := hooks[name]
		cmd, err := template.EvalCommand(hook.Command, template.CommandVars{Files: []string{}})
		if err != nil {
			return nil, err
		}
		toRun = append(toRun, runner.Hook{Name: name, Command: cmd})
	}

	if len(toRun) == 0 {
		return []runner.HookResult{}, nil
	}

	r := &runner.Runner{WorkDir: cfg.CWD, Fs: fs}
	return r.Run(ctx, toRun)
}
