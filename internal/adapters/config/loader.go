// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configTypes "github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// getConfigSearchPaths returns the search paths for configuration files.
// Supports YAML and TOML formats with priority: local files first, then XDG config.
func getConfigSearchPaths() []string {
	return getConfigSearchPathsForRepo("")
}

// getConfigSearchPathsForRepo returns config search paths for a specific repository directory.
// If repoPath is empty, searches in current directory. Otherwise searches in repository directory.
func getConfigSearchPathsForRepo(repoPath string) []string {
	var paths []string

	// Determine base directory for config files with security validation
	baseDir := "."

	if repoPath != "" {
		// Validate the repo path for security (but allow non-git directories for config search)
		cleanPath := filepath.Clean(repoPath)
		absPath, err := filepath.Abs(cleanPath)

		if err != nil || strings.Contains(cleanPath, "..") {
			// Fall back to current directory if path is suspicious
			baseDir = "."
		} else {
			baseDir = absPath
		}
	}

	// Add local config files (in repository or current directory)
	paths = []string{
		filepath.Join(baseDir, ".gommitlint.yaml"),
		filepath.Join(baseDir, ".gommitlint.yml"),
		filepath.Join(baseDir, ".gommitlint.toml"),
	}

	// Add XDG config paths if XDG_CONFIG_HOME is set and directory exists
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		// Validate XDG_CONFIG_HOME path for security
		cleanXDG := filepath.Clean(xdgConfigHome)
		if filepath.IsAbs(cleanXDG) && !strings.Contains(cleanXDG, "..") {
			gommitlintDir := filepath.Join(cleanXDG, "gommitlint")
			if _, err := os.Stat(gommitlintDir); err == nil {
				paths = append(paths,
					filepath.Join(gommitlintDir, "config.yaml"),
					filepath.Join(gommitlintDir, "config.yml"),
					filepath.Join(gommitlintDir, "config.toml"),
				)
			}
		}
	}

	return paths
}

// LoadConfig loads configuration from multiple sources with later configs taking precedence.
func LoadConfig() (configTypes.Config, error) {
	return LoadConfigWithRepoPath("")
}

// LoadConfigWithRepoPath loads configuration with repository path for config file discovery.
// If repoPath is provided, searches for config files in that directory first.
func LoadConfigWithRepoPath(repoPath string) (configTypes.Config, error) {
	return MergeConfigs(
		LoadDefaultConfig(),
		LoadFileConfig(findFirstExistingConfigFileInRepo(repoPath)),
	)
}

// LoadConfigFromPath loads configuration from a specific path using functional composition.
func LoadConfigFromPath(configPath string) (configTypes.Config, error) {
	return MergeConfigs(
		LoadDefaultConfig(),
		LoadFileConfig(configPath),
	)
}

// LoadDefaultConfig returns the default configuration with application-specific defaults.
func LoadDefaultConfig() configTypes.Config {
	return NewConfigWithDefaults()
}

// LoadFileConfig loads configuration from a file.
// Supports both YAML and TOML formats based on file extension.
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

	// Determine parser based on file extension
	var parser koanf.Parser

	var tagName string

	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".toml":
		parser = toml.Parser()
		tagName = "toml"
	case ".yaml", ".yml":
		parser = yaml.Parser()
		tagName = "yaml"
	default:
		// Default to YAML for unknown extensions
		parser = yaml.Parser()
		tagName = "yaml"
	}

	// Load configuration using appropriate parser
	if err := koanfConfig.Load(file.Provider(configPath), parser); err != nil {
		return configTypes.Config{} // Empty config on error
	}

	// Parse into config struct
	var cfg configTypes.Config
	if err := koanfConfig.UnmarshalWithConf("gommitlint", &cfg, koanf.UnmarshalConf{Tag: tagName}); err != nil {
		return configTypes.Config{} // Empty config on error
	}

	// Apply rule priority logic
	cfg = applyRulePriority(cfg)

	return cfg
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

	// Note: RequireImperative is a bool, so we need to check if it's explicitly set
	// For bool fields, we merge if the overlay has a different value than the default
	if overlay.Message.Subject.RequireImperative != base.Message.Subject.RequireImperative {
		result.Message.Subject.RequireImperative = overlay.Message.Subject.RequireImperative
	}

	if len(overlay.Message.Subject.ForbidEndings) > 0 {
		result.Message.Subject.ForbidEndings = overlay.Message.Subject.ForbidEndings
	}

	// Merge body config
	if overlay.Message.Body.Required != base.Message.Body.Required {
		result.Message.Body.Required = overlay.Message.Body.Required
	}

	if overlay.Message.Body.MinLength != 0 {
		result.Message.Body.MinLength = overlay.Message.Body.MinLength
	}

	if overlay.Message.Body.AllowSignoffOnly != base.Message.Body.AllowSignoffOnly {
		result.Message.Body.AllowSignoffOnly = overlay.Message.Body.AllowSignoffOnly
	}

	if overlay.Message.Body.MinSignoffCount != 0 {
		result.Message.Body.MinSignoffCount = overlay.Message.Body.MinSignoffCount
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

	// Merge Signature config
	if overlay.Signature.KeyDirectory != "" {
		result.Signature.KeyDirectory = overlay.Signature.KeyDirectory
	}

	if len(overlay.Signature.AllowedSigners) > 0 {
		result.Signature.AllowedSigners = overlay.Signature.AllowedSigners
	}

	if overlay.Signature.Required != result.Signature.Required {
		result.Signature.Required = overlay.Signature.Required
	}

	if overlay.Signature.VerifyFormat != result.Signature.VerifyFormat {
		result.Signature.VerifyFormat = overlay.Signature.VerifyFormat
	}

	// Merge Identity config
	if len(overlay.Identity.AllowedAuthors) > 0 {
		result.Identity.AllowedAuthors = overlay.Identity.AllowedAuthors
	}

	return result
}

// findFirstExistingConfigFile finds the first existing config file in search paths.
func findFirstExistingConfigFile() string {
	return findFirstExistingConfigFileInRepo("")
}

// findFirstExistingConfigFileInRepo finds the first existing config file in repository-specific search paths.
func findFirstExistingConfigFileInRepo(repoPath string) string {
	for _, path := range getConfigSearchPathsForRepo(repoPath) {
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
