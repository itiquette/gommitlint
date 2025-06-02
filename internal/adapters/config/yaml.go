// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// Service provides configuration management functionality.
// It handles loading, storing, and providing access to configuration.
type Service struct {
	config config.Config
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

	// Merge with any existing disabled rules from the config using functional approach
	uniqueRules := make(map[string]bool)
	for _, rule := range append(defaultDisabledRules, config.Rules.Disabled...) {
		uniqueRules[rule] = true
	}

	// Convert map keys back to slice using functional approach
	config.Rules.Disabled = domain.MapKeys(uniqueRules)

	return Service{
		config: config,
	}, nil
}

// NewServiceWithConfig creates a new service with the provided configuration.
func NewServiceWithConfig(config config.Config) Service {
	return Service{
		config: config,
	}
}

// GetConfig returns the current configuration.
func (s Service) GetConfig() config.Config {
	return s.config
}

// UpdateConfig applies a transformation to the current configuration.
// It returns a new Service instance to maintain immutability.
func (s Service) UpdateConfig(transform func(config.Config) config.Config) Service {
	newConfig := transform(s.config)

	return Service{
		config: newConfig,
	}
}

// Load loads configuration from default paths using the loader.
// Returns a new Service with the loaded configuration.
func (s Service) Load() (Service, error) {
	loader := NewLoader()

	config, err := loader.LoadFromFile()
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

// NewDefaultConfig creates a default configuration.
func NewDefaultConfig() config.Config {
	// Default disabled rules - defined here instead of importing from domain
	// to maintain proper hexagonal architecture boundaries
	disabledRules := []string{
		"jirareference", // JIRAReference rule is disabled by default as it's organization-specific
		"commitbody",    // CommitBody rule is disabled by default as not all projects require detailed bodies
	}

	config := config.Config{
		Message: config.MessageConfig{
			Subject: config.SubjectConfig{
				Case:              "sentence",
				MaxLength:         72,
				RequireImperative: false,
				ForbidEndings:     []string{"."},
			},
			Body: config.BodyConfig{
				MinLength:        10,
				MinLines:         3,
				AllowSignoffOnly: false,
				RequireSignoff:   false,
			},
		},
		Conventional: config.ConventionalConfig{
			RequireScope:         false,
			Types:                []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"},
			AllowBreaking:        true,
			MaxDescriptionLength: 72,
		},
		Rules: config.RulesConfig{
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
		Signing: config.SigningConfig{
			RequireSignature:    false,
			RequireVerification: false,
			RequireMultiSignoff: false,
			KeyDirectory:        "",
			AllowedSigners:      []string{},
		},
		Repo: config.RepoConfig{
			MaxCommitsAhead:   10,
			ReferenceBranch:   "main",
			AllowMergeCommits: true,
		},
		Output: "text",
		Spell: config.SpellConfig{
			Locale:      "en_US",
			IgnoreWords: []string{},
		},
		Jira: config.JiraConfig{
			ProjectPrefixes:      []string{},
			RequireInBody:        false,
			RequireInSubject:     false,
			IgnoreTicketPatterns: []string{},
		},
	}

	// The disabled rules have already been set above
	// No need to apply them again

	return config
}
