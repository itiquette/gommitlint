// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package slices provides functional programming utilities for working with slices and maps.
// It includes common operations like Map, Filter, Reduce, and more, using generic types
// for type safety and reusability.
// This implementation uses the Go standard library slices and maps packages where possible.
package slices

// Map applies a function to each element of a slice and returns a new slice with the results.
// Note: The Go standard library does not currently have a direct equivalent for this function.
// This implementation follows functional programming principles and value semantics.
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

// FilterMap combines Filter and Map operations in a single pass.
// It applies a predicate to each element, then maps only the elements that satisfy the predicate.
// This is more efficient than separate Filter and Map operations when you need both.
// Note: The Go standard library does not have a direct equivalent for this combined operation.
// It would require two separate operations: slices.Filter followed by a Map operation.
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

// MapKeys returns all keys from a map as a slice.
// While Go 1.23+ provides maps.Keys, it returns an iterator.
// This function returns a slice for easier use in functional pipelines.
func MapKeys[K comparable, V any](inputMap map[K]V) []K {
	if inputMap == nil {
		return nil
	}

	result := make([]K, 0, len(inputMap))
	for k := range inputMap {
		result = append(result, k)
	}

	return result
}
