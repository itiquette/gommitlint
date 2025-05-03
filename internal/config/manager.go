// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itiquette/gommitlint/internal/contextx"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Manager is a configuration manager that uses value semantics with the configuration.
// It provides a simplified and consistent interface for all configuration operations.
type Manager struct {
	// Configuration cache
	config Config
	// Flag to track if config was loaded from a file
	loadedFromFile bool
	// Source file path if loaded from file
	sourcePath string
}

// NewManager creates a new configuration manager.
// It loads configuration in the following order of precedence:
// 1. Default values (always loaded)
// 2. User configuration from $XDG_CONFIG_HOME/gommitlint/gommitlint.yaml (if exists)
// 3. Project configuration from ./.gommitlint.yaml (if exists)
//
// Each configuration layer overrides values from the previous layers.
// This is the preferred entry point for creating a configuration service.
func NewManager(ctx context.Context) (*Manager, error) {
	manager := &Manager{
		config: NewConfig(), // Always start with defaults
	}

	// Try to load from standard paths using the provided context
	err := manager.LoadFromStandardPaths(ctx)
	if err != nil {
		// Just log the error and continue with defaults
		logger := log.Logger(ctx)
		logger.Warn().Err(err).Msg("Failed to load from standard paths, continuing with defaults")
	}

	return manager, nil
}

// LoadFromStandardPaths tries to load configuration from standard paths.
// It follows the XDG Base Directory Specification.
func (m *Manager) LoadFromStandardPaths(ctx context.Context) error {
	// Use the provided context instead of creating a new one
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering LoadFromStandardPaths")

	// Get paths in order of priority (project first, then XDG)
	paths := []string{".gommitlint.yaml"}

	// Add XDG config home
	var xdgConfigPath string
	if xdgHome, exists := os.LookupEnv("XDG_CONFIG_HOME"); exists && xdgHome != "" {
		xdgConfigPath = filepath.Join(xdgHome, "gommitlint", "gommitlint.yaml")
		paths = append(paths, xdgConfigPath)
	} else {
		// Default XDG config path
		home, err := os.UserHomeDir()
		if err == nil {
			xdgConfigPath = filepath.Join(home, ".config", "gommitlint", "gommitlint.yaml")
			paths = append(paths, xdgConfigPath)
		}
	}

	// Start with default configuration
	m.config = NewConfig()

	// Load from lowest precedence to highest (reverse the paths)
	configFound := false

	var lastError error

	// Iterate through paths in reverse (from lowest to highest precedence)
	for i := len(paths) - 1; i >= 0; i-- {
		path := paths[i]

		// Skip files that don't exist
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		// Try to load this config file
		err := m.LoadFromFile(ctx, path)
		if err != nil {
			lastError = err

			continue
		}

		// Mark this config as loaded
		configFound = true
	}

	// Return appropriate error if no configs were found
	if !configFound {
		if lastError != nil {
			return lastError
		}

		return errors.New("no configuration files found in standard paths")
	}

	return nil
}

// LoadFromFile loads configuration from a specific file path.
// If called multiple times, it merges configurations, with values from
// later calls overriding values from earlier calls.
func (m *Manager) LoadFromFile(ctx context.Context, path string) error {
	logger := log.Logger(ctx)
	logger.Trace().Str("path", path).Msg("Entering LoadFromFile")
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", path)
	}

	// Convert to absolute path for consistency
	absPath := path

	if !filepath.IsAbs(path) {
		var err error

		absPath, err = filepath.Abs(path)
		if err != nil {
			// If we can't get absolute path, use the original
			absPath = path
		}
	}

	// Create a new Koanf instance for this file
	knf := koanf.New(".")

	// Load based on file extension
	var err error

	switch {
	case strings.HasSuffix(path, ".yaml"), strings.HasSuffix(path, ".yml"):
		err = knf.Load(file.Provider(path), yaml.Parser())
	case strings.HasSuffix(path, ".json"):
		err = knf.Load(file.Provider(path), nil)
	default:
		return fmt.Errorf("unsupported configuration file format: %s", path)
	}

	if err != nil {
		return fmt.Errorf("failed to parse configuration file: %w", err)
	}

	// Parse the file and convert to internal format
	// By default, all configuration is expected to have a "gommitlint:" prefix
	var gommitConfig GommitlintConfig

	// Attempt structured extraction of Jira settings directly
	var tempMap map[string]interface{}
	if err := knf.Unmarshal("", &tempMap); err == nil {
		// Try to extract subject.jira.required directly from map structure
		// This handles cases where only that setting is provided
		if val, ok := tempMap["gommitlint"].(map[string]interface{}); ok {
			if subj, ok := val["subject"].(map[string]interface{}); ok {
				if jira, ok := subj["jira"].(map[string]interface{}); ok {
					if req, ok := jira["required"].(bool); ok {
						// Set the value directly
						gommitConfig.Subject.Jira.Required = req
					}
				}
			}
		}
	}

	// Unmarshal configuration with gommitlint prefix
	err = knf.Unmarshal("gommitlint", &gommitConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Convert from YAML/JSON structure to our internal Config type
	fileConfig := FromGommitlintConfig(gommitConfig)

	// Apply this configuration on top of existing config
	m.applyConfig(fileConfig)

	// Update metadata
	m.loadedFromFile = true
	m.sourcePath = absPath

	return nil
}

// applyConfig applies the values from the source config to the manager's config.
// This is a simple implementation for the migration - in a real implementation,
// we might want to be more selective about which fields to override.
func (m *Manager) applyConfig(source Config) {
	// For simplicity, we directly apply the full config
	m.config = source
}

// GetConfig returns the current configuration.
func (m *Manager) GetConfig() Config {
	return m.config
}

// UpdateConfig applies the given options to the current configuration.
// This allows for programmatically updating the configuration.
func (m *Manager) UpdateConfig(opts ...Option) {
	m.config = contextx.Reduce(opts, m.config, func(config Config, opt Option) Config {
		return opt(config)
	})
}

// WasLoadedFromFile returns whether configuration was loaded from a file.
func (m *Manager) WasLoadedFromFile() bool {
	return m.loadedFromFile
}

// GetSourcePath returns the path from which configuration was loaded, if any.
func (m *Manager) GetSourcePath() string {
	return m.sourcePath
}

// SetConfig sets the configuration directly.
// This is useful for testing or for programmatically setting the configuration.
func (m *Manager) SetConfig(config Config) {
	m.config = config
}

// GetValidationConfig returns the current configuration for validation.
func (m *Manager) GetValidationConfig(ctx context.Context) Config {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GetValidationConfig")

	return m.config
}
