// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package log provides logging adapters for gommitlint.

This package implements the outgoing port for logging operations,
following hexagonal architecture principles. It provides:

  - Structured logging with zerolog
  - Context-based logger management
  - CLI-specific logging configuration
  - Domain logger adapter implementation

The package bridges the gap between the domain's logging interface
and the concrete zerolog implementation, ensuring proper separation
of concerns.

Key components:

  - logger.go: Main logger implementation with zerolog
  - domain_adapter.go: Adapter implementing domain.Logger interface
  - cli_options.go: CLI-specific logging configuration
  - init.go: Logger initialization and registration
  - zerolog_adapter.go: Zerolog-specific adapter implementation

The logging adapter ensures that domain logic can log messages without
depending on specific logging implementations.
*/
package log
