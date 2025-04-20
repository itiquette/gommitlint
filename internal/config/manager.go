// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration management for gommitlint.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itiquette/gommitlint/internal/core/validation"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Manager handles loading, validating, and providing configuration.
type Manager struct {
	// Configuration cache
	config *AppConf
	// Flag to track if config was loaded from a file
	loadedFromFile bool
	// Source file path if loaded from file
	sourcePath string
}

// New creates a new configuration manager.
// It loads configuration in the following order of precedence:
// 1. Default values (always loaded)
// 2. User configuration from $XDG_CONFIG_HOME/gommitlint/gommitlint.yaml (if exists)
// 3. Project configuration from ./.gommitlint.yaml (if exists)
//
// Each configuration layer overrides values from the previous layers.
func New() (*Manager, error) {
	manager := &Manager{
		config: DefaultConfiguration(), // Always start with defaults
	}

	// Try to load from standard paths
	err := manager.LoadFromStandardPaths()
	if err != nil {
		// Just log the error and continue with defaults
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	return manager, nil
}

// LoadFromStandardPaths tries to load configuration from standard paths.
// It follows the XDG Base Directory Specification.
// The function merges configurations in order of precedence, with values from
// higher precedence files overriding values from lower precedence files.
func (m *Manager) LoadFromStandardPaths() error {
	// Get paths in order of priority (project first, then XDG)
	paths := getDefaultConfigPaths()

	// Start with default configuration
	m.config = DefaultConfiguration()

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
		knf := koanf.New(".")

		// Load based on file extension
		var err error

		switch {
		case strings.HasSuffix(path, ".yaml"), strings.HasSuffix(path, ".yml"):
			err = knf.Load(file.Provider(path), yaml.Parser())
		case strings.HasSuffix(path, ".json"):
			err = knf.Load(file.Provider(path), nil)
		default:
			continue // Skip unsupported format
		}

		if err != nil {
			lastError = fmt.Errorf("failed to parse %s: %w", path, err)

			continue
		}

		// Unmarshal into our config
		if err := knf.Unmarshal("", m.config); err != nil {
			lastError = fmt.Errorf("failed to unmarshal %s: %w", path, err)

			continue
		}

		// Mark this config as loaded
		configFound = true

		// Update source path - this will be overwritten by higher precedence
		// configs if they are also loaded
		m.sourcePath = path
		m.loadedFromFile = true
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

// This is used internally by LoadFromStandardPaths.
func (m *Manager) loadFileWithoutUpdatingSourcePath(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", path)
	}

	// Create a new Koanf instance for this file
	knf := koanf.New(".")

	// Load configuration based on file extension
	var err error

	switch {
	case strings.HasSuffix(path, ".yaml"), strings.HasSuffix(path, ".yml"):
		err = knf.Load(file.Provider(path), yaml.Parser())
	case strings.HasSuffix(path, ".json"):
		err = knf.Load(file.Provider(path), nil)
	default:
		return fmt.Errorf("unsupported configuration file format: %s", path)
	}
	// Check if the main 'gommitlint' key exists in the loaded file's content
	if !knf.Exists("gommitlint") {
		// If the file was successfully parsed but is effectively empty regarding gommitlint config
		// (e.g., contains only comments or unrelated keys), return the specific error.
		// Check if the raw loaded data is empty, perhaps k.Raw() map is empty or nil check?
		// This depends on how Koanf represents an empty but valid YAML/JSON file after Load.
		// Let's assume for now an empty file or file without 'gommitlint' key fails this check.
		return errors.New("configuration is empty or missing gommitlint section")
	}

	if err != nil {
		return fmt.Errorf("failed to parse configuration file: %w", err)
	}

	// If we don't have a config yet, start with defaults
	if m.config == nil {
		m.config = DefaultConfiguration()
	}

	// Unmarshal into existing configuration (merging values)
	if err := knf.Unmarshal("", m.config); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Validate configuration
	if errs := validateConfig(m.config); len(errs) > 0 {
		errMsgs := make([]string, len(errs))
		for i, err := range errs {
			errMsgs[i] = err.Error()
		}

		return fmt.Errorf("invalid configuration: %s", strings.Join(errMsgs, "; "))
	}

	// Intentionally don't update sourcePath or loadedFromFile here

	return nil
}

// LoadFromFile loads configuration from a specific file path.
// If called multiple times, it merges configurations, with values from
// later calls overriding values from earlier calls.
func (m *Manager) LoadFromFile(path string) error {
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

	// Load the file using the helper method first
	err := m.loadFileWithoutUpdatingSourcePath(absPath)
	if err != nil {
		return err
	}

	// Update metadata
	m.loadedFromFile = true
	m.sourcePath = absPath

	return nil
}

// GetConfig returns the current configuration.
// If no configuration has been loaded, it returns the default configuration.
func (m *Manager) GetConfig() *AppConf {
	if m.config == nil {
		m.config = DefaultConfiguration()
	}

	return m.config
}

// GetRuleConfig converts the application configuration to validation rule configuration.
func (m *Manager) GetRuleConfig() *validation.RuleConfiguration {
	appConf := m.GetConfig()
	if appConf == nil || appConf.GommitConf == nil {
		// This shouldn't happen due to GetConfig() always returning at least defaults
		return validation.DefaultConfiguration()
	}

	conf := appConf.GommitConf

	// Create rule configuration from app configuration
	ruleConfig := &validation.RuleConfiguration{
		// Subject configuration
		MaxSubjectLength:       conf.Subject.MaxLength,
		SubjectCaseChoice:      conf.Subject.Case,
		SubjectInvalidSuffixes: conf.Subject.InvalidSuffixes,

		// Conventional commit configuration
		ConventionalTypes:    conf.ConventionalCommit.Types,
		MaxDescLength:        conf.ConventionalCommit.MaxDescriptionLength,
		IsConventionalCommit: conf.ConventionalCommit.Required,

		// Body configuration
		RequireBody: conf.Body.Required,

		// Security configuration
		RequireSignOff: *conf.SignOffRequired,

		// Reference configuration
		Reference: conf.Reference,

		// Disabled rules
		DisabledRules: []string{},
	}

	// Add JIRA configuration if present
	if conf.Subject.Jira != nil {
		ruleConfig.JiraRequired = conf.Subject.Jira.Required
		ruleConfig.JiraPattern = conf.Subject.Jira.Pattern

		// If Jira is not required, add it to disabled rules
		if !conf.Subject.Jira.Required {
			ruleConfig.DisabledRules = append(ruleConfig.DisabledRules, "JiraReference")
		}
	} else {
		// If no Jira config, disable Jira rule
		ruleConfig.DisabledRules = append(ruleConfig.DisabledRules, "JiraReference")
	}

	// Add signature configuration if present
	if conf.Signature != nil {
		ruleConfig.RequireSignature = conf.Signature.Required
	}

	return ruleConfig
}

// WasLoadedFromFile returns whether configuration was loaded from a file.
func (m *Manager) WasLoadedFromFile() bool {
	return m.loadedFromFile
}

// GetSourcePath returns the path from which configuration was loaded, if any.
func (m *Manager) GetSourcePath() string {
	return m.sourcePath
}
