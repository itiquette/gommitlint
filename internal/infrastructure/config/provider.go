// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration loading and access.
package config

// No imports needed for now.

// Provider provides access to configuration.
type Provider struct {
	config Config
}

// Config is the main configuration structure.
type Config struct {
	// GommitConf is the configuration for gommitlint.
	GommitConf GommitLintConfig
}

// GommitLintConfig defines the configuration for commit validation.
type GommitLintConfig struct {
	// Subject is the configuration for commit subject validation.
	Subject SubjectConfig

	// ConventionalCommit is the configuration for conventional commit validation.
	ConventionalCommit ConventionalConfig

	// RequireSignature indicates whether commit signatures are required.
	RequireSignature bool
}

// SubjectConfig defines configuration for commit subject validation.
// This struct already used value semantics.
type SubjectConfig struct {
	// MaxLength is the maximum length of the commit subject.
	MaxLength int
}

// ConventionalConfig defines configuration for conventional commit validation.
// This struct already used value semantics.
type ConventionalConfig struct {
	// MaxDescriptionLength is the maximum length of the description in a conventional commit.
	MaxDescriptionLength int

	// Types are the allowed commit types.
	Types []string

	// Scopes are the allowed commit scopes.
	Scopes []string
}

// NewProvider creates a new configuration provider.
func NewProvider() (Provider, error) {
	// Create a default configuration with value semantics
	// Initialize all nested structures without pointers
	types := []string{
		"feat", "fix", "docs", "style", "refactor",
		"perf", "test", "build", "ci", "chore", "revert",
	}

	// Create a deep copy of the types slice to ensure immutability
	typesCopy := make([]string, len(types))
	copy(typesCopy, types)

	config := Config{
		GommitConf: GommitLintConfig{
			Subject: SubjectConfig{
				MaxLength: 100,
			},
			ConventionalCommit: ConventionalConfig{
				MaxDescriptionLength: 72,
				Types:                typesCopy,
				Scopes:               make([]string, 0), // Initialize to empty slice, not nil
			},
		},
	}

	return Provider{
		config: config,
	}, nil
}

// GetConfig returns the complete configuration.
// This implementation returns a copy of the configuration.
func (p Provider) GetConfig() Config {
	return p.config
}

// GetGommitConfig returns the GommitLintConfig.
// This implementation returns a copy of the GommitLintConfig.
func (p Provider) GetGommitConfig() GommitLintConfig {
	return p.config.GommitConf
}
