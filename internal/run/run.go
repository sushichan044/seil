package run

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/afero"

	"github.com/sushichan044/seil/internal/config"
)

const (
	logDirPerm    = 0700
	logFilePerm   = 0600
	randSuffixLen = 8
)

func Prepare(fs afero.Fs, cfg *config.ResolvedConfig) (*JobRunner, error) {
	if cfg.RootDir() == "" {
		return nil, errors.New("could not determine config root directory")
	}

	var logRoot string
	var customDir bool
	if dir := cfg.LogDir(); dir != "" {
		resolved, err := prepareCustomLogDir(fs, dir, cfg.RootDir())
		if err != nil {
			return nil, err
		}
		logRoot = resolved
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

func prepareCustomLogDir(fs afero.Fs, dir, rootDir string) (string, error) {
	if err := fs.MkdirAll(dir, logDirPerm); err != nil {
		return "", fmt.Errorf("failed to create log directory: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve log directory: %w", err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(rootDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve config root: %w", err)
	}
	if resolved != resolvedRoot && !strings.HasPrefix(resolved, resolvedRoot+string(filepath.Separator)) {
		return "", errors.New("log_dir resolves outside config root via symlink")
	}
	return resolved, nil
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
		return r.fs.OpenFile(filepath.Join(r.logRoot, filename), os.O_CREATE|os.O_EXCL|os.O_WRONLY, logFilePerm)
	}
	filename = hookType + "-" + job.PathSafeName() + ".log"
	return r.fs.Create(filepath.Join(r.logRoot, filename))
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
