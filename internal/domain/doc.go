// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package domain contains the core business logic and entities for gommitlint,
following hexagonal architecture principles.

This package is the heart of the application and contains:
  - Core domain entities (CommitInfo, ValidationResults, etc.)
  - Domain interfaces (Rule, CommitRepository, etc.)
  - Business logic (RulePriority, RuleRegistry)
  - Pure functions for domain operations

# Architecture Principles

The domain package follows these key principles:

1. Value Semantics: All domain entities use value semantics and immutability
2. No External Dependencies: The domain layer has no dependencies on infrastructure
3. Pure Functions: Domain logic is implemented as pure functions where possible
4. Interface Segregation: Interfaces are minimal and focused

# Key Components

  - CommitInfo: Core entity representing a git commit
  - Rule: Interface for all validation rules
  - ValidationResults: Immutable results of validation operations
  - RuleRegistry: Manages rule creation and lifecycle
  - RulePriorityService: Determines which rules are active

# Design Patterns

The domain layer uses several patterns:
  - Repository pattern for data access (interfaces only)
  - Factory pattern for rule creation
  - Functional options for configuration
  - Value objects for immutable data

All entities and services in this package focus on the "what" of the business
logic, not the "how" of the implementation details.
*/
package domain
