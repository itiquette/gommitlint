// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package git provides Git repository adapters for gommitlint.

This package implements Git repository operations for commit validation. It provides:

  - Repository access abstraction
  - Commit retrieval and analysis
  - Branch comparison operations

The package uses go-git library internally but exposes only domain
interfaces, ensuring the domain remains independent of Git implementation
details.

Key components:

  - repository.go: Main repository adapter implementing domain interfaces

The adapter implements multiple domain interfaces:
  - CommitRepository: For basic commit access
  - RepositoryInfoProvider: For repository metadata
  - CommitAnalyzer: For advanced analysis operations

All Git-specific logic is encapsulated within this package.
*/
package git
