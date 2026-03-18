package run

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"

	"github.com/sushichan044/seil/internal/config"
)

const (
	logDirPerm    = 0700
	randSuffixLen = 8
)

func Prepare(fs afero.Fs, cfg *config.ResolvedConfig) (*JobRunner, error) {
	if cfg.RootDir() == "" {
		return nil, errors.New("could not determine config root directory")
	}

	var logRoot string
	var customDir bool
	if dir := cfg.LogDir(); dir != "" {
		if err := fs.MkdirAll(dir, logDirPerm); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		logRoot = dir
		customDir = true
	} else {
		var err error
		logRoot, err = afero.TempDir(fs, "", "seil-logs")
		if err != nil {
			return nil, err
		}
	}
	return &JobRunner{fs: fs, cfg: cfg, logRoot: logRoot, customDir: customDir}, nil
}

type JobRunner struct {
	fs        afero.Fs
	cfg       *config.ResolvedConfig
	logRoot   string
	customDir bool
}

func (r *JobRunner) logFileForJob(hookType string, job *config.Job) (afero.File, error) {
	var filename string
	if r.customDir {
		b := make([]byte, randSuffixLen)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("failed to generate log filename: %w", err)
		}
		filename = fmt.Sprintf("%s-%s-%s.log", hookType, job.PathSafeName(), hex.EncodeToString(b))
	} else {
		filename = hookType + "-" + job.PathSafeName() + ".log"
	}
	logFile, err := r.fs.Create(filepath.Join(r.logRoot, filename))
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func (r *JobRunner) runJobs(ctx context.Context, hookType string, jobs []config.Job) []Result {
	results := make([]Result, len(jobs))
	var wg sync.WaitGroup

	for i, job := range jobs {
		wg.Go(func() {
			cmd, err := EvalJob(job.Run, JobEvaluationParams{})
			if err != nil {
				results[i] = Failure(job.DisplayName(), "", err)
				return
			}
			logFile, err := r.logFileForJob(hookType, &job)
			if err != nil {
				results[i] = Failure(job.DisplayName(), "", err)
				return
			}
			defer logFile.Close()

			proc := exec.CommandContext(ctx, "sh", "-c", cmd)
			proc.Dir = r.cfg.RootDir()
			proc.Stdout = logFile
			proc.Stderr = logFile

			if execErr := proc.Run(); execErr == nil {
				results[i] = Success(job.DisplayName(), logFile.Name())
			} else {
				results[i] = Failure(job.DisplayName(), logFile.Name(), execErr)
			}
		})
	}
	wg.Wait()
	return results
}

func (r *JobRunner) RunSetup(ctx context.Context) ([]Result, error) {
	return r.runJobs(ctx, "setup", r.cfg.Config.Setup.Jobs), nil
}

func (r *JobRunner) RunTeardown(ctx context.Context) ([]Result, error) {
	return r.runJobs(ctx, "teardown", r.cfg.Config.Teardown.Jobs), nil
}

// RunPostEdit executes the given pre-filtered jobs for the edited file.
// Callers are responsible for filtering jobs (via GlobJob.Matches + gitignore) before calling.
func (r *JobRunner) RunPostEdit(
	ctx context.Context,
	wsPath config.WorkspacePath,
	jobs []config.GlobJob,
) ([]Result, error) {
	results := make([]Result, len(jobs))
	var wg sync.WaitGroup

	for i, job := range jobs {
		wg.Go(func() {
			cmd, err := EvalJob(job.Run, JobEvaluationParams{Filepath: wsPath})
			if err != nil {
				results[i] = Failure(job.DisplayName(), "", err)
				return
			}
			logFile, err := r.logFileForJob("post-edit", &job.Job)
			if err != nil {
				results[i] = Failure(job.DisplayName(), "", err)
				return
			}
			defer logFile.Close()

			proc := exec.CommandContext(ctx, "sh", "-c", cmd)
			proc.Dir = r.cfg.RootDir()
			proc.Stdout = logFile
			proc.Stderr = logFile

			if execErr := proc.Run(); execErr == nil {
				results[i] = Success(job.DisplayName(), logFile.Name())
			} else {
				results[i] = Failure(job.DisplayName(), logFile.Name(), execErr)
			}
		})
	}
	wg.Wait()
	return results, nil
}
