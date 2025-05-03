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
