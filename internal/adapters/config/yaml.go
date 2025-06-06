// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// NewConfigWithDefaults creates configuration with application-specific defaults.
func NewConfigWithDefaults() config.Config {
	cfg := config.NewDefault()

	// Apply default disabled rules for this application
	cfg.Rules.Disabled = []string{
		"jirareference", // JIRAReference rule is disabled by default as it's organization-specific
		"commitbody",    // CommitBody rule is disabled by default as not all projects require detailed bodies
		"spell",         // Spell checking disabled by default (requires additional setup)
	}

	return cfg
}

// WithDefaults applies application-specific defaults to a configuration.
func WithDefaults(cfg config.Config) config.Config {
	defaults := NewConfigWithDefaults()

	// If no disabled rules are specified, use defaults
	if len(cfg.Rules.Disabled) == 0 {
		cfg.Rules.Disabled = defaults.Rules.Disabled
	}

	return cfg
}
