// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package loader

import (
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal/config"
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

// Load loads configuration from default paths.
func (l Loader) Load() (config.Config, error) {
	// Try to load from each standard path
	for _, path := range l.configPaths {
		if _, err := os.Stat(path); err == nil {
			return l.LoadFromPath(path)
		}
	}

	// No config file found, return defaults
	return NewDefaultConfig(), nil
}

// LoadFromPath loads configuration from a specific path.
func (l Loader) LoadFromPath(configPath string) (config.Config, error) {
	// Create koanf instance
	koanfConfig := koanf.New(".")

	// Load YAML configuration
	if err := koanfConfig.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return config.Config{}, fmt.Errorf("error loading config file %s: %w", configPath, err)
	}

	// Start with defaults
	cfg := NewDefaultConfig()

	// Store original default disabled rules to preserve
	defaultDisabledRules := cfg.Rules.Disabled

	// Unmarshal the root configuration
	// Use yaml tag instead of json to properly load yaml files
	if err := koanfConfig.UnmarshalWithConf("gommitlint", &cfg, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return config.Config{}, fmt.Errorf("error unmarshaling config for gommitlint namespace: %w", err)
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
func (l Loader) applyRulePriority(config config.Config) config.Config {
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
