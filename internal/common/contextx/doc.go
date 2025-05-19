// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package contextx provides context utilities for type-safe value storage and retrieval.
// It wraps the standard context package with generic functions for better type safety
// and provides specialized functions for common cross-cutting concerns like logging
// and configuration.
//
// Features:
//   - Type-safe value storage and retrieval with generics
//   - Logger context management
//   - Configuration context management
//   - Context merging operations
//
// The package is designed to have minimal dependencies and can be imported by any
// package without creating import cycles.
//
// Example usage:
//
//	// Store a value with type safety
//	ctx = contextx.WithValue(ctx, contextkeys.UserIDKey, userID)
//
//	// Retrieve value with automatic type assertion
//	userID, ok := contextx.Value[string](ctx, contextkeys.UserIDKey)
//	if !ok {
//	    // Handle missing or wrong type
//	}
//
//	// Access configuration
//	cfg := contextx.GetConfig(ctx)
//	maxRetries := cfg.GetInt("max_retries")
package contextx
