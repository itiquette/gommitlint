// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal/adapters/cli/commands"
	logadapter "github.com/itiquette/gommitlint/internal/adapters/logging"
	"github.com/urfave/cli/v3"
)

// These variables are set by the build process.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// preprocessVerboseArgs converts -vv to -v -v to work with the CLI library.
func preprocessVerboseArgs(args []string) []string {
	var processed []string

	for _, arg := range args {
		if arg == "-vv" {
			// Convert -vv to -v -v so the CLI library can handle it
			processed = append(processed, "-v", "-v")
		} else {
			processed = append(processed, arg)
		}
	}

	return processed
}

func main() {
	// Create the root context
	ctx := context.Background()

	// Initialize logger early in the application flow
	ctx = logadapter.InitLogger(ctx, "text", false, "info") // Basic logger setup

	// Preprocess arguments to handle -vv flag
	args := preprocessVerboseArgs(os.Args)

	app := &cli.Command{
		Name:  "gommitlint",
		Usage: "Git commit message validator",
		UsageText: `gommitlint [flags] <command> [args]

Examples:
  gommitlint validate                        # Validate HEAD commit
  gommitlint validate --base-branch=main     # Validate branch commits
  gommitlint install-hook                    # Install commit-msg hook
  gommitlint config show --format=yaml > .gommitlint.yaml # Generate config file`,
		Version: fmt.Sprintf("%s (Commit: %s, Build date: %s)", version, commit, date),

		// Enable shell completion for all supported shells
		EnableShellCompletion: true,

		// Hide version from help output (only show via --version flag)
		HideVersion: true,

		// Global flags
		Flags: []cli.Flag{
			// Configuration flags
			&cli.StringFlag{
				Name:     "gommitconfig",
				Usage:    "gommitlint config file `FILE`",
				Category: "Configuration",
			},
			&cli.BoolFlag{
				Name:     "ignore-config",
				Usage:    "ignore config files",
				Category: "Configuration",
			},

			// Repository flags
			&cli.StringFlag{
				Name:     "repo-path",
				Usage:    "repository `PATH` (defaults to current directory)",
				Category: "Repository",
			},

			// Output flags
			&cli.StringFlag{
				Name:     "format",
				Value:    "text",
				Usage:    "output `FORMAT` (text, json, github, gitlab)",
				Category: "Output",
			},
			&cli.StringFlag{
				Name:     "color",
				Value:    "auto",
				Usage:    "color `MODE` (auto, always, never)",
				Category: "Output",
			},
			&cli.StringFlag{
				Name:     "log-level",
				Value:    "info",
				Usage:    "log `LEVEL` (error, warn, info, debug, trace)",
				Category: "Output",
			},
			&cli.BoolFlag{
				Name:     "quiet",
				Aliases:  []string{"q"},
				Usage:    "suppress all output except errors",
				Category: "Output",
			},
			&cli.BoolFlag{
				Name:     "debug",
				Usage:    "enable debug output with source locations",
				Category: "Output",
			},
		},

		// Before hook for global setup
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// Setup logging based on flags
			output := cmd.String("format")
			debug := cmd.Bool("debug")
			logLevel := cmd.String("log-level")
			ctx = logadapter.InitLogger(ctx, output, debug, logLevel)

			return ctx, nil
		},

		Action: func(_ context.Context, cmd *cli.Command) error {
			// If no subcommand, show help
			return cli.ShowAppHelp(cmd)
		},

		Commands: []*cli.Command{
			commands.NewValidateCommand(),
			commands.NewConfigCommand(),
			commands.NewInstallHookCommand(),
			commands.NewRemoveHookCommand(),
		},
	}

	if err := app.Run(ctx, args); err != nil {
		// Get logger from context and handle error
		zerologLogger := logadapter.GetLogger(ctx)
		logger := logadapter.NewDomainLogger(zerologLogger)
		logger.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}
