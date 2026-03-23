# Replace `internal/runner` with `internal/run` Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the deleted `internal/runner` package with `internal/run`, redesigning `run.Result`, template evaluation (`EvalJob`), and glob filtering (`GlobJob.Matches`) while migrating all consumers.

**Architecture:** The `internal/run` package becomes the single execution layer — `JobRunner` runs jobs after template expansion; the public `seil` package handles filtering via `config.GlobJob.Matches` (gitignore + glob) and delegates execution to `JobRunner`; `reporter` works directly with `run.Result`. The `internal/template` package is deleted entirely.

**Tech Stack:** Go 1.26, `github.com/bmatcuk/doublestar/v4` (glob), `afero` (fs), `text/template` (template eval), `testify` (testing)

---

## Key Design Decisions

- `run.Status` is a `string` type (`"success"` / `"failure"` / `"skipped"`) — clean JSON serialization, no MarshalJSON needed
- `run.Result` has **no** `ExitCode` or `Summary` fields — reporter computes exit codes from failure counts; LLMs read log files directly
- `EvalJob(tmpl string, vars Vars) (string, error)` — `Vars{File string}`, funcMap: `dir`/`base`/`ext` (path manipulation); `ConfigRoot` is a future addition
- `GlobJob.Matches(filePath, configRoot string) bool` — normalizes path relative to configRoot before glob matching; gitignore check stays in `postedit.go`
- `RunPostEdit` takes pre-filtered `[]config.GlobJob` — filtering logic lives in `postedit.go`, `JobRunner` only executes

---

## Chunk 1: Core building blocks

### Task 1: Redesign `internal/run/template.go`

**Files:**
- Modify: `internal/run/template.go`
- Create: `internal/run/template_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/run/template_test.go
package run_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/run"
)

func TestEvalJob(t *testing.T) {
	t.Run("substitutes .File", func(t *testing.T) {
		got, err := run.EvalJob("gofmt -w {{.File}}", run.Vars{File: "main.go"})
		require.NoError(t, err)
		assert.Equal(t, "gofmt -w main.go", got)
	})

	t.Run("dir function returns directory of file", func(t *testing.T) {
		got, err := run.EvalJob("go test {{.File | dir}}", run.Vars{File: "pkg/foo/bar.go"})
		require.NoError(t, err)
		assert.Equal(t, "go test pkg/foo", got)
	})

	t.Run("base function returns base name", func(t *testing.T) {
		got, err := run.EvalJob("echo {{.File | base}}", run.Vars{File: "pkg/foo/bar.go"})
		require.NoError(t, err)
		assert.Equal(t, "echo bar.go", got)
	})

	t.Run("ext function returns extension", func(t *testing.T) {
		got, err := run.EvalJob("echo {{.File | ext}}", run.Vars{File: "main.go"})
		require.NoError(t, err)
		assert.Equal(t, "echo .go", got)
	})

	t.Run("no template directives returns input unchanged", func(t *testing.T) {
		got, err := run.EvalJob("echo hello", run.Vars{File: "main.go"})
		require.NoError(t, err)
		assert.Equal(t, "echo hello", got)
	})

	t.Run("returns error on invalid template syntax", func(t *testing.T) {
		_, err := run.EvalJob("{{.File | unknown}}", run.Vars{File: "main.go"})
		assert.Error(t, err)
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

```
go test ./internal/run/... -run TestEvalJob -v
```
Expected: compile error or FAIL (EvalJob / Vars not yet defined with new API)

- [ ] **Step 3: Implement new `internal/run/template.go`**

```go
package run

import (
	"path/filepath"
	"strings"
	"text/template"
)

//nolint:gochecknoglobals // funcMap is a static map of template functions
var funcMap = template.FuncMap{
	"dir":  filepath.Dir,
	"base": filepath.Base,
	"ext":  filepath.Ext,
}

// Vars holds template variables available during hook command evaluation.
type Vars struct {
	File string
}

