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
	"github.com/sushichan044/himo/internal/postedit"
	"github.com/sushichan044/himo/internal/version"
)

type CLI struct {
	Config  configFilePath   `short:"c" placeholder:"<path>" help:"Path to configuration file." type:"existingfile"`
	Version kong.VersionFlag `short:"v"`

	PostEdit PostEditCmd `cmd:"" help:"Run post-edit hooks for a file."`
}

type (
	configFilePath string
	resolvedConfig = config.ResolvedConfig
)

// AfterApply is called after kong has parsed the command-line arguments and before executing the command.
// We use this to load the configuration file and set into binding struct.
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
	FilePath string `arg:"" name:"file" help:"Path to the edited file."       type:"existingfile"`
	JSON     bool   `       name:"json" help:"Output results in JSON format."`
}

func (c *PostEditCmd) Run(cfg *resolvedConfig) error {
	ws, err := himo.NewWorkspace(cfg)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	results, err := ws.RunPostEditHooks(context.Background(), c.FilePath)
	if err != nil {
		return fmt.Errorf("failed to run post-edit hooks: %w", err)
	}

	if c.JSON {
		return printJSON(results)
	}
	return printText(results)
}

func printJSON(results []postedit.HookResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func printText(results []postedit.HookResult) error {
	if _, err := fmt.Fprintln(os.Stdout, "=== post-edit hooks result ==="); err != nil {
		return err
	}

	succeeded, failed, skipped := 0, 0, 0
	for _, r := range results {
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

		switch r.Status {
		case postedit.HookStatusSuccess:
			succeeded++
		case postedit.HookStatusFailure:
			failed++
		case postedit.HookStatusSkipped:
			skipped++
		}
	}

	_, err := fmt.Fprintf(os.Stdout, "\n---\n%d succeeded, %d failed, %d skipped\n",
		succeeded, failed, skipped)
	return err
}

func main() {
	cfg := resolvedConfig{}

	ctx := kong.Parse(&CLI{}, kong.Vars{
		"version": fmt.Sprintf("himo %s", version.Get()),
	}, kong.Bind(&cfg))
	ctx.FatalIfErrorf(ctx.Run())
}
