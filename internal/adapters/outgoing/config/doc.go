// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package config provides configuration adapters for gommitlint.

This package implements the outgoing port for configuration management,
following hexagonal architecture principles. It provides:

  - Configuration loading from various sources
  - Adapter pattern implementation for config access
  - Service for managing configuration lifecycle
  - Key-based configuration value retrieval

The configuration adapter provides a unified interface for accessing
configuration values throughout the application, abstracting away the
specifics of configuration storage and retrieval.

Key components:

  - adapter.go: Main configuration adapter implementing config interfaces
  - service.go: Configuration service for loading and managing config
  - loader.go: Configuration loader for reading from files/environment

The adapter pattern ensures that domain logic remains independent of
configuration implementation details.
*/
package config