// EvalJob evaluates a Go template string with the given Vars.
func EvalJob(tmpl string, vars Vars) (string, error) {
	t, err := template.New("").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err = t.Execute(&sb, vars); err != nil {
		return "", err
	}
	return sb.String(), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```
go test ./internal/run/... -run TestEvalJob -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/run/template.go internal/run/template_test.go
git commit -m "feat(run): redesign EvalJob with Vars{File} and path funcMap (dir/base/ext)"
```

---

### Task 2: Add `GlobJob.Matches` to `internal/config/job.go`

**Files:**
- Modify: `internal/config/job.go`
- Modify: `internal/config/job_test.go`

- [ ] **Step 1: Check existing `job_test.go` for the right package/import pattern**

Read `internal/config/job_test.go` to see existing test structure before adding.

- [ ] **Step 2: Write failing tests** (append to `job_test.go`)

```go
func TestGlobJob_Matches(t *testing.T) {
	t.Run("returns true when glob is empty (always run)", func(t *testing.T) {
		job := config.GlobJob{Glob: ""}
		assert.True(t, job.Matches("main.go", "/project"))
	})

	t.Run("returns true when file matches glob pattern", func(t *testing.T) {
		job := config.GlobJob{Glob: "**/*.go"}
		assert.True(t, job.Matches("/project/main.go", "/project"))
	})

	t.Run("returns false when file does not match glob pattern", func(t *testing.T) {
		job := config.GlobJob{Glob: "**/*.ts"}
		assert.False(t, job.Matches("/project/main.go", "/project"))
	})

	t.Run("matches relative to configRoot", func(t *testing.T) {
		job := config.GlobJob{Glob: "src/**/*.go"}
		assert.True(t, job.Matches("/project/src/foo/bar.go", "/project"))
		assert.False(t, job.Matches("/project/main.go", "/project"))
	})
}
```

- [ ] **Step 3: Run tests to verify they fail**

```
go test ./internal/config/... -run TestGlobJob_Matches -v
```
Expected: FAIL (Matches not defined)

- [ ] **Step 4: Implement `GlobJob.Matches`** in `internal/config/job.go`

Add imports `"path/filepath"` and `"github.com/bmatcuk/doublestar/v4"`, then add:

```go
// Matches reports whether this job should run for the given filePath.
// filePath and configRoot should both be absolute paths.
// If Glob is empty, the job always matches.
func (j *GlobJob) Matches(filePath, configRoot string) bool {
	if j.Glob == "" {
		return true
	}
	rel, err := filepath.Rel(configRoot, filePath)
	if err != nil {
		rel = filePath
	}
	normalized := filepath.ToSlash(rel)
	matched, err := doublestar.Match(j.Glob, normalized)
	return err == nil && matched
}
```

- [ ] **Step 5: Run tests to verify they pass**

```
go test ./internal/config/... -run TestGlobJob_Matches -v
```
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/config/job.go internal/config/job_test.go
git commit -m "feat(config): add GlobJob.Matches for glob-based file filtering"
```

---

### Task 3: Update `internal/run/result.go` — export `Status` as string type

**Files:**
- Modify: `internal/run/result.go`

- [ ] **Step 1: Rewrite `result.go`**

Change `status` from `int8` to string type, add JSON tags:

```go
package run

// Status represents the outcome of a job execution.
type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
	StatusSkipped Status = "skipped"
)

// Result holds the outcome of a single job run.
type Result struct {
	Name    string `json:"name"`
	Status  Status `json:"status"`
	LogFile string `json:"log_file"`
	err     error
}

func Success(name, logFile string) Result {
	return Result{Name: name, Status: StatusSuccess, LogFile: logFile}
}

func Failure(name, logFile string, err error) Result {
	return Result{Name: name, Status: StatusFailure, LogFile: logFile, err: err}
}

