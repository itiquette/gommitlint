// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package fsutils provides filesystem utility functions following functional programming
// principles. All functions are pure with no side effects, using value semantics.
//
// This package follows the hexagonal architecture by providing clean interfaces
// for filesystem operations that can be used throughout the application without
// introducing unnecessary dependencies.
//
// The utilities are divided into:
//   - Path operations: Functions that work with directories and paths
//   - File operations: Functions that work with file content and discovery
//
// All functions consistently use explicit error handling and maintain value semantics.
package fsutils
