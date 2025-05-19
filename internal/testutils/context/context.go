// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package context provides test utilities for context creation
package context

import "context"

// CreateTestContext creates a new context for testing.
// This is the only place in test code where context.Background() should be called.
func CreateTestContext() context.Context {
	return context.Background()
}
