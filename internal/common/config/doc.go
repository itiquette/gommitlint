// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides common interfaces and utilities for configuration management.
// This package defines the core Config interface and provides context-based access
// patterns for configuration throughout the application.
//
// The package uses private context keys to enforce proper encapsulation and access
// patterns. Configuration must be accessed through the GetConfig() function rather
// than direct context access, following hexagonal architecture principles.
//
// Design Principles:
//   - Minimal interface for maximum flexibility
//   - Context-based access for clean dependency injection
//   - Private keys to enforce proper access patterns
//   - Empty implementations for safe fallbacks
//
// Usage:
//
//	ctx = config.WithConfig(ctx, myConfig)
//	cfg := config.GetConfig(ctx)
//	value := cfg.GetString("my.key")
package config
