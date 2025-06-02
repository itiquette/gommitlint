// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package log

import (
	"context"
	"fmt"
	"os"
)

// StderrLogger is a simple logger that writes to stderr.
// It's used during initialization before the full logger is configured.
type StderrLogger struct{}

// Debug implements Logger.
func (StderrLogger) Debug(_ string, _ ...interface{}) {
	// Debug messages are suppressed in stderr logger
}

// Info implements Logger.
func (StderrLogger) Info(msg string, _ ...interface{}) {
	fmt.Fprintf(os.Stderr, "[INFO] %s\n", msg)
}

// Warn implements Logger.
func (StderrLogger) Warn(msg string, _ ...interface{}) {
	fmt.Fprintf(os.Stderr, "[WARN] %s\n", msg)
}

// Error implements Logger.
func (StderrLogger) Error(msg string, _ ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] %s\n", msg)
}

// WithContext implements Logger.
func (s StderrLogger) WithContext(ctx context.Context) context.Context {
	return ctx
}

// NewStderrLogger creates a new stderr logger.
func NewStderrLogger() StderrLogger {
	return StderrLogger{}
}
