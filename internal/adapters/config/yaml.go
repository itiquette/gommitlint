// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain/config"
)

// Service provides configuration management functionality.
// It handles loading, storing, and providing access to configuration.
type Service struct {
	config config.Config
}

// NewService creates a new configuration service with default configuration.
func NewService() (Service, error) {
	// Use the domain's default configuration
	cfg := config.NewDefault()

	// Apply default disabled rules
	cfg.Rules.Disabled = []string{
		"jirareference", // JIRAReference rule is disabled by default as it's organization-specific
		"commitbody",    // CommitBody rule is disabled by default as not all projects require detailed bodies
	}

	return Service{
		config: cfg,
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
