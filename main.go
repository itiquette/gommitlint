// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal/adapters/incoming/cli"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/log"
	"github.com/itiquette/gommitlint/internal/application/options"
	"github.com/itiquette/gommitlint/internal/composition"
	"github.com/itiquette/gommitlint/internal/config"
)

// These variables are set by the build process.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// The ONLY ctx that should flow through the application
	ctx := context.Background()

	// Add default CLI options to context
	defaultOptions := options.DefaultCLIOptions()
	ctx = options.WithCLIOptions(ctx, defaultOptions)

	configLoader, err := config.NewLoader(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config loader: %v\n", err)
		os.Exit(1)
	}

	// Create a simple stderr logger for initialization
	// The actual logger will be configured when CLI flags are parsed
	stderrLogger := log.NewStderrLogger()
	container := composition.NewContainer(stderrLogger, configLoader.GetConfig())

	cli.Execute(ctx, version, commit, date, container)
}
