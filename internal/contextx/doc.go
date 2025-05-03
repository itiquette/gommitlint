// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package contextx provides utility functions and helpers for working with Go contexts
// and functional programming patterns.
//
// This package contains the following major components:
//
// # Context Utilities
//
// Functions for working with context.Context, including value extraction and context chaining.
//
// # Functional Utilities
//
// Pure functions for working with slices and other data structures in a functional style:
//   - Filter: Create filtered slices based on a predicate
//   - Map: Transform slices by applying a function to each element
//   - Reduce: Accumulate values across a slice
//   - DeepCopy: Create immutable copies of data structures
//
// These utilities promote a functional programming style with value semantics,
// ensuring immutability and side-effect-free operations.
package contextx
