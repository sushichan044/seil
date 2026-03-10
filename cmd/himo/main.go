package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/sushichan044/himo/internal/config"
	"github.com/sushichan044/himo/internal/version"
)

type CLI struct {
	Config  configFilePath   `short:"c" placeholder:"<path>" help:"Path to configuration file." type:"existingfile"`
	Version kong.VersionFlag `short:"v"`

	CheckFile CheckFileCmd `cmd:"" help:"Check if a file exists."`
}

type (
	configFilePath string
	resolvedConfig = config.ResolvedConfig
)

// AfterApply is called after kong has parsed the command-line arguments and before executing the command.
// We use this to load the configuration file and set into binding struct.
func (cli *CLI) AfterApply(r *resolvedConfig) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	var cfg *resolvedConfig
	if cli.Config == "" {
		cfg, err = config.FindUpAndLoad(wd)
		if err != nil {
			return fmt.Errorf("failed to find config: %w", err)
		}
	} else {
		cfg, err = config.Load(string(cli.Config))
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	*r = *cfg
	return nil
}

type CheckFileCmd struct {
	FilePath string `arg:"" name:"file" help:"Path to the file to check." type:"existingfile"`
}

func (c *CheckFileCmd) Run(cli *CLI, cfg *resolvedConfig) error {
	fmt.Printf("File '%s' exists.\n", c.FilePath)
	fmt.Printf("Loaded config: %+v\n", cfg.Config)
	fmt.Printf("Config cwd: %s\n", cfg.CWD)
	return nil
}

func main() {
	cfg := resolvedConfig{}

	ctx := kong.Parse(&CLI{}, kong.Vars{
		"version": fmt.Sprintf("himo %s", version.Get()),
	}, kong.Bind(&cfg))
	ctx.FatalIfErrorf(ctx.Run())
}
