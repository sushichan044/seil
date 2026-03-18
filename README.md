# seil

`seil` is a tool for managing lifecycle hooks across AI-native development workflows.

It lets you define setup, teardown, and post-edit automation once in `.seil.yml` and reuse it across integrations.

## Installation

### Using `go install`

```bash
go install github.com/sushichan044/seil/cmd/seil@latest
```

### Using `mise`

```bash
mise install github:sushichan044/seil
```

### Download binary

Download the latest release binaries from [Releases](https://github.com/sushichan044/seil/releases).

## Why seil

- One config file for AI agent lifecycle hooks.
- Human-readable output by default, JSON output when automation needs structured results
- `post-edit` hooks can be filtered by glob and skipped for `.gitignore`d files

## Integrations

### Claude Code hooks

You can route file edit events through `seil post-edit` instead of embedding formatter or lint commands directly in Claude Code hooks.

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "jq -r '.tool_input.file_path' | xargs -I {} env AI_AGENT=claude seil post-edit {} || exit 2"
          }
        ]
      }
    ]
  }
}
```

This keeps Claude Code responsible for detecting edits, while `.seil.yml` owns what should happen after the edit.

### Git Worktree Management (e.g. with `k1low/git-wt`)

You can also connect workspace lifecycle events to `setup` and `teardown`.

```ini
[wt]
basedir = ".git/wt"
hook = "seil setup"
deletehook = "seil teardown"
```

This works well when a worktree should run initialization on create and cleanup on delete.

## Usage

### Minimal config

Create `.seil.yml` in your repository:

```yaml
setup:
  jobs:
    - name: tidy
      run: go mod tidy

post_edit:
  jobs:
    - name: go-format
      glob: "**/*.go"
      run: mise run lint && mise run fmt

teardown:
  jobs:
    - name: cleanup
      run: echo "done"
```

### Run hooks

```bash
# Run setup hooks
seil setup

# Run post-edit hooks for a file
seil post-edit internal/config/config.go

# Run teardown hooks
seil teardown
```

### JSON output

Use `--reporter json` when another tool needs structured results.

```bash
seil --reporter json post-edit internal/config/config.go
```

The JSON result is grouped into `failure`, `success`, and `skipped`.

## Configuration

### File location

`seil` uses `.seil.yml`.

- If `-c, --config <path>` is provided, that file is used.
- Otherwise, `seil` searches from the current directory upward until the filesystem root.
- If `.seil.yml` is not found during auto-discovery, `seil` treats it as a no-op and exits successfully.
- To use a different filename such as `seil.yml`, pass it with `-c, --config`.

### Schema

```yaml
setup:
  jobs:
    - name: optional-name
      run: required shell command

post_edit:
  jobs:
    - name: optional-name
      glob: optional doublestar pattern
      run: required shell command

teardown:
  jobs:
    - name: optional-name
      run: required shell command
```

### Notes

- `run` is executed through `sh -c`.
- Hook commands run with the directory containing `.seil.yml` as their working directory.
- `name` is optional. If omitted, `seil` derives a normalized name from `run`.
- `post_edit.jobs[].glob` uses doublestar matching such as `**/*.go`.
- `post-edit` receives the edited file path and makes it available to command templating.

> [!NOTE]
> `run` commands use the directory containing the loaded `.seil.yml` as `cwd`.
> This is true both when `seil` finds the config automatically and when you pass it with `-c`.

## Behavior

- `setup` and `teardown` run all configured jobs and preserve result order from the config.
- `post-edit` skips jobs when the `glob` does not match the file path.
- `post-edit` also skips jobs when the file is matched by `.gitignore`.
- If no config file is found during auto-discovery, all commands return an empty result set and exit with status code `0`.
- The default human-readable output includes grouped results, status, log path, and a short summary.
- If any hook fails, `seil` exits with status code `1` for the default and JSON reporters, and `2` for the Claude reporter.

### AI agent behavior

- `--reporter auto` is the default.
- `AI_AGENT` overrides automatic agent detection when `--reporter auto` is used.
- `seil` can auto-detect `claude`, `cursor`, `devin`, `replit`, `gemini`, `codex`, `auggie`, `opencode`, `kiro`, `goose`, and `pi`.
- `--reporter auto` selects an agent-specific reporter when one exists.
- `--reporter json` always keeps the JSON reporter, even when an agent is detected.
- `--reporter default` forces the default human-readable reporter.
- `--reporter claude` forces the Claude reporter.
- Agents without a dedicated reporter fall back to the default human-readable output.

## Development

### Requirements

- Go 1.26+
- [mise](https://mise.jdx.dev/) for project tasks

### Quick commands

```bash
mise run fmt
mise run lint:fix
mise run test
```

## License

MIT
