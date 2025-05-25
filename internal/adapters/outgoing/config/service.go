// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/config/types"
)

// Service provides configuration management functionality.
// It handles loading, storing, and providing access to configuration.
type Service struct {
	config types.Config
}

// NewService creates a new configuration service with default configuration.
func NewService() (Service, error) {
	// Start with default config
	config := NewDefaultConfig()

	// Apply default disabled rules
	// In hexagonal architecture, adapters should not depend on domain.
	// The default disabled rules should be part of the configuration layer.
	defaultDisabledRules := []string{
		"jirareference", // JIRAReference rule is disabled by default as it's organization-specific
		"commitbody",    // CommitBody rule is disabled by default as not all projects require detailed bodies
	}

	// Merge with any existing disabled rules from the config
	for _, rule := range config.Rules.Disabled {
		found := false

		for _, defaultRule := range defaultDisabledRules {
			if rule == defaultRule {
				found = true

				break
			}
		}

		if !found {
			defaultDisabledRules = append(defaultDisabledRules, rule)
		}
	}

	config.Rules.Disabled = defaultDisabledRules

	return Service{
		config: config,
	}, nil
}

// NewServiceWithConfig creates a new service with the provided configuration.
func NewServiceWithConfig(config types.Config) Service {
	return Service{
		config: config,
	}
}

// GetConfig returns the current configuration.
func (s Service) GetConfig() types.Config {
	return s.config
}

// UpdateConfig applies a transformation to the current configuration.
// It returns a new Service instance to maintain immutability.
func (s Service) UpdateConfig(transform func(types.Config) types.Config) Service {
	newConfig := transform(s.config)

	return Service{
		config: newConfig,
	}
}

// Load loads configuration from default paths using the loader.
// Returns a new Service with the loaded configuration.
func (s Service) Load() (Service, error) {
	loader := NewLoader()

	config, err := loader.Load()
	if err != nil {
		return s, fmt.Errorf("failed to load config: %w", err)
	}

	return Service{config: config}, nil
}

// LoadFromPath loads configuration from a specific path.
// Returns a new Service with the loaded configuration.
func (s Service) LoadFromPath(path string) (Service, error) {
	loader := NewLoader()

	config, err := loader.LoadFromPath(path)
	if err != nil {
		return s, fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	return Service{config: config}, nil
}

// GetAdapter returns a config adapter for the current configuration.
func (s Service) GetAdapter() *Adapter {
	return NewAdapter(s.config)
}

// NewDefaultConfig creates a default configuration.
func NewDefaultConfig() types.Config {
	// Default disabled rules - defined here instead of importing from domain
	// to maintain proper hexagonal architecture boundaries
	disabledRules := []string{
		"jirareference", // JIRAReference rule is disabled by default as it's organization-specific
		"commitbody",    // CommitBody rule is disabled by default as not all projects require detailed bodies
	}

	config := types.Config{
		Message: types.MessageConfig{
			Subject: types.SubjectConfig{
				Case:              "sentence",
				MaxLength:         72,
				RequireImperative: false,
				ForbidEndings:     []string{"."},
			},
			Body: types.BodyConfig{
				MinLength:        10,
				MinLines:         3,
				AllowSignoffOnly: false,
				RequireSignoff:   false,
			},
		},
		Conventional: types.ConventionalConfig{
			RequireScope:         false,
			Types:                []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"},
			AllowBreaking:        true,
			MaxDescriptionLength: 72,
		},
		Rules: types.RulesConfig{
			Enabled: []string{
				"SubjectLength",
				"CommitBody",
				"Conventional",
				"Imperative",
				"SubjectCase",
				"SubjectSuffix",
			},
			Disabled: disabledRules,
		},
		Signing: types.SigningConfig{
			RequireSignature:      false,
			AllowMultipleSignoffs: false,
		},
		Repo: types.RepoConfig{
			Path:            ".",
			MaxCommitsAhead: 10,
			Branch:          "main",
			IgnoreMerges:    true,
		},
		Output: "text",
		Spell: types.SpellConfig{
			Language:    "en_US",
			IgnoreWords: []string{},
		},
		Jira: types.JiraConfig{
			Projects: []string{},
		},
	}

	// The disabled rules have already been set above
	// No need to apply them again

	return config
}
