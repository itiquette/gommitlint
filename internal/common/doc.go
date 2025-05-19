// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package common provides shared utilities and interfaces that are used throughout
// the application. This package is designed to have minimal dependencies and can
// be imported by any other package without creating import cycles.
//
// The package follows these principles:
//   - Minimal dependencies to avoid import cycles
//   - Type-safe operations with generics where appropriate
//   - Clear separation of concerns
//   - Hexagonal architecture alignment with proper encapsulation
//
// Subpackages:
//   - config: Configuration interfaces and context-based access patterns
//   - contextkeys: Shared context keys for cross-cutting concerns
//   - contextx: Context utilities for type-safe value storage and retrieval
//   - slices: Functional programming utilities for working with slices and maps
//
// All packages in common are designed to be foundational utilities that other
// packages can safely depend on.
package common
