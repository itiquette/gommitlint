// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package slices provides functional programming utilities for working with slices and maps.
// It includes common operations like Map, Filter, Reduce, and more, using generic types
// for type safety and reusability.
// This implementation uses the Go standard library slices and maps packages where possible.
package slices

import (
	"maps"   // Go 1.21+
	"slices" // Go 1.21+
)

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
func Filter[T any](items []T, predicate func(T) bool) []T {
	if items == nil {
		return nil
	}

	// Pre-allocate with capacity of original slice
	// (in worst case all items match predicate)
	result := make([]T, 0, len(items))

	// Add only items that match the predicate
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
// This function has minimal usage in the codebase and is maintained for backward compatibility.
// Consider using a standard for loop for better clarity in new code.
func ForEach[T any](items []T, fn func(T)) {
	for _, item := range items {
		fn(item)
	}
}

// Contains checks if a slice contains a specific value.
// Uses the standard library slices package under the hood.
func Contains[T comparable](items []T, value T) bool {
	return slices.Contains(items, value)
}

// Note: We've removed the DeepCopy function as it was just a wrapper around slices.Clone.
// Please use slices.Clone directly instead.

// DeepCopyMap returns a deep copy of a map.
// Uses the standard library maps package under the hood.
func DeepCopyMap[K comparable, V any](inputMap map[K]V) map[K]V {
	return maps.Clone(inputMap)
}

// Some returns true if at least one element satisfies the predicate.
func Some[T any](slice []T, predicate func(T) bool) bool {
	return slices.IndexFunc(slice, predicate) >= 0
}

// FindIndex returns the index of the first element in the slice that satisfies the predicate.
// Uses the standard library slices package under the hood.
func FindIndex[T any](slice []T, predicate func(T) bool) int {
	return slices.IndexFunc(slice, predicate)
}

// Find returns the first element in the slice that satisfies the predicate.
func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	idx := slices.IndexFunc(slice, predicate)
	if idx >= 0 {
		return slice[idx], true
	}

	var zero T

	return zero, false
}

// Note: The Range function has been removed as it was unused in production code.
// If you need to create a slice of sequential integers, use a for loop directly:
//
//	result := make([]int, n)
//	for i := range result {
//		result[i] = i
//	}

// Every applies a predicate to every element in a slice and returns true if all elements satisfy it.
func Every[T any](items []T, predicate func(T) bool) bool {
	return slices.IndexFunc(items, func(item T) bool {
		return !predicate(item)
	}) < 0
}

// FilterMap combines Filter and Map operations in a single pass.
// It applies a predicate to each element, then maps only the elements that satisfy the predicate.
// This is more efficient than separate Filter and Map operations when you need both.
// This function has limited usage in the codebase, primarily in error formatting.
func FilterMap[T, U any](items []T, predicate func(T) bool, mapFn func(T) U) []U {
	if items == nil {
		return nil
	}

	result := make([]U, 0, len(items))

	for _, item := range items {
		if predicate(item) {
			result = append(result, mapFn(item))
		}
	}

	return result
}

// FilterMapKeys creates a new map excluding specified keys.
// This is useful for creating a copy of a map without certain keys.
func FilterMapKeys[K comparable, V any](inputMap map[K]V, excludeKeys []K) map[K]V {
	if inputMap == nil {
		return nil
	}

	// Clone the map first
	result := maps.Clone(inputMap)

	// Delete the excluded keys
	for _, key := range excludeKeys {
		delete(result, key)
	}

	return result
}
