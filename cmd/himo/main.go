package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/sushichan044/himo"
	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/runner"
	"github.com/sushichan044/himo/internal/version"
)

type (
	CLI struct {
		Config  configFilePath   `short:"c" placeholder:"<path>" help:"Path to configuration file."    type:"existingfile"`
		Version kong.VersionFlag `short:"v"`
		JSON    bool             `                               help:"Output results in JSON format."`

		PostEdit PostEditCmd `cmd:"" help:"Run post-edit hooks for a file."`
		Setup    SetupCmd    `cmd:"" help:"Run setup hooks."`
		Teardown TeardownCmd `cmd:"" help:"Run teardown hooks."`
	}

	configFilePath string
	resolvedConfig = config.ResolvedConfig
)

func (cli *CLI) AfterApply(r *resolvedConfig) error {
	cfgPath := string(cli.Config)

	if cfgPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		cfgPath, err = config.ResolveConfigFilePathFrom(wd)
		if err != nil {
			return fmt.Errorf("failed to resolve config file path: %w", err)
		}
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	*r = *cfg
	return nil
}

type PostEditCmd struct {
	FilePath string `arg:"" name:"file" help:"Path to the edited file." type:"existingfile"`
}

func (c *PostEditCmd) Run(cli *CLI, cfg *resolvedConfig) error {
	ws, err := himo.NewWorkspace(cfg)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	results, err := ws.RunPostEditHooks(context.Background(), c.FilePath)
	if err != nil {
		return fmt.Errorf("failed to run post-edit hooks: %w", err)
	}

	var hasFailure bool
	if cli.JSON {
		hasFailure, err = printJSON(results)
	} else {
		hasFailure, err = printText(results, "post-edit hooks result")
	}
	if err != nil {
		return err
	}
	if hasFailure {
		os.Exit(1)
	}
	return nil
}

type SetupCmd struct {
}

func (c *SetupCmd) Run(cli *CLI, cfg *resolvedConfig) error {
	ws, err := himo.NewWorkspace(cfg)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	results, err := ws.RunSetupHooks(context.Background())
	if err != nil {
		return fmt.Errorf("failed to run setup hooks: %w", err)
	}

	var hasFailure bool
	if cli.JSON {
		hasFailure, err = printJSON(results)
	} else {
		hasFailure, err = printText(results, "setup hooks result")
	}
	if err != nil {
		return err
	}
	if hasFailure {
		os.Exit(1)
	}
	return nil
}

type TeardownCmd struct {
}

func (c *TeardownCmd) Run(cli *CLI, cfg *resolvedConfig) error {
	ws, err := himo.NewWorkspace(cfg)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	results, err := ws.RunTeardownHooks(context.Background())
	if err != nil {
		return fmt.Errorf("failed to run teardown hooks: %w", err)
	}

	var hasFailure bool
	if cli.JSON {
		hasFailure, err = printJSON(results)
	} else {
		hasFailure, err = printText(results, "teardown hooks result")
	}
	if err != nil {
		return err
	}
	if hasFailure {
		os.Exit(1)
	}
	return nil
}

type groupedHookResults struct {
	Failure []runner.HookResult `json:"failure"`
	Success []runner.HookResult `json:"success"`
	Skipped []runner.HookResult `json:"skipped"`
}

func groupResults(results []runner.HookResult) groupedHookResults {
	g := groupedHookResults{
		Failure: []runner.HookResult{},
		Success: []runner.HookResult{},
		Skipped: []runner.HookResult{},
	}
	for _, r := range results {
		switch r.Status {
		case runner.HookStatusFailure:
			g.Failure = append(g.Failure, r)
		case runner.HookStatusSuccess:
			g.Success = append(g.Success, r)
		case runner.HookStatusSkipped:
			g.Skipped = append(g.Skipped, r)
		}
	}
	return g
}

func printJSON(results []runner.HookResult) (bool, error) {
	g := groupResults(results)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return len(g.Failure) > 0, enc.Encode(g)
}

func printHookResult(r runner.HookResult) error {
	if _, err := fmt.Fprintf(os.Stdout, "\nhook: %s\nstatus: %s\nexit_code: %d\nlog: %s\n",
		r.Name, r.Status, r.ExitCode, r.LogPath); err != nil {
		return err
	}
	if r.Summary != "" {
		lines := strings.Split(r.Summary, "\n")
		if _, err := fmt.Fprintln(os.Stdout, "summary:"); err != nil {
			return err
		}
		for _, line := range lines {
			if _, err := fmt.Fprintf(os.Stdout, "  %s\n", line); err != nil {
				return err
			}
		}
	}
	return nil
}

func printText(results []runner.HookResult, title string) (bool, error) {
	g := groupResults(results)

	if _, err := fmt.Fprintf(os.Stdout, "=== %s ===\n", title); err != nil {
		return false, err
	}

	if _, err := fmt.Fprintf(os.Stdout, "\n--- Failures (%d) ---\n", len(g.Failure)); err != nil {
		return false, err
	}
	for _, r := range g.Failure {
		if err := printHookResult(r); err != nil {
			return false, err
		}
	}

	if _, err := fmt.Fprintf(os.Stdout, "\n--- Successes (%d) ---\n", len(g.Success)); err != nil {
		return false, err
	}
	for _, r := range g.Success {
		if err := printHookResult(r); err != nil {
			return false, err
		}
	}

	if _, err := fmt.Fprintf(os.Stdout, "\n--- Skipped (%d) ---\n", len(g.Skipped)); err != nil {
		return false, err
	}
	for _, r := range g.Skipped {
		if err := printHookResult(r); err != nil {
			return false, err
		}
	}

	_, err := fmt.Fprintf(os.Stdout, "\n---\n%d succeeded, %d failed, %d skipped\n",
		len(g.Success), len(g.Failure), len(g.Skipped))
	return len(g.Failure) > 0, err
}

func main() {
	cfg := resolvedConfig{}

	ctx := kong.Parse(&CLI{}, kong.Vars{
		"version": fmt.Sprintf("himo %s", version.Get()),
	}, kong.Bind(&cfg))
	ctx.FatalIfErrorf(ctx.Run())
}
