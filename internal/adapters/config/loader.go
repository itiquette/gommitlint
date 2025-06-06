// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"fmt"
	"os"

	configTypes "github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// defaultConfigPaths defines the standard configuration file search paths.
var defaultConfigPaths = []string{
	".gommitlint.yaml",
	".gommitlint.yml",
	".config/gommitlint/config.yaml",
	".config/gommitlint/config.yml",
}

// LoadConfig loads configuration from multiple sources with later configs taking precedence.
func LoadConfig() (configTypes.Config, error) {
	return MergeConfigs(
		LoadDefaultConfig(),
		LoadFileConfig(findFirstExistingConfigFile()),
		LoadEnvConfig(),
	)
}

// LoadConfigFromPath loads configuration from a specific path using functional composition.
func LoadConfigFromPath(configPath string) (configTypes.Config, error) {
	return MergeConfigs(
		LoadDefaultConfig(),
		LoadFileConfig(configPath),
		LoadEnvConfig(),
	)
}

// LoadDefaultConfig returns the default configuration with application-specific defaults.
func LoadDefaultConfig() configTypes.Config {
	return NewConfigWithDefaults()
}

// LoadFileConfig loads configuration from a file.
// Returns empty config if file doesn't exist or can't be loaded.
func LoadFileConfig(configPath string) configTypes.Config {
	if configPath == "" {
		return configTypes.Config{} // Empty config
	}

	// Check if file exists
	if _, err := os.Stat(configPath); err != nil {
		return configTypes.Config{} // Empty config
	}

	// Create koanf instance
	koanfConfig := koanf.New(".")

	// Load YAML configuration
	if err := koanfConfig.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return configTypes.Config{} // Empty config on error
	}

	// Parse into config struct
	var cfg configTypes.Config
	if err := koanfConfig.UnmarshalWithConf("gommitlint", &cfg, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return configTypes.Config{} // Empty config on error
	}

	// Apply rule priority logic
	return applyRulePriority(cfg)
}

// LoadEnvConfig loads configuration from environment variables.
func LoadEnvConfig() configTypes.Config {
	return LoadFromEnv(configTypes.Config{})
}

// MergeConfigs merges multiple configurations with later configs taking precedence.
func MergeConfigs(configs ...configTypes.Config) (configTypes.Config, error) {
	if len(configs) == 0 {
		return configTypes.NewDefault(), nil
	}

	// Start with first config
	result := configs[0]

	// Merge each subsequent config
	for _, cfg := range configs[1:] {
		result = mergeConfig(result, cfg)
	}

	// Validate final configuration
	if validationErrors := result.Validate(); len(validationErrors) > 0 {
		return configTypes.Config{}, fmt.Errorf("configuration validation failed: %v", validationErrors)
	}

	return result, nil
}

// mergeConfig merges two configurations with the second taking precedence.
func mergeConfig(base, overlay configTypes.Config) configTypes.Config {
	result := base

	// Merge non-zero values from overlay
	if overlay.Output != "" {
		result.Output = overlay.Output
	}

	// Merge message config
	if overlay.Message.Subject.MaxLength != 0 {
		result.Message.Subject.MaxLength = overlay.Message.Subject.MaxLength
	}

	if overlay.Message.Subject.Case != "" {
		result.Message.Subject.Case = overlay.Message.Subject.Case
	}

	if len(overlay.Message.Subject.ForbidEndings) > 0 {
		result.Message.Subject.ForbidEndings = overlay.Message.Subject.ForbidEndings
	}

	// Merge conventional config
	if len(overlay.Conventional.Types) > 0 {
		result.Conventional.Types = overlay.Conventional.Types
	}

	if len(overlay.Conventional.Scopes) > 0 {
		result.Conventional.Scopes = overlay.Conventional.Scopes
	}

	// Merge repo config
	if overlay.Repo.ReferenceBranch != "" {
		result.Repo.ReferenceBranch = overlay.Repo.ReferenceBranch
	}

	if overlay.Repo.MaxCommitsAhead != 0 {
		result.Repo.MaxCommitsAhead = overlay.Repo.MaxCommitsAhead
	}

	// Merge rules config - always override if present
	if len(overlay.Rules.Enabled) > 0 {
		result.Rules.Enabled = overlay.Rules.Enabled
	}

	if len(overlay.Rules.Disabled) > 0 {
		result.Rules.Disabled = overlay.Rules.Disabled
	}

	// Merge Jira config
	if len(overlay.Jira.ProjectPrefixes) > 0 {
		result.Jira.ProjectPrefixes = overlay.Jira.ProjectPrefixes
	}

	if len(overlay.Jira.IgnoreTicketPatterns) > 0 {
		result.Jira.IgnoreTicketPatterns = overlay.Jira.IgnoreTicketPatterns
	}

	// Merge Spell config
	if len(overlay.Spell.IgnoreWords) > 0 {
		result.Spell.IgnoreWords = overlay.Spell.IgnoreWords
	}

	if overlay.Spell.Locale != "" {
		result.Spell.Locale = overlay.Spell.Locale
	}

	// Merge Signing config
	if overlay.Signing.KeyDirectory != "" {
		result.Signing.KeyDirectory = overlay.Signing.KeyDirectory
	}

	if len(overlay.Signing.AllowedSigners) > 0 {
		result.Signing.AllowedSigners = overlay.Signing.AllowedSigners
	}

	return result
}

// findFirstExistingConfigFile finds the first existing config file in default paths.
func findFirstExistingConfigFile() string {
	for _, path := range defaultConfigPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// applyRulePriority ensures that rules in disabled_rules are removed from enabled_rules.
// If a rule appears in both lists, disabled takes precedence.
func applyRulePriority(config configTypes.Config) configTypes.Config {
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
