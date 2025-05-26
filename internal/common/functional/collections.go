// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

//nolint:varnamelen // Short parameter names are idiomatic in functional programming
package functional

import "errors"

// MapSlice applies a function to each element of a slice and returns a new slice with the results.
// This follows functional programming principles by creating a new slice instead of modifying the input.
func MapSlice[T, U any](items []T, fn func(T) U) []U {
	if items == nil {
		return nil
	}

	result := make([]U, len(items))
	for i, item := range items {
		result[i] = fn(item)
	}

	return result
}

// FilterSlice returns a new slice containing only elements that satisfy the predicate.
// This follows functional programming principles by creating a new slice instead of modifying the input.
func FilterSlice[T any](items []T, predicate func(T) bool) []T {
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

// Reduce applies a function against an accumulator and each element in the slice to reduce it to a single value.
func Reduce[T, U any](items []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, item := range items {
		result = fn(result, item)
	}

	return result
}

// Find returns the first element that satisfies the predicate, wrapped in a Result.
// If no element is found, returns an error.
func Find[T any](items []T, predicate func(T) bool) Result[T] {
	for _, item := range items {
		if predicate(item) {
			return Ok(item)
		}
	}

	return Err[T](errors.New("no element found matching predicate"))
}

// All returns true if all elements satisfy the predicate.
func All[T any](items []T, predicate func(T) bool) bool {
	for _, item := range items {
		if !predicate(item) {
			return false
		}
	}

	return true
}

// Any returns true if any element satisfies the predicate.
func Any[T any](items []T, predicate func(T) bool) bool {
	for _, item := range items {
		if predicate(item) {
			return true
		}
	}

	return false
}

// Partition splits a slice into two slices based on a predicate.
// Returns (matching, non-matching) slices.
func Partition[T any](items []T, predicate func(T) bool) ([]T, []T) {
	if items == nil {
		return nil, nil
	}

	trues := make([]T, 0)
	falses := make([]T, 0)

	for _, item := range items {
		if predicate(item) {
			trues = append(trues, item)
		} else {
			falses = append(falses, item)
		}
	}

	return trues, falses
}

// GroupBy groups elements by a key function.
func GroupBy[T any, K comparable](items []T, keyFn func(T) K) map[K][]T {
	if items == nil {
		return nil
	}

	result := make(map[K][]T)

	for _, item := range items {
		key := keyFn(item)
		result[key] = append(result[key], item)
	}

	return result
}

// Unique returns a new slice with duplicate elements removed.
// Elements must be comparable.
func Unique[T comparable](items []T) []T {
	if items == nil {
		return nil
	}

	seen := make(map[T]bool)
	result := make([]T, 0)

	for _, item := range items {
		if !seen[item] {
			seen[item] = true

			result = append(result, item)
		}
	}

	return result
}

// Zip combines two slices into a slice of pairs.
// The result has the length of the shorter input slice.
func Zip[T, U any](items1 []T, items2 []U) []Pair[T, U] {
	minLen := len(items1)
	if len(items2) < minLen {
		minLen = len(items2)
	}

	result := make([]Pair[T, U], minLen)
	for i := 0; i < minLen; i++ {
		result[i] = Pair[T, U]{First: items1[i], Second: items2[i]}
	}

	return result
}

// Pair represents a pair of values.
type Pair[T, U any] struct {
	First  T
	Second U
}

// Take returns the first n elements of the slice.
func Take[T any](items []T, n int) []T {
	if items == nil || n <= 0 {
		return []T{}
	}

	if n >= len(items) {
		// Return a copy of the entire slice
		result := make([]T, len(items))
		copy(result, items)

		return result
	}

	result := make([]T, n)
	copy(result, items[:n])

	return result
}

// Drop returns a new slice with the first n elements removed.
func Drop[T any](items []T, n int) []T {
	if items == nil || n <= 0 {
		// Return a copy of the entire slice
		result := make([]T, len(items))
		copy(result, items)

		return result
	}

	if n >= len(items) {
		return []T{}
	}

	result := make([]T, len(items)-n)
	copy(result, items[n:])

	return result
}

// AppendSlice creates a new slice with the additional elements appended.
// This follows functional programming by not modifying the original slice.
func AppendSlice[T any](items []T, newItems ...T) []T {
	result := make([]T, len(items), len(items)+len(newItems))
	copy(result, items)

	return append(result, newItems...)
}

// PrependSlice creates a new slice with the elements prepended.
func PrependSlice[T any](items []T, newItems ...T) []T {
	result := make([]T, len(newItems)+len(items))
	copy(result, newItems)
	copy(result[len(newItems):], items)

	return result
}
