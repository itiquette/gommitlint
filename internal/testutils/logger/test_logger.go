// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package logger provides utility functions for logging in tests.
// This package is intended for testing purposes only.
package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

// NewTestLogger creates a new zerolog.Logger suitable for testing.
// It can be used in tests as a drop-in replacement for production loggers.
// The logger writes to a discarded writer by default, but can be configured
// to write to a file or other io.Writer for debugging.
func NewTestLogger() *zerolog.Logger {
	// Discard all output by default
	writer := io.Discard

	// Check if TEST_DEBUG is set to enable logging to stderr
	if os.Getenv("TEST_DEBUG") == "1" {
		writer = os.Stderr
	}

	// Create a logger with minimal context
	logger := zerolog.New(writer).With().Timestamp().Logger()

	// Set level to Trace to capture everything
	logger = logger.Level(zerolog.TraceLevel)

	return &logger
}
