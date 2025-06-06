// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal/adapters/cli"
	"github.com/itiquette/gommitlint/internal/adapters/config"
	logadapter "github.com/itiquette/gommitlint/internal/adapters/logging"
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

	// Initialize logger early in the application flow
	ctx = logadapter.InitLogger(ctx, nil, "text") // Basic logger setup

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Logger is already set up by log.InitLogger, no additional setup needed

	// Pass config to CLI with logger context
	cli.Execute(ctx, version, commit, date, cfg)
}
