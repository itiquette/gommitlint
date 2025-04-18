// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration loading and access.
package config

import (
// No imports needed for now.
)

// Provider provides access to configuration.
type Provider struct {
	config *Config
}

// Config is the main configuration structure.
type Config struct {
	// GommitConf is the configuration for gommitlint.
	GommitConf *GommitLintConfig
}

// GommitLintConfig defines the configuration for commit validation.
type GommitLintConfig struct {
	// Subject is the configuration for commit subject validation.
	Subject *SubjectConfig

	// ConventionalCommit is the configuration for conventional commit validation.
	ConventionalCommit *ConventionalConfig

	// RequireSignature indicates whether commit signatures are required.
	RequireSignature bool
}

// SubjectConfig defines configuration for commit subject validation.
type SubjectConfig struct {
	// MaxLength is the maximum length of the commit subject.
	MaxLength int
}

// ConventionalConfig defines configuration for conventional commit validation.
type ConventionalConfig struct {
	// MaxDescriptionLength is the maximum length of the description in a conventional commit.
	MaxDescriptionLength int

	// Types are the allowed commit types.
	Types []string

	// Scopes are the allowed commit scopes.
	Scopes []string
}

// NewProvider creates a new configuration provider.
func NewProvider() (*Provider, error) {
	// Create a default configuration
	config := &Config{
		GommitConf: &GommitLintConfig{
			Subject: &SubjectConfig{
				MaxLength: 100,
			},
			ConventionalCommit: &ConventionalConfig{
				MaxDescriptionLength: 72,
				Types: []string{
					"feat", "fix", "docs", "style", "refactor",
					"perf", "test", "build", "ci", "chore", "revert",
				},
				Scopes: []string{},
			},
		},
	}

	return &Provider{
		config: config,
	}, nil
}

// GetConfig returns the complete configuration.
func (p *Provider) GetConfig() *Config {
	return p.config
}

// GetGommitConfig returns the GommitLintConfig.
func (p *Provider) GetGommitConfig() *GommitLintConfig {
	if p.config == nil || p.config.GommitConf == nil {
		return nil
	}

	return p.config.GommitConf
}
