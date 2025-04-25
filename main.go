// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"os"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/infrastructure/git"
	"github.com/itiquette/gommitlint/internal/ports/cli"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// These variables are set by the build process.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Configure logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Create dependencies
	// This is the composition root where all dependencies are wired together
	configManager, err := config.New()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create configuration manager")
		os.Exit(1)
	}

	// Create a factory function that returns a domain.RepositoryFactory interface
	// This follows the Dependency Inversion Principle by depending on domain interfaces
	// instead of concrete implementations
	createRepositoryFactory := func(path string) (domain.RepositoryFactory, error) {
		return git.NewRepositoryFactory(path)
	}

	// Execute root command with our dependencies
	// This approach allows us to replace dependencies for testing
	cli.ExecuteWithDependencies(
		version,
		commit,
		date,
		configManager,
		createRepositoryFactory,
	)
}
