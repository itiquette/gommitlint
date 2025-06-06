// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package logging provides a simple logging adapter for gommitlint.

This package provides a logging adapter that implements the domain.Logger
interface using zerolog.

Key features:
  - Single adapter implementation (no overengineering)
  - Consistent key-value argument handling
  - CLI integration for configuration

Usage:

	// Create logger from zerolog
	logger := logging.New(zerologInstance)

	// Use in domain functions
	logger.Info("Processing commit", "hash", commit.Hash)

	// Initialize from CLI context
	ctx = logging.InitLogger(ctx, cmd, "text")

The adapter ensures that domain logic can log messages without
depending on specific logging implementations.
*/
package logging
