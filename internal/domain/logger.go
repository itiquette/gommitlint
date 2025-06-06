// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

// Logger defines logging interface for domain layer.
// This interface allows domain functions to log without depending on specific logging implementations.
type Logger interface {
	// Log outputs a message with the specified level and optional arguments.
	Log(level string, msg string, args ...interface{})

	// Debug outputs a debug-level message with optional arguments.
	Debug(msg string, args ...interface{})

	// Info outputs an info-level message with optional arguments.
	Info(msg string, args ...interface{})

	// Error outputs an error-level message with optional arguments.
	Error(msg string, args ...interface{})
}
