// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"os"

	"github.com/itiquette/gommitlint/cmd"
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

	// Execute root command
	cmd.Execute(version, commit, date)
}
