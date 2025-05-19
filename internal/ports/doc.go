// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package ports defines the port interfaces for hexagonal architecture.
//
// This package contains only interface definitions, no implementations.
// It is divided into:
//   - incoming: Driver ports that external adapters use to interact with the application
//   - outgoing: Driven ports that the application uses to interact with external systems
//
// All implementations of these interfaces should be placed in the adapters package.
package ports
