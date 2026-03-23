# Plan: Improve Test Coverage to ~80%

## Context

Current overall coverage: 52.7%. Target: ~80%.

Key gaps:
- `internal/run`: 9.2% (result.go and run.go have 0%)
- `internal/agent`: 53.8% (many DetectFromLookup cases untested)
- `internal/reporter`: 72.3% (UnmarshalText, HumanReporter no-failure path)

## Tasks

### Task 1: Add `internal/run/result_test.go`

Test all three constructors: Success, Failure, Skipped.
Check Name, Status, LogFile fields.

Verify:
- `go test ./internal/run/...` passes

### Task 2: Add `internal/run/run_test.go`

Test Prepare, RunSetup, RunTeardown, RunPostEdit using real OS fs + temp dirs.
Use `config.Load(afero.NewOsFs(), cfgPath)` pattern from postedit_test.go.

Cases:
- RunSetup: success, failure (exit 1), empty jobs
- RunTeardown: success, failure, empty jobs
- RunPostEdit: success, template var expansion, failure

Verify:
- `go test ./internal/run/...` passes

### Task 3: Add more `internal/agent` tests

Missing cases in `agent_test.go`:
- Parse: all agents, whitespace/case normalization, unknown
- DetectFromLookup: Devin (EDITOR contains "devin"), Replit, Gemini, OpenCode, Auggie, Goose, Kiro, Pi (Windows path `.pi\agent`)

Verify:
- `go test ./internal/agent/...` passes

### Task 4: Add more `internal/reporter` tests

Missing in `reporter_test.go`:
- Name.UnmarshalText: valid (claude, json, etc.) and invalid name
- HumanReporter.Report: success-only results (exit code 0)
- ClaudeReporter.Report: success-only results (exit code 0)

Verify:
- `go test ./internal/reporter/...` passes

### Task 5: Verify overall coverage

Run `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -5`
Confirm total >= 80%.
