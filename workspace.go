package himo

import (
	"context"

	"github.com/spf13/afero"

	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/gitignore"
	"github.com/sushichan044/himo/internal/postedit"
)

// Workspace provides the public API for himo operations.
type Workspace struct {
	config    *config.ResolvedConfig
	gitignore *gitignore.Matcher
	fs        afero.Fs
}

// NewWorkspace creates a Workspace from the given resolved config.
func NewWorkspace(cfg *config.ResolvedConfig) (*Workspace, error) {
	fs := afero.NewOsFs()
	m, err := gitignore.NewMatcherFromRoot(fs, cfg.CWD)
	if err != nil {
		return nil, err
	}
	return &Workspace{config: cfg, gitignore: m, fs: fs}, nil
}

// RunPostEditHooks executes all post-edit hooks for the given file path.
func (w *Workspace) RunPostEditHooks(ctx context.Context, filePath string) ([]postedit.HookResult, error) {
	r := &postedit.Runner{
		Config:    w.config,
		Gitignore: w.gitignore,
		Fs:        w.fs,
	}
	return r.Run(ctx, filePath)
}