func Skipped(name string) Result {
	return Result{Name: name, Status: StatusSkipped}
}
```

- [ ] **Step 2: Verify `internal/run` package compiles**

```
go build ./internal/run/...
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/run/result.go
git commit -m "feat(run): export Status as string type with JSON tags, remove ExitCode/Summary"
```

---

## Chunk 2: Update `run.go` and reporter

### Task 4: Update `internal/run/run.go`

**Files:**
- Modify: `internal/run/run.go`

Changes:
- `RunSetup` / `RunTeardown` — call `EvalJob(job.Run, Vars{})` before executing
- `RunPostEdit` — change signature to `(ctx context.Context, filePath string, jobs []config.GlobJob)`, call `EvalJob(job.Run, Vars{File: filePath})`
- Use `job.DisplayName()` instead of `job.Name` for Result names

- [ ] **Step 1: Rewrite `run.go`**

```go
package run

import (
	"context"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"

	"github.com/sushichan044/seil/internal/config"
)

func Prepare(fs afero.Fs, cfg *config.ResolvedConfig) (*JobRunner, error) {
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

func (r *JobRunner) RunSetup(ctx context.Context) ([]Result, error) {
	results := make([]Result, len(r.cfg.Config.Setup.Jobs))
	var wg sync.WaitGroup

	for i, job := range r.cfg.Config.Setup.Jobs {
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
			proc.Dir = r.cfg.CWD()
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

func (r *JobRunner) RunTeardown(ctx context.Context) ([]Result, error) {
	results := make([]Result, len(r.cfg.Config.Teardown.Jobs))
	var wg sync.WaitGroup

	for i, job := range r.cfg.Config.Teardown.Jobs {
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
			proc.Dir = r.cfg.CWD()
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

// RunPostEdit executes the given pre-filtered jobs for the edited file at filePath.
// Callers are responsible for filtering jobs (via GlobJob.Matches + gitignore) before calling.
func (r *JobRunner) RunPostEdit(ctx context.Context, filePath string, jobs []config.GlobJob) ([]Result, error) {
	results := make([]Result, len(jobs))
	var wg sync.WaitGroup

	for i, job := range jobs {
		wg.Go(func() {
			cmd, err := EvalJob(job.Run, Vars{File: filePath})
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
			proc.Dir = r.cfg.CWD()
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
```

- [ ] **Step 2: Verify `internal/run` compiles**

```
go build ./internal/run/...
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/run/run.go
git commit -m "feat(run): update RunPostEdit to take pre-filtered GlobJobs, add EvalJob to all RunX methods"
```

---

### Task 5: Migrate `reporter` to `run.Result`

**Files:**
- Modify: `internal/reporter/reporter.go`
- Modify: `internal/reporter/claude.go`
- Modify: `internal/reporter/reporter_test.go`
- Modify: `internal/reporter/testdata/TestHumanReporter_Report.golden`
- Modify: `internal/reporter/testdata/TestClaudeReporter_ReportStdout.golden`
- Modify: `internal/reporter/testdata/TestClaudeReporter_ReportStderr.golden`

- [ ] **Step 1: Rewrite `reporter.go`**

Replace all `runner.HookResult` with `run.Result`. Key changes:
- `groupedHookResults` → `groupedResults` using `[]run.Result`
- `writeHookResult` → `writeResult` (no `exit_code:` / `summary:` lines)
- Switch on `run.StatusFailure` / `run.StatusSuccess` / `run.StatusSkipped`

```go
package reporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/sushichan044/seil/internal/agent"
	"github.com/sushichan044/seil/internal/run"
)

type Reporter interface {
	Report(results []run.Result, stdout io.Writer, stderr io.Writer) (int, error)
}

// ... Name, ParseName, UnmarshalText, ReporterNames, Resolve stay unchanged ...

type groupedResults struct {
	Failure []run.Result `json:"failure"`
	Success []run.Result `json:"success"`
	Skipped []run.Result `json:"skipped"`
}

func (HumanReporter) Report(results []run.Result, stdout io.Writer, _ io.Writer) (int, error) {
	grouped := groupResults(results)

	if _, err := fmt.Fprintf(stdout, "--- Failures (%d) ---\n", len(grouped.Failure)); err != nil {
		return 0, err
	}
	for _, result := range grouped.Failure {
		if err := writeResult(stdout, result); err != nil {
			return 0, err
		}
	}

	if _, err := fmt.Fprintf(stdout, "\n--- Successes (%d) ---\n", len(grouped.Success)); err != nil {
		return 0, err
	}
	for _, result := range grouped.Success {
		if err := writeResult(stdout, result); err != nil {
			return 0, err
		}
	}

	if _, err := fmt.Fprintf(stdout, "\n--- Skipped (%d) ---\n", len(grouped.Skipped)); err != nil {
		return 0, err
	}
	for _, result := range grouped.Skipped {
		if err := writeResult(stdout, result); err != nil {
			return 0, err
		}
	}

	if _, err := fmt.Fprintf(stdout, "\n---\n%s\n", summaryLine(grouped)); err != nil {
		return 0, err
	}
	return defaultExitCode(grouped), nil
}

func (JSONReporter) Report(results []run.Result, stdout io.Writer, _ io.Writer) (int, error) {
	grouped := groupResults(results)
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(grouped); err != nil {
		return 0, err
	}
	return defaultExitCode(grouped), nil
}

func groupResults(results []run.Result) groupedResults {
	grouped := groupedResults{
		Failure: []run.Result{},
		Success: []run.Result{},
		Skipped: []run.Result{},
	}
	for _, result := range results {
		switch result.Status {
		case run.StatusFailure:
			grouped.Failure = append(grouped.Failure, result)
		case run.StatusSuccess:
			grouped.Success = append(grouped.Success, result)
		case run.StatusSkipped:
			grouped.Skipped = append(grouped.Skipped, result)
		}
	}
	return grouped
}

func writeResult(w io.Writer, result run.Result) error {
	_, err := fmt.Fprintf(w, "\nhook: %s\nstatus: %s\nlog: %s\n",
		result.Name, result.Status, result.LogFile)
	return err
}

func defaultExitCode(grouped groupedResults) int {
	if len(grouped.Failure) > 0 {
		return 1
	}
	return 0
}

func summaryLine(grouped groupedResults) string {
	return fmt.Sprintf("%d succeeded, %d failed, %d skipped",
		len(grouped.Success), len(grouped.Failure), len(grouped.Skipped))
}
```

- [ ] **Step 2: Rewrite `claude.go`**

```go
package reporter

import (
	"fmt"
	"io"

	"github.com/sushichan044/seil/internal/run"
)

type ClaudeReporter struct{}

const claudeFailureExitCode = 2

func (ClaudeReporter) Report(results []run.Result, stdout io.Writer, stderr io.Writer) (int, error) {
	grouped := groupResults(results)
	for _, result := range grouped.Failure {
		if err := writeResult(stderr, result); err != nil {
			return 0, err
		}
	}
	if _, err := fmt.Fprintf(stdout, "%s\n", summaryLine(grouped)); err != nil {
		return 0, err
	}
	return claudeExitCode(grouped), nil
}

func claudeExitCode(grouped groupedResults) int {
	if len(grouped.Failure) > 0 {
		return claudeFailureExitCode
	}
	return 0
}
```

- [ ] **Step 3: Update `reporter_test.go`** — replace `runner.HookResult` with `run.Result`

```go
import (
	"errors"
	// remove: "github.com/sushichan044/seil/internal/runner"
	// add:
	"github.com/sushichan044/seil/internal/run"
)

type groupedResultsJSON struct {
	Failure []run.Result `json:"failure"`
	Success []run.Result `json:"success"`
	Skipped []run.Result `json:"skipped"`
}

func sampleResults() []run.Result {
	return []run.Result{
		run.Success("ok", "/tmp/ok.log"),
		run.Failure("fail", "/tmp/fail.log", errors.New("exit status 1")),
		run.Skipped("skip"),
	}
}
```

- [ ] **Step 4: Update golden files**

`internal/reporter/testdata/TestHumanReporter_Report.golden`:
```
--- Failures (1) ---

hook: fail
status: failure
log: /tmp/fail.log

--- Successes (1) ---

hook: ok
status: success
log: /tmp/ok.log

--- Skipped (1) ---

hook: skip
status: skipped
log:

---
1 succeeded, 1 failed, 1 skipped
```

`internal/reporter/testdata/TestClaudeReporter_ReportStdout.golden`:
```
1 succeeded, 1 failed, 1 skipped
```

`internal/reporter/testdata/TestClaudeReporter_ReportStderr.golden`:
```

hook: fail
status: failure
log: /tmp/fail.log
```

- [ ] **Step 5: Run reporter tests**

```
go test ./internal/reporter/... -v
```
Expected: PASS. If golden files don't match exactly, run with `-update` flag:
```
go test ./internal/reporter/... -update
```
Then review the diff and commit.

- [ ] **Step 6: Commit**

```bash
git add internal/reporter/
git commit -m "feat(reporter): migrate from runner.HookResult to run.Result, remove exit_code/summary"
```

---

## Chunk 3: Rewrite consumers and delete `internal/template`

### Task 6: Rewrite `postedit.go`, `setup.go`, `teardown.go`

**Files:**
- Modify: `postedit.go`
- Modify: `setup.go`
- Modify: `teardown.go`

- [ ] **Step 1: Rewrite `postedit.go`**

```go
package seil

import (
	"context"

	"github.com/spf13/afero"

	"github.com/sushichan044/seil/internal/config"
	"github.com/sushichan044/seil/internal/gitignore"
	"github.com/sushichan044/seil/internal/run"
)

func runPostEditHooks(
	ctx context.Context,
	cfg *config.ResolvedConfig,
	gitignoreMatcher *gitignore.Matcher,
	fs afero.Fs,
	filePath string,
) ([]run.Result, error) {
	jobs := cfg.Config.PostEdit.Jobs
	results := make([]run.Result, len(jobs))
	var toRun []config.GlobJob
	var toRunIdx []int

	for i, job := range jobs {
		if !job.Matches(filePath, cfg.CWD()) || gitignoreMatcher.IsIgnored(filePath) {
			results[i] = run.Skipped(job.DisplayName())
			continue
		}
		toRun = append(toRun, job)
		toRunIdx = append(toRunIdx, i)
	}

	if len(toRun) == 0 {
		return results, nil
	}

	r, err := run.Prepare(fs, cfg)
	if err != nil {
		return nil, err
	}

	runResults, err := r.RunPostEdit(ctx, filePath, toRun)
	if err != nil {
		return nil, err
	}

	for k, idx := range toRunIdx {
		results[idx] = runResults[k]
	}
	return results, nil
}
```

- [ ] **Step 2: Rewrite `setup.go`**

```go
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
```

- [ ] **Step 3: Rewrite `teardown.go`**

```go
package seil

import (
	"context"

	"github.com/spf13/afero"

	"github.com/sushichan044/seil/internal/config"
	"github.com/sushichan044/seil/internal/run"
)

func runTeardownHooks(
	ctx context.Context,
	cfg *config.ResolvedConfig,
	fs afero.Fs,
) ([]run.Result, error) {
	if len(cfg.Config.Teardown.Jobs) == 0 {
		return []run.Result{}, nil
	}
	r, err := run.Prepare(fs, cfg)
	if err != nil {
		return nil, err
	}
	return r.RunTeardown(ctx)
}
```

- [ ] **Step 4: Commit**

```bash
git add postedit.go setup.go teardown.go
git commit -m "refactor: rewrite postedit/setup/teardown to use internal/run.JobRunner"
```

---

### Task 7: Update `workspace.go` and `cmd/seil/main.go`

**Files:**
- Modify: `workspace.go`
- Modify: `cmd/seil/main.go`

- [ ] **Step 1: Rewrite `workspace.go`**

```go
package seil

import (
	"context"

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
	m, err := gitignore.NewMatcherFromRoot(w.fs, w.config.CWD())
	if err != nil {
		return nil, err
	}
	return runPostEditHooks(ctx, w.config, m, w.fs, filePath)
}

// RunSetupHooks executes all setup hooks.
func (w *Workspace) RunSetupHooks(ctx context.Context) ([]run.Result, error) {
	return runSetupHooks(ctx, w.config, w.fs)
}

// RunTeardownHooks executes all teardown hooks.
func (w *Workspace) RunTeardownHooks(ctx context.Context) ([]run.Result, error) {
	return runTeardownHooks(ctx, w.config, w.fs)
}
```

- [ ] **Step 2: Update `cmd/seil/main.go`** — change `reportResults` signature

Remove `"github.com/sushichan044/seil/internal/runner"` import, add `"github.com/sushichan044/seil/internal/run"`, update:

```go
func reportResults(results []run.Result, name reporter.Name) error {
	r := reporter.Resolve(name, agent.Detect())
	exitCode, err := r.Report(results, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}
```

- [ ] **Step 3: Build the whole project**

```
go build ./...
```
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add workspace.go cmd/seil/main.go
git commit -m "refactor: update workspace and main to use run.Result"
```

---

### Task 8: Update all tests

**Files:**
- Modify: `setup_test.go`
- Modify: `teardown_test.go`
- Modify: `postedit_test.go`
- Modify: `cmd/seil/main_test.go`

- [ ] **Step 1: Update `setup_test.go`**

- Remove `"github.com/sushichan044/seil/internal/runner"` import
- Add `"github.com/sushichan044/seil/internal/run"`
- Replace `runner.HookStatusSuccess` → `run.StatusSuccess`
- Replace `runner.HookStatusFailure` → `run.StatusFailure`
- Remove all `assert.Equal(t, 0, results[0].ExitCode)` lines
- Remove all `assert.Equal(t, "...", results[0].Summary)` lines

- [ ] **Step 2: Update `teardown_test.go`** (same pattern as `setup_test.go`)

- [ ] **Step 3: Update `postedit_test.go`**

Same import/status swap as above, plus:
- Remove `assert.Equal(t, 0, results[0].ExitCode)` lines
- For the `"template variables are expanded in command"` test: the original assertion was `assert.Equal(t, "main.go", results[0].Summary)`. Replace with just checking `assert.Equal(t, run.StatusSuccess, results[0].Status)` — the template is tested by the command succeeding.

- [ ] **Step 4: Update `cmd/seil/main_test.go`**

Update `hookResultJSON` struct:
```go
type hookResultJSON struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	LogFile string `json:"log_file"`
}
```

For each test that asserts `ExitCode` or `Summary`, remove those assertions. For `LogPath` assertions, update to `LogFile`.
Example change in `TestPostEdit_JSON_Schema`:
```go
// Before:
assert.Equal(t, 0, results.Success[0].ExitCode)
assert.NotEmpty(t, results.Success[0].LogPath)
assert.Equal(t, "hello", results.Success[0].Summary)

// After:
assert.NotEmpty(t, results.Success[0].LogFile)
```

- [ ] **Step 5: Run all tests**

```
go test ./...
```
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add setup_test.go teardown_test.go postedit_test.go cmd/seil/main_test.go
git commit -m "test: migrate all tests from runner.HookResult to run.Result"
```

---

### Task 9: Delete `internal/template`

**Files:**
- Delete: `internal/template/eval.go`
- Delete: `internal/template/eval_test.go`

- [ ] **Step 1: Remove files**

```bash
git rm internal/template/eval.go internal/template/eval_test.go
```

- [ ] **Step 2: Verify no remaining references**

```
grep -r "internal/template" --include="*.go" .
```
Expected: no output

- [ ] **Step 3: Build and run all tests**

```
go build ./...
go test ./...
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git commit -m "chore: delete internal/template (replaced by internal/run EvalJob)"
```

---

## Final Verification

- [ ] Run full test suite: `go test ./...`
- [ ] Build binary: `go build ./cmd/seil/`
- [ ] Confirm no references to `internal/runner` or `internal/template`: `grep -r "internal/runner\|internal/template" --include="*.go" .`
