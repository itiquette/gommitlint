// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package contextx

// NOTE: The basic functional utilities (Map, Filter, Reduce, ForEach, Contains, DeepCopy, DeepCopyMap)
// are now defined in contextx.go. This file contains additional utilities that build upon those.

// Every returns true if all elements in the slice satisfy the predicate.
// Returns true for empty slices.
func Every[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if !predicate(item) {
			return false
		}
	}

	return true
}

// Some returns true if at least one element in the slice satisfies the predicate.
// Returns false for empty slices.
func Some[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if predicate(item) {
			return true
		}
	}

	return false
}

// FindIndex returns the index of the first element in the slice that satisfies the predicate.
// If no element satisfies the predicate, it returns -1.
func FindIndex[T any](slice []T, predicate func(T) bool) int {
	for i, item := range slice {
		if predicate(item) {
			return i
		}
	}

	return -1
}

// Find returns the first element in the slice that satisfies the predicate.
// If no element satisfies the predicate, it returns the zero value of T and false.
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
// This is useful for functional-style iteration over a range of indices.
func Range(n int) []int {
	result := make([]int, n)
	for i := range result {
		result[i] = i
	}

	return result
}

// Keys returns a slice of all keys in a map.
func Keys[K comparable, V any](inputMap map[K]V) []K {
	if inputMap == nil {
		return nil
	}

	result := make([]K, 0, len(inputMap))
	for k := range inputMap {
		result = append(result, k)
	}

	return result
}

// FlatMap applies a function to each element of a slice that returns a slice,
// then flattens the result into a single slice.
func FlatMap[T, U any](items []T, mapFunction func(T) []U) []U {
	if items == nil {
		return nil
	}

	var result []U
	for _, item := range items {
		mapped := mapFunction(item)
		result = append(result, mapped...)
	}

	return result
}

// Unique returns a slice with duplicate elements removed.
// Order is preserved, with the first occurrence of each element being kept.
func Unique[T comparable](items []T) []T {
	if items == nil {
		return nil
	}

	if len(items) <= 1 {
		return DeepCopy(items)
	}

	seen := make(map[T]struct{}, len(items))
	result := make([]T, 0, len(items))

	for _, item := range items {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
