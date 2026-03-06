package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/badele/ansi-compositor/compositor"
	"github.com/badele/ansi-compositor/config"
)

// Version information (set by build flags)
var (
	version = "dev"
	commit  = "none"
)

// CLI defines the command-line interface.
type CLI struct {
	// Config file path (required)
	Config string `arg:"" type:"existingfile" help:"YAML configuration file"`

	// Output file (overrides config)
	Output string `short:"o" type:"path" help:"Output file (overrides config, stdout if not specified)"`

	// Output format (overrides config)
	Format string `short:"F" help:"Output format: ansi, neotex, plaintext (overrides config)"`

	// Inline output
	Inline bool `short:"I" help:"Output on single line"`

	// Keep trailing empty lines
	KeepTrailingLines bool `short:"K" help:"Preserve trailing empty lines in ansi/neotex output"`

	// VGA colors
	VGAColors bool `short:"v" help:"Use VGA colors (splitans palette)"`

	// Version flag
	Version kong.VersionFlag `short:"V" help:"Show version"`
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("ansi-compositor"),
		kong.Description("Compose ANSI art from multiple sources using a YAML configuration"),
		kong.UsageOnError(),
		kong.Vars{
			"version": fmt.Sprintf("%s (%s)", version, commit),
		},
	)

	if err := run(&cli); err != nil {
		ctx.Fatalf("error: %v", err)
	}
}

func run(cli *CLI) error {
	// Load configuration
	cfg, err := config.Load(cli.Config)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Apply CLI overrides
	if cli.Output != "" {
		cfg.Output.File = cli.Output
	}
	if cli.Format != "" {
		cfg.Output.Format = cli.Format
	}
	if cli.Inline {
		cfg.Output.Inline = true
	}
	if cli.KeepTrailingLines {
		cfg.Output.KeepTrailingLines = true
	}
	if cli.VGAColors {
		cfg.Term.UseVGAColors = true
	}

	// Create compositor and compose
	comp := compositor.New(cfg)
	if err := comp.Compose(); err != nil {
		return fmt.Errorf("compose: %w", err)
	}

	// Export
	output, err := comp.Export()
	if err != nil {
		return fmt.Errorf("export: %w", err)
	}

	// Write output
	if cfg.Output.File != "" {
		if err := os.WriteFile(cfg.Output.File, []byte(output), 0644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	} else {
		fmt.Print(output)
	}

	return nil
}
