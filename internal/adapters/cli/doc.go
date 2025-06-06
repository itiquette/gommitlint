// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package cli provides the command-line interface adapter for gommitlint.

This package implements the incoming port for CLI interactions. It handles:

  - Command parsing and validation
  - Parameter conversion to domain types
  - Hook installation and removal
  - Direct validation service calls

The package acts as the primary user interface, translating CLI commands
into domain operations while maintaining clear separation of concerns.

Key components:

  - rootcmd.go: Main command structure and initialization
  - validatecmd.go: Validation command implementation
  - installhookcmd.go/removehookcmd.go: Git hook management
  - validateparams.go: Parameter parsing and conversion
  - validation.go: Core validation orchestration

All CLI-specific logic is contained within this package, ensuring the
domain remains free from UI concerns.
*/
package cli
