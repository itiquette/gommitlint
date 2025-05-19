// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package slices provides functional programming utilities for working with slices and maps.
// It includes common operations like Map, Filter, Reduce, and more, using generic types
// for type safety and reusability.
package slices

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

// Some returns true if at least one element satisfies the predicate.
func Some[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if predicate(item) {
			return true
		}
	}

	return false
}

// FindIndex returns the index of the first element in the slice that satisfies the predicate.
func FindIndex[T any](slice []T, predicate func(T) bool) int {
	for i, item := range slice {
		if predicate(item) {
			return i
		}
	}

	return -1
}

// Find returns the first element in the slice that satisfies the predicate.
func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}

	var zero T

	return zero, false
}

// Range creates a slice of integers from 0 to n-1.
func Range(n int) []int {
	result := make([]int, n)
	for i := range result {
		result[i] = i
	}

	return result
}

// Every applies a predicate to every element in a slice and returns true if all elements satisfy it.
func Every[T any](items []T, predicate func(T) bool) bool {
	for _, item := range items {
		if !predicate(item) {
			return false
		}
	}

	return true
}

// FilterMap combines Filter and Map operations in a single pass.
// It applies a predicate to each element, then maps only the elements that satisfy the predicate.
// This is more efficient than separate Filter and Map operations when you need both.
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

	result := make(map[K]V, len(inputMap))

	for key, value := range inputMap {
		exclude := false

		for _, excludeKey := range excludeKeys {
			if key == excludeKey {
				exclude = true

				break
			}
		}

		if !exclude {
			result[key] = value
		}
	}

	return result
}
