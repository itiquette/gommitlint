// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package report provides report generation services for validation results.

This package is responsible for taking validation results from the domain
and transforming them into various output formats suitable for different
contexts (terminal display, CI systems, machine processing).

The report generator:
  - Orchestrates between validation results and output formatters
  - Manages report options and configuration
  - Handles output writing to specified destinations
  - Provides consistent error reporting across formats

The package follows the application service pattern, coordinating between
domain results and outgoing adapters (formatters) while maintaining proper
separation of concerns.
*/
package report
