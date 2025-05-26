// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal/adapters/incoming/cli"
	"github.com/itiquette/gommitlint/internal/application/options"
	"github.com/itiquette/gommitlint/internal/config"
)

// These variables are set by the build process.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Create the root context
	ctx := context.Background()

	// Add default CLI options
	ctx = options.WithCLIOptions(ctx, options.DefaultCLIOptions())

	// Create config loader
	configLoader, err := config.NewLoader(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config loader: %v\n", err)
		os.Exit(1)
	}

	// Pass config to CLI - container will be created after logger initialization
	cli.Execute(ctx, version, commit, date, configLoader.GetConfig())
}
