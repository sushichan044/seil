package himo

import (
	"context"

	"github.com/spf13/afero"

	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/gitignore"
	"github.com/sushichan044/himo/internal/runner"
)

// Workspace provides the public API for himo operations.
type Workspace struct {
	config *config.ResolvedConfig
	fs     afero.Fs
}

// NewWorkspace creates a Workspace from the given resolved config.
func NewWorkspace(cfg *config.ResolvedConfig) (*Workspace, error) {
	return &Workspace{config: cfg, fs: afero.NewOsFs()}, nil
}

// RunPostEditHooks executes all post-edit hooks for the given file path.
func (w *Workspace) RunPostEditHooks(ctx context.Context, filePath string) ([]runner.HookResult, error) {
	m, err := gitignore.NewMatcherFromRoot(w.fs, w.config.CWD)
	if err != nil {
		return nil, err
	}
	return runPostEditHooks(ctx, w.config, m, w.fs, filePath)
}

// RunSetupHooks executes all setup hooks.
func (w *Workspace) RunSetupHooks(ctx context.Context) ([]runner.HookResult, error) {
	return runSetupHooks(ctx, w.config, w.fs)
}

// RunTeardownHooks executes all teardown hooks.
func (w *Workspace) RunTeardownHooks(ctx context.Context) ([]runner.HookResult, error) {
	return runTeardownHooks(ctx, w.config, w.fs)
}
