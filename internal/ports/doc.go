// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package ports defines all the port interfaces for the hexagonal architecture.
// These interfaces represent the boundaries between the core domain and the outside world.
//
// This package centralizes all adapter-facing port interfaces to eliminate duplication
// across adapter packages. Domain interfaces remain in the domain package.
//
// Port interfaces defined here:
//   - Logger: Used by all adapters for structured logging
//   - ValidationService: Used by CLI adapter for validation operations
//
// Domain interfaces (defined in domain package):
//   - Repository: Git repository operations
//   - CryptoVerifier: Cryptographic verification
//   - Formatter: Output formatting
//
// This approach provides clean boundaries while avoiding interface duplication.
package ports
