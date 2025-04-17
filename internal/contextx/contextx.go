// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package contextx provides context utilities for the application.
package contextx

import "context"

// ContextKey is a type for context value keys to avoid collisions.
type ContextKey string

// Predefined context keys for the application.
const (
	// UserIDKey is the context key for user ID.
	UserIDKey ContextKey = "user_id"

	// TraceIDKey is the context key for trace ID.
	TraceIDKey ContextKey = "trace_id"

	// RepositoryPathKey is the context key for repository path.
	RepositoryPathKey ContextKey = "repository_path"

	// RevisionKey is the context key for git revision.
	RevisionKey ContextKey = "revision"

	// RuleNameKey is the context key for rule name.
	RuleNameKey ContextKey = "rule_name"
)

// WithValue wraps context.WithValue with type safety for our ContextKey type.
func WithValue(ctx context.Context, key ContextKey, val interface{}) context.Context {
	return context.WithValue(ctx, key, val)
}

// Value retrieves a value from the context with type assertion.
func Value[T any](ctx context.Context, key ContextKey) (T, bool) {
	value := ctx.Value(key)
	if value == nil {
		var zero T

		return zero, false
	}

	result, ok := value.(T)

	return result, ok
}
