// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package context provides test utilities for context creation
package context

import (
	"context"

	"github.com/itiquette/gommitlint/internal/common/contextkeys"
)

// CreateTestContext creates a new context for testing.
// This is the only place in test code where context.Background() should be called.
func CreateTestContext() context.Context {
	return context.Background()
}

// MergeContext combines two contexts, with values from the second context
// taking precedence over values from the first context. This is useful when
// you need to override specific context values while preserving others.
//
// The base context (ctx1) must be non-nil or this function will panic.
// If ctx2 is nil, ctx1 is returned unchanged.
func MergeContext(ctx1, ctx2 context.Context) context.Context {
	// Validate base context
	if ctx1 == nil {
		panic("MergeContext: base context cannot be nil")
	}

	// If second context is nil, return first unchanged
	if ctx2 == nil {
		return ctx1
	}

	// Start with base context
	result := ctx1

	// Copy all known context keys from override context
	for _, key := range contextkeys.AllContextKeys() {
		if val := ctx2.Value(key); val != nil {
			result = context.WithValue(result, key, val)
		}
	}

	return result
}
