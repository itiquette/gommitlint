// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package contextx provides context utilities for the application.
package contextx

import (
	"context"

	"github.com/rs/zerolog"
)

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

	// LoggerKey is the context key for the logger.
	LoggerKey ContextKey = "logger"
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

// Logger retrieves a logger from the context or returns a default logger.
func Logger(ctx context.Context) *zerolog.Logger {
	if ctx == nil {
		return defaultLogger()
	}

	if logger, ok := ctx.Value(LoggerKey).(*zerolog.Logger); ok {
		return logger
	}

	return defaultLogger()
}

// WithLogger adds a logger to the context.
func WithLogger(ctx context.Context, logger *zerolog.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// defaultLogger returns a default zerolog logger.
func defaultLogger() *zerolog.Logger {
	logger := zerolog.New(zerolog.NewConsoleWriter()).
		With().
		Timestamp().
		Caller().
		Logger()

	return &logger
}

// LoggerWith returns a logger with additional context fields.
func LoggerWith(ctx context.Context, key string, value interface{}) *zerolog.Logger {
	logger := Logger(ctx)
	newLogger := logger.With().Interface(key, value).Logger()

	return &newLogger
}

// MergeContext combines two contexts, with values from the second context
// taking precedence over values from the first context.
func MergeContext(ctx1, ctx2 context.Context) context.Context {
	if ctx1 == nil {
		return ctx2
	}

	if ctx2 == nil {
		return ctx1
	}

	// Start with a copy of ctx1
	result := ctx1

	// Add all context keys from ctx2 to result
	for _, key := range contextKeys {
		if val := ctx2.Value(key); val != nil {
			result = context.WithValue(result, key, val)
		}
	}

	// Handle special configuration key
	// This ensures configuration is properly merged
	if val := ctx2.Value(ConfigContextKey); val != nil {
		result = context.WithValue(result, ConfigContextKey, val)
	}

	// Also handle OutputFormatContextKey and VerbosityContextKey
	if val := ctx2.Value(OutputFormatContextKey); val != nil {
		result = context.WithValue(result, OutputFormatContextKey, val)
	}

	if val := ctx2.Value(VerbosityContextKey); val != nil {
		result = context.WithValue(result, VerbosityContextKey, val)
	}

	return result
}

// List of all context keys for use in MergeContext.
var contextKeys = []ContextKey{
	UserIDKey,
	TraceIDKey,
	RepositoryPathKey,
	RevisionKey,
	RuleNameKey,
	LoggerKey,
	ConfigContextKey,
	OutputFormatContextKey,
	VerbosityContextKey,
}

// Predefined context keys for the configuration system.
const (
	// ConfigContextKey is the context key for gommitlint config.
	ConfigContextKey ContextKey = "gommitlint-config"

	// OutputFormatContextKey is the context key for output format.
	OutputFormatContextKey ContextKey = "output-format"

	// VerbosityContextKey is the context key for verbosity.
	VerbosityContextKey ContextKey = "verbosity"

	// ParentContextPreserveKey is used to mark contexts that should preserve parent values.
	ParentContextPreserveKey ContextKey = "preserve-parent-context"
)

// Additionally, we need to handle any non-ContextKey values that might be in the context
// This is used to ensure configuration values are properly merged

// Map applies a function to each element of a slice and returns a new slice with the results.
func Map[T, U any](items []T, mapFunction func(T) U) []U {
	if items == nil {
		return nil
	}

	result := make([]U, len(items))
	for i, item := range items {
		result[i] = mapFunction(item)
	}

	return result
}

// MapMap applies a function to each key-value pair of a map and returns a slice of the results.
func MapMap[K comparable, V, U any](inputMap map[K]V, mapFunction func(K, V) U) []U {
	if inputMap == nil {
		return nil
	}

	result := make([]U, 0, len(inputMap))
	for k, v := range inputMap {
		result = append(result, mapFunction(k, v))
	}

	return result
}

// Filter returns a new slice containing only the elements that satisfy the predicate.
// If the input slice is nil, nil is returned. If the input slice is empty or
// no elements match the predicate, an empty slice is returned.
func Filter[T any](items []T, predicate func(T) bool) []T {
	if items == nil {
		return nil
	}

	result := make([]T, 0)

	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}

	return result
}

// Reduce applies a function to each element of a slice, accumulating a result.
func Reduce[T, U any](items []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, item := range items {
		result = fn(result, item)
	}

	return result
}

// ForEach applies a function to each element of a slice without returning a result.
func ForEach[T any](items []T, fn func(T)) {
	for _, item := range items {
		fn(item)
	}
}

// Contains checks if a slice contains a specific value.
func Contains[T comparable](items []T, value T) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}

	return false
}

// DeepCopy returns a deep copy of a slice.
func DeepCopy[T any](items []T) []T {
	if items == nil {
		return nil
	}

	result := make([]T, len(items))
	copy(result, items)

	return result
}

// DeepCopyMap returns a deep copy of a map.
func DeepCopyMap[K comparable, V any](inputMap map[K]V) map[K]V {
	if inputMap == nil {
		return nil
	}

	result := make(map[K]V, len(inputMap))
	for k, v := range inputMap {
		result[k] = v
	}

	return result
}
