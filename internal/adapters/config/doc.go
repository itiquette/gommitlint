// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration loading and parsing adapters.
//
// This package handles the external concerns of loading configuration from
// files and environment variables, then converting them to domain config types.
//
// Files:
//   - loader.go: File loading and path resolution
//   - yaml.go: YAML parsing and unmarshaling
//   - env.go: Environment variable override support
//
// The package follows hexagonal architecture by providing adapters that
// convert external configuration sources into domain configuration objects.
package config
