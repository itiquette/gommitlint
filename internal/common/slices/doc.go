// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package slices provides functional programming utilities for working with slices and maps.
// It offers a collection of generic functions that enable functional-style operations
// on collections, promoting immutability and value semantics.
//
// Functions include:
//   - Map: Transform elements of a slice
//   - Filter: Select elements matching a predicate
//   - Reduce: Aggregate elements into a single value
//   - Find: Locate first matching element
//   - Every/Some: Test all/any elements against a predicate
//   - Contains: Check if slice contains a value
//   - FilterMap: Combined filter and map operation
//   - DeepCopy: Create independent copies of slices and maps
//
// All functions follow value semantics and do not modify input data.
//
// Example usage:
//
//	// Transform a slice of strings to their lengths
//	lengths := slices.Map(strings, func(s string) int { return len(s) })
//
//	// Filter only positive numbers
//	positives := slices.Filter(numbers, func(n int) bool { return n > 0 })
//
//	// Sum all numbers
//	sum := slices.Reduce(numbers, 0, func(acc, n int) int { return acc + n })
package slices
