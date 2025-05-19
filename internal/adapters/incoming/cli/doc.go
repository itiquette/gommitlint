// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package cli provides the command-line interface adapter for gommitlint.

This package implements the incoming port for CLI interactions, following
hexagonal architecture principles. It handles:

  - Command parsing and validation
  - Parameter conversion to domain types
  - Hook installation and removal
  - Validation workflow orchestration

The package acts as the primary user interface, translating CLI commands
into domain operations while maintaining clear separation of concerns.

Key components:

  - root.go: Main command structure and initialization
  - validate.go: Validation command implementation
  - installhook.go/removehook.go: Git hook management
  - validationparameters.go: Parameter parsing and conversion
  - dependencies.go: Service factory and dependency injection

All CLI-specific logic is contained within this package, ensuring the
domain remains free from UI concerns.
*/
package cli
