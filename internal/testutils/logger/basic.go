// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package logger provides test utilities for logging.
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// InitBasicLogger initializes and returns a zerolog.Logger configured with sensible defaults for testing.
//
// Returns:
//   - zerolog.Logger: A basic logger instance
func InitBasicLogger() zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Logger()
}
