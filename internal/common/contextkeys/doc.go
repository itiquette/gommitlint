// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package contextkeys defines shared context keys used throughout the application.
// It provides a centralized location for context key definitions to avoid collisions
// and ensure consistency across the codebase.
//
// The package uses a hybrid approach to context key management:
//   - Public keys defined here are used for cross-cutting concerns (logging, CLI options)
//   - Private keys are used within packages for encapsulation (config, providers)
//
// This design follows hexagonal architecture principles by:
//   - Preventing direct access to package internals via context
//   - Forcing interaction through defined interfaces
//   - Maintaining clear boundaries between architectural layers
//
// Most configuration access uses private keys to enforce access through proper
// interfaces rather than direct context access.
package contextkeys
