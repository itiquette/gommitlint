// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"context"
	"os"

	"github.com/itiquette/gommitlint/internal/adapters/incoming/cli"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/log"
	"github.com/itiquette/gommitlint/internal/application/options"
	"github.com/itiquette/gommitlint/internal/composition"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/rs/zerolog"
	stdlog "github.com/rs/zerolog/log"
)

// These variables are set by the build process.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Configure initial basic logging
	// Set global log level to panic to avoid any logging during initial setup
	// The logger will be properly configured in InitLogger later
	zerolog.SetGlobalLevel(zerolog.PanicLevel)

	// IMPORTANT: Use Stderr for logs to separate them from regular report output
	stdlog.Logger = stdlog.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04PM",
		NoColor:    false,
	})

	// Create the root context - this is the ONLY context.Background() in the application
	ctx := context.Background()

	// Add logger to the context for future use - this will be replaced by a proper logger in cli.ExecuteWithContext
	logger := stdlog.Logger
	ctx = logger.WithContext(ctx)

	// No need to register domain logger provider anymore
	// Logger is accessed via context

	// Add default CLI options to context
	defaultOptions := options.CLIOptions{
		Verbosity:           "brief",
		Quiet:               false,
		VerbosityWithCaller: false,
		OutputFormat:        "text",
	}
	ctx = options.WithCLIOptions(ctx, defaultOptions)

	// Create config manager
	configManager, err := config.NewManager(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create config manager")
	}

	// Create logger adapter
	loggerAdapter := log.NewSimpleAdapter(logger)

	// Create composition root with dependencies
	root := composition.NewRoot(loggerAdapter, configManager.GetConfig())

	// Pass the context and composition root to cli.ExecuteWithContext
	// All dependencies and configuration are now set up
	cli.ExecuteWithContext(ctx, version, commit, date, root)
}
