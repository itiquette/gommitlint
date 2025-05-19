// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package git provides Git repository adapters for gommitlint.

This package implements the outgoing port for Git operations, following
hexagonal architecture principles. It provides:

  - Repository access abstraction
  - Commit retrieval and analysis
  - Branch comparison operations
  - Value-semantic repository adapter

The package uses go-git library internally but exposes only domain
interfaces, ensuring the domain remains independent of Git implementation
details.

Key components:

  - repository.go: Main repository adapter implementing domain interfaces
  - repository_factory.go: Factory for creating repository instances
  - repository_helpers.go: Helper functions for repository operations

The adapter implements multiple domain interfaces:
  - CommitRepository: For basic commit access
  - RepositoryInfoProvider: For repository metadata
  - CommitAnalyzer: For advanced analysis operations

All Git-specific logic is encapsulated within this package.
*/
package git
