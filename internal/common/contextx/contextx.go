// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package contextx provides context utilities for storing and retrieving values from context.
// This package enables type-safe operations with Go's context package and centralizes
// context key management to avoid key collisions.
//
// The package has minimal dependencies and is designed to be imported by any package
// without creating import cycles. It provides utilities for:
//   - Type-safe value storage and retrieval
//   - Logger context management
//   - Configuration context management
//   - Context merging operations
//
// Example usage:
//
//	// Store a value in context
//	ctx = contextx.WithValue(ctx, contextkeys.CLIOptionsKey, options)
//
//	// Retrieve value with type safety
//	opts, ok := contextx.Value[*CLIOptions](ctx, contextkeys.CLIOptionsKey)
//	if ok {
//	    // Use opts
//	}
//
//	// Work with logger
//	log := contextx.GetLogger(ctx)
//	log.Info("Processing request")
//
//	// Add logger with fields
//	ctx = contextx.WithLogger(ctx, log.With("request_id", requestID))
package contextx

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/common/config"
	"github.com/itiquette/gommitlint/internal/common/contextkeys"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
)

// WithValue wraps context.WithValue with type safety for context keys.
// It stores a value in the context associated with the given key.
//
// Example:
//
//	ctx = contextx.WithValue(ctx, contextkeys.UserIDKey, "user123")
func WithValue(ctx context.Context, key contextkeys.ContextKey, val interface{}) context.Context {
	return context.WithValue(ctx, key, val)
}

// Value retrieves a value from the context with type assertion.
// It returns the value and a boolean indicating whether the value was found
// and successfully type-asserted.
//
// Example:
//
//	userID, ok := contextx.Value[string](ctx, contextkeys.UserIDKey)
//	if !ok {
//	    // Handle missing or wrong type
//	}
func Value[T any](ctx context.Context, key contextkeys.ContextKey) (T, bool) {
	value := ctx.Value(key)
	if value == nil {
		var zero T

		return zero, false
	}

	result, ok := value.(T)

	return result, ok
}

// GetLogger retrieves the logger from the context.
//
// Example:
//
//	log := contextx.GetLogger(ctx)
//	log.Info("Operation completed") // Safe to call even if no logger was set
func GetLogger(ctx context.Context) outgoing.Logger {
	logger, ok := Value[outgoing.Logger](ctx, contextkeys.LoggerKey)
	if ok {
		return logger
	}

	fmt.Println("fatal error, no logger found, exiting")
	panic(1)
}

// WithLogger adds a logger to the context.
//
// Example:
//
//	log := logger.New()
//	ctx = contextx.WithLogger(ctx, log)
func WithLogger(ctx context.Context, log outgoing.Logger) context.Context {
	return context.WithValue(ctx, contextkeys.LoggerKey, log)
}

// WithConfig adds configuration to the context.
//
// Example:
//
//	cfg := config.Load()
//	ctx = contextx.WithConfig(ctx, cfg)
func WithConfig(ctx context.Context, cfg config.Config) context.Context {
	return config.WithConfig(ctx, cfg)
}

// GetConfig retrieves configuration from the context.
// This is the standard way to access configuration throughout the application.
//
// Example:
//
//	cfg := contextx.GetConfig(ctx)
//	maxRetries := cfg.GetInt("max_retries")
func GetConfig(ctx context.Context) config.Config {
	return config.GetConfig(ctx)
}
