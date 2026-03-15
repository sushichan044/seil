package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/sushichan044/seil"
	"github.com/sushichan044/seil/internal/agent"
	"github.com/sushichan044/seil/internal/config"
	"github.com/sushichan044/seil/internal/reporter"
	"github.com/sushichan044/seil/internal/runner"
	"github.com/sushichan044/seil/internal/version"
)

type (
	CLI struct {
		Config   configFilePath   `short:"c" placeholder:"<path>" help:"Path to configuration file."                                    type:"existingfile"`
		Version  kong.VersionFlag `short:"v"`
		Reporter reporter.Name    `                               help:"Reporter to use for output. Possible values: ${reporter_names}"                     default:"auto" enum:"${reporter_names}"`

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
	ws, err := seil.NewWorkspace(cfg)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	results, err := ws.RunPostEditHooks(context.Background(), c.FilePath)
	if err != nil {
		return fmt.Errorf("failed to run post-edit hooks: %w", err)
	}
	return reportResults(results, cli.Reporter)
}

type SetupCmd struct {
}

func (c *SetupCmd) Run(cli *CLI, cfg *resolvedConfig) error {
	ws, err := seil.NewWorkspace(cfg)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	results, err := ws.RunSetupHooks(context.Background())
	if err != nil {
		return fmt.Errorf("failed to run setup hooks: %w", err)
	}
	return reportResults(results, cli.Reporter)
}

type TeardownCmd struct {
}

func (c *TeardownCmd) Run(cli *CLI, cfg *resolvedConfig) error {
	ws, err := seil.NewWorkspace(cfg)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	results, err := ws.RunTeardownHooks(context.Background())
	if err != nil {
		return fmt.Errorf("failed to run teardown hooks: %w", err)
	}
	return reportResults(results, cli.Reporter)
}

func reportResults(results []runner.HookResult, name reporter.Name) error {
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

func main() {
	cfg := resolvedConfig{}

	ctx := kong.Parse(&CLI{}, kong.Vars{
		"version":        fmt.Sprintf("seil %s", version.Get()),
		"reporter_names": strings.Join(reporter.ReporterNames, ","),
	}, kong.Bind(&cfg))
	ctx.FatalIfErrorf(ctx.Run())
}
