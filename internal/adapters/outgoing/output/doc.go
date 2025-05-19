// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package output provides output formatting adapters for gommitlint.

This package implements the outgoing port for formatting validation
results, following hexagonal architecture principles. It provides:

  - Multiple output format support (text, JSON, GitHub, GitLab)
  - Configurable formatting options
  - Color and symbol customization
  - CI/CD-specific output formats

The output adapters transform domain validation results into various
presentation formats suitable for different contexts (terminal, CI systems).

Key components:

  - text.go: Human-readable text formatter with colors and symbols
  - json.go: JSON formatter for machine-readable output
  - github.go: GitHub Actions-specific formatter
  - gitlab.go: GitLab CI-specific formatter

Each formatter implements the domain.ResultFormatter interface,
allowing the domain to remain independent of presentation concerns.
*/
package output
