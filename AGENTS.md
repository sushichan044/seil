# AGENTS.md

This file provides guidance to Coding Agents when working with code in this repository.

---

## Quick Commands

```bash
mise run test               # Run tests with coverage
mise run lint               # Run golangci-lint for code quality checks
mise run lint-fix           # Run golangci-lint and Auto-fix linting issues
mise run fmt                # Format code
mise run build-snapshot     # Build cross-platform binaries with goreleaser
mise run clean              # Remove generated files

# Standard Go commands
go run ./cmd/himo           # Run CLI in development mode
go test ./...               # Run all tests
go mod tidy                 # Clean up dependencies
```

## Project Context

## Sources of Truth

Keep this file light. For implementation details, refer to:

- Product and usage overview: `README.md`
- CLI entry point: `cmd/root.go`
- Package layout and behavior: `internal/`
- Dependencies and versions: `go.mod`, `go.sum`
- Task runner and scripts: `mise.toml`
- Lint/format rules: `.golangci.yml`
- Release/build configuration: `.goreleaser.yaml`

## Coding Standard

- use `testify` in tests
- keep CLI integration tests build-based; prefer `helper_test.go` for shared test helpers
- use `alecthomas/kong` to define CLI commands and flags
- prefer template-first config validation, then schema-parse the evaluated shape
- prefer `Oudwins/zog` (unreleased when needed for `z.EXPERIMENTAL_MAP`) for config schema validation
- return aggregated validation errors via `errors.Join` instead of fail-fast when multiple config problems exist
- use `bmatcuk/doublestar` for glob matching
- use `spf13/afero` for file system operations
  - High testability and flexibility for different environments
- use `sabhiram/go-gitignore` for parsing and matching `.gitignore` files
  - do not emit an external command N times inside a file loop; collect `.gitignore` files under `ConfigRoot`, prepend relative roots as needed, compile once, then match in memory
