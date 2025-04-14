// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package internal contains the private application code for gommitlint.
// This package follows functional hexagonal architecture principles with
// pure functions, value semantics, and explicit dependencies.
//
// # Structure
//
// The internal package is organized into the following sub-packages:
//
//   - domain: Core business logic and entities (the hexagon center)
//   - adapters: External integrations (hexagon edges)
//   - integrationtest: End-to-end integration tests
//
// # Functional Hexagonal Architecture
//
// This package implements functional hexagonal architecture where:
//   - Domain logic uses pure functions with no external dependencies
//   - All dependencies are passed as explicit function parameters
//   - Adapters implement domain interfaces using value semantics
//   - Configuration flows through explicit parameters, never context
//   - No service objects or dependency injection frameworks
//
// # Core Principles
//
//   - Pure Functions: Validation logic is implemented as pure functions
//   - Value Semantics: All data structures use value semantics for immutability
//   - Explicit Dependencies: Dependencies passed as parameters, never hidden
//   - Simple Composition: Complex operations built from simple function composition
//
// For detailed architecture documentation, see docs/ARCHITECTURE.md
package internal
