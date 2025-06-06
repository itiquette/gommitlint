// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package domain contains the core business logic for gommitlint.

This package implements pure functions for commit validation following functional
hexagonal architecture principles. All validation logic is side-effect free and
uses explicit dependencies.

# Core Components

  - commit.go: Commit entity with value semantics
  - rules.go: Rule interfaces and core validation functions
  - validation.go: Pure validation functions that compose rules
  - verification.go: Signature verification domain logic
  - reporting.go: Result aggregation and reporting functions
  - errors.go: Domain-specific error types

# Functional Architecture

The domain follows these principles:

  - Pure Functions: All validation logic is implemented as pure functions
  - Value Semantics: All entities use value types, no pointers or mutation
  - Explicit Dependencies: All dependencies passed as function parameters
  - No I/O: Domain functions never perform I/O operations directly
  - Composition: Complex operations built from simple function composition

# Core Validation Function

All validation flows through the main function:

	func ValidateCommit(commit Commit, rules []Rule, repo Repository, cfg *config.Config) ValidationResult

This function composes individual rule validations into a complete result.

# Interfaces

The domain defines interfaces for external dependencies:

  - Repository: For Git operations (implemented by adapters/git)
  - Logger: For logging operations (implemented by adapters/logging)
  - Rule: For individual validation rules (implemented by domain/rules)

All interfaces are defined where they are consumed, maintaining clean dependency flow.
*/
package domain
