// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"context"
	"os"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	customlog "github.com/itiquette/gommitlint/internal/infrastructure/log"
	"github.com/itiquette/gommitlint/internal/ports/cli"
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
	stdlog.Logger = stdlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Create the root context
	ctx := context.Background()

	// Add default CLI options to context
	ctx = domain.WithCLIOptions(ctx, domain.DefaultCLIOptions())

	// Add logger to the context for future use
	logger := stdlog.Logger
	ctx = logger.WithContext(ctx)

	// Create dependencies
	// This is the composition root where all dependencies are wired together

	// Create config manager using the context
	configManager, err := config.NewManager(ctx)
	if err != nil {
		customlog.Logger(ctx).Error().Err(err).Msg("Failed to create configuration manager")
		os.Exit(1)
	}

	// Execute root command with our dependencies
	// This approach allows us to replace dependencies for testing
	cli.ExecuteWithDependencies(
		ctx,
		version,
		commit,
		date,
		configManager,
	)
}
