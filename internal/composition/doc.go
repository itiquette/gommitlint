// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package composition provides the composition root for the application,
following hexagonal architecture principles.

# Overview

The composition package is responsible for:
  - Wiring up all dependencies according to the Dependency Inversion Principle
  - Creating and configuring adapters for both incoming and outgoing ports
  - Managing the application lifecycle and initialization
  - Providing a single source of truth for dependency resolution

# Architecture

The package follows these architectural patterns:

1. Composition Root Pattern:
  - All dependency injection happens in this package
  - The Root struct acts as the central wiring point
  - Dependencies flow from the composition root outward

2. Factory Pattern:
  - Separate factories for incoming and outgoing adapters
  - Encapsulates the complexity of creating adapters
  - Ensures proper configuration of each component

3. Adapter Pattern:
  - Adapts between different interfaces across architectural boundaries
  - GitRepositoryAdapter combines multiple Git interfaces
  - ValidationServiceAdapter bridges validation interfaces

# Structure

  - root.go: Main composition root with dependency getters
  - initialization.go: Phased initialization of components
  - factories.go: Factory classes for creating adapters
  - git_repository.go: Composite Git repository implementation
  - service_adapter.go: Validation service interface adapter

# Usage

The composition root is created and initialized in main.go:

	root := composition.NewRoot()
	if err := root.Initialize(ctx); err != nil {
		// Handle error
	}

	// Get dependencies for CLI
	cliDeps := root.GetCLIDependencies()

# Extension Points

To add new components:

1. Define the interface in the appropriate domain/ports package
2. Implement the adapter in the adapters package
3. Add creation logic to the appropriate factory
4. Wire it up in the initialization phase
5. Add a getter method to the Root struct

# Design Decisions

- Separate factories for incoming/outgoing adapters maintain clear boundaries
- Phased initialization ensures proper dependency ordering
- Composite pattern for Git repository provides unified interface
- Context-based configuration allows runtime flexibility
*/
package composition
