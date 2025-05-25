// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration management for gommitlint.
//
// It follows a value-based approach with immutable configuration structures
// and functional transformation methods. The package uses the adapter pattern
// to interface with external configuration sources while maintaining clean
// architecture boundaries.
//
// Architecture:
//   - types/: Core configuration data structures with value semantics
//   - Loader: Configuration loader using adapter pattern
//   - Rules extensions: Helper functions for rule configuration
//   - Priority system: Rule enable/disable priority logic
//
// Design Principles:
//   - Value semantics and immutability
//   - Functional transformations
//   - Context-based access throughout the application
//   - Clear separation between types and behavior
//
// Usage:
//
//	// Create a default configuration
//	cfg := config.NewDefaultConfig()
//
//	// Transform configuration values
//	cfg = cfg.WithSubject(cfg.Subject.WithMaxLength(80))
//
//	// Use the loader for loading from files
//	loader, err := config.NewLoader(ctx)
//	if err != nil {
//	    // handle error
//	}
//	cfg = loader.GetConfig()
package config
