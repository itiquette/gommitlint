// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides pure configuration types for gommitlint.
//
// This package contains only data structures with no behavior,
// following the principle of separating data from logic.
// All configuration types use value semantics for immutability.
//
// Configuration is loaded by the loader adapter and passed
// throughout the application as immutable values.
package config
