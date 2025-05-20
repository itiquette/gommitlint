// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
)

// Service provides configuration management functionality.
// It handles loading, storing, and providing access to configuration.
type Service struct {
	config types.Config
}

// NewService creates a new configuration service with default configuration.
func NewService() (*Service, error) {
	// Start with default config
	config := NewDefaultConfig()

	// Apply default disabled rules
	defaultDisabledRules := append([]string{}, config.Rules.Disabled...)

	// Add rules that should be disabled by default from the central list
	domainDefaultDisabled := domain.GetDefaultDisabledRules()
	for ruleName := range domainDefaultDisabled {
		// Check if the rule is explicitly enabled in the default config
		isExplicitlyEnabled := false

		for _, enabledRule := range config.Rules.Enabled {
			if enabledRule == ruleName {
				isExplicitlyEnabled = true

				break
			}
		}

		// Only add to disabled if not explicitly enabled
		if !isExplicitlyEnabled {
			defaultDisabledRules = append(defaultDisabledRules, ruleName)
		}
	}

	config.Rules.Disabled = defaultDisabledRules

	return &Service{
		config: config,
	}, nil
}

// NewServiceWithConfig creates a new service with the provided configuration.
func NewServiceWithConfig(config types.Config) *Service {
	return &Service{
		config: config,
	}
}

// GetConfig returns the current configuration.
func (s *Service) GetConfig() types.Config {
	return s.config
}

// UpdateConfig applies a transformation to the current configuration.
// It returns a new Service instance to maintain immutability.
func (s *Service) UpdateConfig(transform func(types.Config) types.Config) *Service {
	newConfig := transform(s.config)

	return &Service{
		config: newConfig,
	}
}

// Load loads configuration from default paths using the loader.
func (s *Service) Load() error {
	loader := NewLoader()

	config, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	s.config = config

	return nil
}

// LoadFromPath loads configuration from a specific path.
func (s *Service) LoadFromPath(path string) error {
	loader := NewLoader()

	config, err := loader.LoadFromPath(path)
	if err != nil {
		return fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	s.config = config

	return nil
}

// GetAdapter returns a config adapter for the current configuration.
func (s *Service) GetAdapter() *Adapter {
	return NewAdapter(s.config)
}

// NewDefaultConfig creates a default configuration.
func NewDefaultConfig() types.Config {
	// Get default disabled rules from domain
	defaultDisabledRules := domain.GetDefaultDisabledRules()
	disabledRules := make([]string, 0, len(defaultDisabledRules))

	for rule := range defaultDisabledRules {
		disabledRules = append(disabledRules, rule)
	}

	config := types.Config{
		Subject: types.SubjectConfig{
			Case:               "sentence",
			MaxLength:          72,
			Imperative:         false,
			DisallowedSuffixes: []string{"."},
		},
		Body: types.BodyConfig{
			MinLength:        10,
			MinLines:         3,
			AllowSignOffOnly: false,
			RequireSignOff:   false,
		},
		Conventional: types.ConventionalConfig{
			RequireScope:         false,
			Types:                []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"},
			AllowBreakingChanges: true,
			MaxDescriptionLength: 72,
		},
		Rules: types.RulesConfig{
			Enabled: []string{
				"SubjectLength",
				"Body",
				"Conventional",
				"Imperative",
				"SubjectCase",
				"SubjectSuffix",
			},
			Disabled: disabledRules,
		},
		Security: types.SecurityConfig{
			GPGRequired:      false,
			MultipleSignoffs: false,
		},
		Repository: types.RepositoryConfig{
			Path:               ".",
			MaxCommitsAhead:    10,
			ReferenceBranch:    "main",
			IgnoreMergeCommits: true,
		},
		Output: types.OutputConfig{
			Format: "text",
		},
		SpellCheck: types.SpellCheckConfig{
			Language:         "en_US",
			CustomDictionary: []string{},
		},
		Jira: types.JiraConfig{
			Projects: []string{},
		},
	}

	// Apply default disabled rules from domain
	domainDisabledRules := []string{}

	domainDefaultDisabled := domain.GetDefaultDisabledRules()
	for ruleName := range domainDefaultDisabled {
		// Check if the rule is explicitly enabled
		isEnabled := false

		for _, enabledRule := range config.Rules.Enabled {
			if enabledRule == ruleName {
				isEnabled = true

				break
			}
		}

		// Only add to disabled if not explicitly enabled
		if !isEnabled {
			domainDisabledRules = append(domainDisabledRules, ruleName)
		}
	}

	config.Rules.Disabled = domainDisabledRules

	return config
}
