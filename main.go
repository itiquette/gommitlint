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
	// Set global log level to panic to avoid any logging during initial setup
	// The logger will be properly configured in InitLogger later
	zerolog.SetGlobalLevel(zerolog.PanicLevel)

	// IMPORTANT: Use Stderr for logs to separate them from regular report output
	stdlog.Logger = stdlog.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04PM",
		NoColor:    false,
	})

	// Create the root context
	ctx := context.Background()

	// Add default CLI options to context
	ctx = domain.WithCLIOptions(ctx, domain.DefaultCLIOptions())

	// Add logger to the context for future use
	logger := stdlog.Logger
	ctx = logger.WithContext(ctx)

	// Create dependencies
	// This is the composition root where all dependencies are wired together

	// Create config provider using the context
	// Use the config provider which provides configuration
	configProvider, err := config.NewProvider()
	if err != nil {
		customlog.Logger(ctx).Error().Err(err).Msg("Failed to create configuration provider")
		os.Exit(1)
	}

	// Add provider to context
	ctx = config.WithProviderInContext(ctx, configProvider)

	// Execute root command with our dependencies
	// This approach allows us to replace dependencies for testing
	cli.ExecuteWithDependencies(
		ctx,
		version,
		commit,
		date,
		&cli.AppDependencies{
			ConfigProvider: configProvider,
		},
	)
}
