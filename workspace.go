package seil

import (
	"context"
	"errors"

	"github.com/spf13/afero"

	"github.com/sushichan044/seil/internal/config"
	"github.com/sushichan044/seil/internal/gitignore"
	"github.com/sushichan044/seil/internal/run"
)

// Workspace provides the public API for seil operations.
type Workspace struct {
	config *config.ResolvedConfig
	fs     afero.Fs
}

// NewWorkspace creates a Workspace from the given resolved config.
func NewWorkspace(cfg *config.ResolvedConfig) (*Workspace, error) {
	return &Workspace{config: cfg, fs: afero.NewOsFs()}, nil
}

// RunPostEditHooks executes all post-edit hooks for the given file path.
func (w *Workspace) RunPostEditHooks(ctx context.Context, filePath string) ([]run.Result, error) {
	if len(w.config.Config.PostEdit.Jobs) == 0 {
		return []run.Result{}, nil
	}

	m, err := gitignore.NewMatcherFromRoot(w.fs, w.config.RootDir())
	if err != nil {
		return nil, err
	}
	wsPath, err := config.NewWorkspacePath(w.config.RootDir(), filePath)
	if err != nil {
		if outsideErr, ok := errors.AsType[*config.OutsideWorkspaceError](err); ok {
			jobs := w.config.Config.PostEdit.Jobs
			results := make([]run.Result, len(jobs))
			for i, job := range jobs {
				results[i] = run.Skipped(job.DisplayName(), run.SkipReason{
					Code:    run.SkipReasonOutsideWorkspace,
					Message: outsideErr.Error(),
				})
			}
			return results, nil
		}
		return nil, err
	}
	return runPostEditHooks(ctx, w.config, m, w.fs, wsPath)
}

// RunSetupHooks executes all setup hooks.
func (w *Workspace) RunSetupHooks(ctx context.Context) ([]run.Result, error) {
	if len(w.config.Config.Setup.Jobs) == 0 {
		return []run.Result{}, nil
	}

	return runSetupHooks(ctx, w.config, w.fs)
}

// RunTeardownHooks executes all teardown hooks.
func (w *Workspace) RunTeardownHooks(ctx context.Context) ([]run.Result, error) {
	if len(w.config.Config.Teardown.Jobs) == 0 {
		return []run.Result{}, nil
	}

	return runTeardownHooks(ctx, w.config, w.fs)
}
