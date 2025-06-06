// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// This file provides CLI-specific context management.
// This consolidates context keys and utilities used only by the CLI layer,
// by keeping context concerns within the CLI adapter.
package cli

import (
	stdcontext "context"
)

// ContextKey is a type for context value keys to avoid collisions.
//

type ContextKey string

// Predefined context keys for CLI concerns.
const (
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
