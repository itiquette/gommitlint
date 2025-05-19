// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package contextkeys defines shared context keys used throughout the application.
//
// This package uses a hybrid approach to context key management:
// - Public keys defined here are used for cross-cutting concerns (logging, CLI options)
// - Private keys are used within packages for encapsulation (config, providers)
//
// This design follows hexagonal architecture principles by:
// - Preventing direct access to package internals via context
// - Forcing interaction through defined interfaces
// - Maintaining clear boundaries between architectural layers
//
// Note: Most configuration access uses private keys (e.g., configKey{} in common/config)
// to enforce access through the GetConfig() interface rather than direct context access.
package contextkeys

// ContextKey is a type for context value keys to avoid collisions.
type ContextKey string

// Predefined context keys for cross-cutting concerns.
const (
	// LoggerKey is the context key for the logger.
	// Used by the logging adapters to store and retrieve loggers.
	LoggerKey ContextKey = "logger"

	// CLIOptionsKey is the context key for CLI options.
	// Used to pass CLI configuration through the application layers.
	CLIOptionsKey ContextKey = "cli_options"
)

// AllContextKeys returns a slice of all defined context keys.
// This is used for context merging operations in contextx package.
func AllContextKeys() []ContextKey {
	return []ContextKey{
		LoggerKey,
		CLIOptionsKey,
	}
}
