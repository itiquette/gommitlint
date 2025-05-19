// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package formatting provides pure functions for formatting validation results.

This package contains formatting utilities that transform domain objects into
human-readable output. All functions are pure and stateless, following
functional programming principles.

The formatting functions support various output formats including:
  - Concise result messages
  - Verbose result messages with detailed error information
  - Help text generation based on rule violations

All formatting functions operate on immutable domain entities and return
formatted strings without side effects.
*/
package formatting
