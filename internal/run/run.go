package run

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"

	"github.com/sushichan044/seil/internal/config"
)

func Prepare(fs afero.Fs, cfg *config.ResolvedConfig) (*JobRunner, error) {
	if cfg.RootDir() == "" {
		return nil, errors.New("could not determine config root directory")
	}
	logRoot, err := afero.TempDir(fs, "", "seil-logs")
	if err != nil {
		return nil, err
	}
	return &JobRunner{fs, cfg, logRoot}, nil
}

type JobRunner struct {
	fs      afero.Fs
	cfg     *config.ResolvedConfig
	logRoot string
}

func (r *JobRunner) logFileForJob(job *config.Job) (afero.File, error) {
	log := filepath.Join(r.logRoot, "setup-"+job.PathSafeName()+".log")
	logFile, err := r.fs.Create(log)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func (r *JobRunner) runJobs(ctx context.Context, jobs []config.Job) []Result {
	results := make([]Result, len(jobs))
	var wg sync.WaitGroup

	for i, job := range jobs {
		wg.Go(func() {
			cmd, err := EvalJob(job.Run, Vars{})
			if err != nil {
				results[i] = Failure(job.DisplayName(), "", err)
				return
			}
			logFile, err := r.logFileForJob(&job)
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
	return r.runJobs(ctx, r.cfg.Config.Setup.Jobs), nil
}

func (r *JobRunner) RunTeardown(ctx context.Context) ([]Result, error) {
	return r.runJobs(ctx, r.cfg.Config.Teardown.Jobs), nil
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

	relFile := wsPath.Rel()

	for i, job := range jobs {
		wg.Go(func() {
			cmd, err := EvalJob(job.Run, Vars{File: relFile})
			if err != nil {
				results[i] = Failure(job.DisplayName(), "", err)
				return
			}
			logFile, err := r.logFileForJob(&job.Job)
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
