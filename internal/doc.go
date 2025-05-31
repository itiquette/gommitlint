// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package internal contains the private application code for gommitlint.
// This package follows hexagonal architecture principles with clear separation
// between domain logic, adapters, and application services.
//
// # Structure
//
// The internal package is organized into the following sub-packages:
//
//   - domain: Core business logic and entities (the hexagon center)
//   - adapters: External integrations (hexagon edges)
//   - config: Configuration types and loading
//   - common: Shared utilities and helpers
//   - integrationtest: Integration tests
//   - testutils: Test utilities and helpers
//
// # Architecture
//
// This package implements hexagonal architecture (ports and adapters) where:
//   - The domain package contains pure business logic with no external dependencies
//   - Adapters implement specific technologies (git, CLI, logging, etc.)
//   - Port interfaces are defined where they are consumed, not in a central location
//   - Dependencies flow inward: adapters depend on domain, never the reverse
//
// For detailed architecture documentation, see docs/ARCHITECTURE.md
package internal
