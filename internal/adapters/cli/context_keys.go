// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// This file provides CLI-specific context management.
// This consolidates context keys and utilities used only by the CLI layer,
// following hexagonal architecture by keeping context concerns within the CLI adapter.
package cli

import (
	stdcontext "context"
)

// ContextKey is a type for context value keys to avoid collisions.
//

type ContextKey string

// Predefined context keys for CLI concerns.
const (
	// LoggerKey is the context key for the logger.
	LoggerKey ContextKey = "logger"

	// CLIOptionsKey is the context key for CLI options.
	CLIOptionsKey ContextKey = "cli_options"
)

// WithValue wraps context.WithValue with type safety for context keys.
func WithValue(ctx stdcontext.Context, key ContextKey, val interface{}) stdcontext.Context {
	return stdcontext.WithValue(ctx, key, val)
}

// Value retrieves a value from the context with type assertion.
func Value[T any](ctx stdcontext.Context, key ContextKey) (T, bool) {
	value := ctx.Value(key)
	if value == nil {
		var zero T

		return zero, false
	}

	result, ok := value.(T)

	return result, ok
}

// GetLogger retrieves the logger from the context.
// The logger is stored as interface{} to avoid coupling to specific logger types.
func GetLogger(ctx stdcontext.Context) interface{} {
	logger := ctx.Value(LoggerKey)
	if logger == nil {
		panic("logger not found in context - logger must be set early in application flow")
	}

	return logger
}

// WithLogger adds a logger to the context.
// The logger is stored as interface{} to allow different logger implementations.
func WithLogger(ctx stdcontext.Context, log interface{}) stdcontext.Context {
	return stdcontext.WithValue(ctx, LoggerKey, log)
}
