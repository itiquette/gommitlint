// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"fmt"
	"os"

	domainConfig "github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Loader handles configuration file loading operations.
type Loader struct {
	// Standard configuration paths to search
	configPaths []string
}

// NewLoader creates a new configuration loader.
func NewLoader() *Loader {
	return &Loader{
		configPaths: []string{
			".gommitlint.yaml",
			".gommitlint.yml",
			".config/gommitlint/config.yaml",
			".config/gommitlint/config.yml",
		},
	}
}

// LoadFromFile loads configuration from default paths.
func (l Loader) LoadFromFile() (domainConfig.Config, error) {
	// Try to load from each standard path
	for _, path := range l.configPaths {
		if _, err := os.Stat(path); err == nil {
			return l.LoadFromPath(path)
		}
	}

	// No config file found, return defaults
	return domainConfig.NewDefault(), nil
}

// LoadFromPath loads configuration from a specific path.
func (l Loader) LoadFromPath(configPath string) (domainConfig.Config, error) {
	// Create koanf instance
	koanfConfig := koanf.New(".")

	// Load YAML configuration
	if err := koanfConfig.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return domainConfig.Config{}, fmt.Errorf("error loading config file %s: %w", configPath, err)
	}

	// Start with defaults
	cfg := domainConfig.NewDefault()

	// Store original default disabled rules to preserve
	defaultDisabledRules := cfg.Rules.Disabled

	// Unmarshal the root configuration
	// Use yaml tag instead of json to properly load yaml files
	if err := koanfConfig.UnmarshalWithConf("gommitlint", &cfg, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return domainConfig.Config{}, fmt.Errorf("error unmarshaling config for gommitlint namespace: %w", err)
	}

	// Merge default disabled rules with YAML disabled rules
	if koanfConfig.Exists("gommitlint.rules.disabled") {
		// YAML has disabled, apply priority logic
		cfg = l.applyRulePriority(cfg)
	} else {
		// No disabled in YAML, apply defaults and merge
		cfg.Rules.Disabled = l.mergeDefaultDisabledRules(cfg.Rules.Enabled, defaultDisabledRules)
	}

	return cfg, nil
}

// applyRulePriority ensures that rules in disabled_rules are removed from enabled_rules.
// If a rule appears in both lists, disabled takes precedence.
// Returns a new config with the updated rules.
func (l Loader) applyRulePriority(config domainConfig.Config) domainConfig.Config {
	// Create a map of disabled rules for efficient lookup
	disabledMap := make(map[string]bool)
	for _, rule := range config.Rules.Disabled {
		disabledMap[rule] = true
	}

	// Filter enabled rules to remove any that are also disabled
	var filteredEnabled []string

	for _, rule := range config.Rules.Enabled {
		if !disabledMap[rule] {
			filteredEnabled = append(filteredEnabled, rule)
		}
	}

	// Create a new config with the filtered rules
	result := config
	result.Rules.Enabled = filteredEnabled

	return result
}

// mergeDefaultDisabledRules merges default disabled rules with the current configuration.
// When no disabled are specified in YAML, we keep default disabled rules
// even if they're explicitly enabled (both lists can contain the rule).
func (l Loader) mergeDefaultDisabledRules(_ []string, defaultDisabledRules []string) []string {
	// Just return the default disabled rules - they coexist with enabled rules
	return defaultDisabledRules
}

// Load provides a simple entry point for loading configuration.
// It combines file loading and environment variable overrides.
func Load() (domainConfig.Config, error) {
	loader := NewLoader()

	// Try to load from file
	cfg, err := loader.LoadFromFile()
	if err != nil {
		// If no config file found, use defaults
		cfg = domainConfig.NewDefault()
	}

	// Apply environment variable overrides
	cfg = LoadFromEnv(cfg)

	// Validate the final configuration
	if validationErrors := cfg.Validate(); len(validationErrors) > 0 {
		return domainConfig.Config{}, fmt.Errorf("configuration validation failed: %v", validationErrors)
	}

	return cfg, nil
}
