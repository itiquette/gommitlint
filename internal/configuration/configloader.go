// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package configuration provides configuration loading and access.
package configuration

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/itiquette/gommitlint/internal/core/validation"
	"github.com/itiquette/gommitlint/internal/infrastructure/config"
)

// ConfigManager manages loading and validating configuration.
type ConfigManager struct {
	provider *config.Provider
	loader   *ConfigLoader
	cache    *AppConf
}

// CreateDefaultConfigManager creates a new ConfigManager with default configuration.
func CreateDefaultConfigManager() *ConfigManager {
	// Create config provider from infrastructure layer
	provider, err := config.NewProvider()
	if err != nil {
		// If there's an error, create with empty provider
		provider = &config.Provider{}
	}

	// Create config loader with default paths
	loader := NewConfigLoader(nil) // No validator for now

	return &ConfigManager{
		provider: provider,
		loader:   loader,
	}
}

// GetConfiguration returns the application configuration.
func (c *ConfigManager) GetConfiguration() (*AppConf, error) {
	// Return cached configuration if available
	if c.cache != nil {
		return c.cache, nil
	}

	// Use the Koanf-based loader to load configuration
	if c.loader != nil {
		appConf, err := c.loader.LoadConfiguration()
		if err == nil && appConf != nil {
			// Cache the configuration
			c.cache = appConf

			return appConf, nil
		}
	}

	// Fallback to default configuration
	appConf := DefaultConfiguration()

	// Cache the configuration
	c.cache = appConf

	return appConf, nil
}

// GetRuleConfiguration returns the rule configuration for the validation engine.
func (c *ConfigManager) GetRuleConfiguration() (*validation.RuleConfiguration, error) {
	// First get the application configuration
	appConf, err := c.GetConfiguration()
	if err != nil {
		// Log the error but continue with defaults
		log.Printf("Error loading configuration: %v, using defaults", err)
		return validation.DefaultConfiguration(), nil
	}

	if appConf == nil || appConf.GommitConf == nil {
		return validation.DefaultConfiguration(), nil
	}

	// Use the loaded AppConf to create rule configuration
	conf := appConf.GommitConf

	// Create rule configuration
	ruleConfig := &validation.RuleConfiguration{
		// Subject configuration
		MaxSubjectLength: conf.Subject.MaxLength,

		// Conventional commit configuration
		ConventionalTypes:    conf.ConventionalCommit.Types,
		MaxDescLength:        conf.ConventionalCommit.MaxDescriptionLength,
		ConventionalScopes:   conf.ConventionalCommit.Scopes,
		IsConventionalCommit: conf.ConventionalCommit.Required,

		// Jira configuration - only set if Jira rule is configured
		JiraRequired: false, // Default value
	}

	// Check if Jira rule is configured
	if conf.Subject != nil && conf.Subject.Jira != nil {
		ruleConfig.JiraRequired = conf.Subject.Jira.Required

		// If Jira is not required, add it to the disabled rules
		if !conf.Subject.Jira.Required {
			ruleConfig.DisabledRules = append(ruleConfig.DisabledRules, "JiraReference")
		}
	} else {
		// If Jira rule is not configured, disable it by default
		ruleConfig.DisabledRules = append(ruleConfig.DisabledRules, "JiraReference")
	}

	return ruleConfig, nil
}

// FindConfigFile finds the configuration file in standard locations.
func FindConfigFile() (string, error) {
	// XDG specification locations
	configDirs := []string{
		".", // Current directory
	}

	// Add XDG_CONFIG_HOME if available
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		configDirs = append(configDirs, filepath.Join(xdgConfigHome, "gommitlint"))
	} else {
		// Default XDG location
		home, err := os.UserHomeDir()
		if err == nil {
			configDirs = append(configDirs, filepath.Join(home, ".config", "gommitlint"))
		}
	}

	// Add system-wide config location
	configDirs = append(configDirs, "/etc/gommitlint")

	// File names to check
	configFiles := []string{
		"gommitlint.yaml",
		"gommitlint.yml",
		"gommitlint.json",
		"gommitlint.toml",
		".gommitlint.yaml",
		".gommitlint.yml",
		".gommitlint.json",
		".gommitlint.toml",
	}

	// Check all possible file locations
	for _, dir := range configDirs {
		for _, file := range configFiles {
			path := filepath.Join(dir, file)
			if fileExists(path) {
				return path, nil
			}
		}
	}

	// No config file found
	return "", errors.New("no configuration file found in standard locations")
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil || !os.IsNotExist(err)
}

// LoadConfigFile loads configuration from a file.
func LoadConfigFile(path string) (*config.Config, error) {
	// Check if file exists
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration file not found: %s", path)
		}

		return nil, fmt.Errorf("failed to access configuration file: %w", err)
	}

	// Read file
	_, err = os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse file based on extension
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		// TODO: Implement YAML parsing
		return nil, errors.New("YAML parsing not implemented")
	case ".json":
		// TODO: Implement JSON parsing
		return nil, errors.New("JSON parsing not implemented")
	case ".toml":
		// TODO: Implement TOML parsing
		return nil, errors.New("TOML parsing not implemented")
	default:
		return nil, fmt.Errorf("unsupported configuration file format: %s", ext)
	}
}

// FindGitConfigDir finds the Git configuration directory.
func FindGitConfigDir(startPath string) (string, error) {
	// Use provided path or current directory
	path := startPath
	if path == "" {
		var err error

		path, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Walk up the directory tree to find .git directory
	for {
		gitPath := filepath.Join(path, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return gitPath, nil
		}

		// Check if .git is a file (submodule or worktree)
		if info, err := os.Stat(gitPath); err == nil && !info.IsDir() {
			// TODO: Handle submodules and worktrees
			return gitPath, nil
		}

		// Go up one level
		parent := filepath.Dir(path)
		if parent == path {
			// Reached root directory
			return "", errors.New("not a Git repository (or any of the parent directories)")
		}

		path = parent
	}
}

// LoadGitConfig loads Git configuration for commit validation.
func LoadGitConfig(gitDir string) (*config.Config, error) {
	// Check if hooks directory exists
	hooksDir := filepath.Join(gitDir, "hooks")
	if _, err := os.Stat(hooksDir); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("hooks directory not found: %s", hooksDir)
		}

		return nil, fmt.Errorf("failed to access hooks directory: %w", err)
	}

	// Check if commit-msg hook exists
	commitMsgHook := filepath.Join(hooksDir, "commit-msg")
	if _, err := os.Stat(commitMsgHook); os.IsNotExist(err) {
		// No commit-msg hook
		return nil, fs.ErrNotExist
	}

	// TODO: Parse commit-msg hook to extract configuration

	// Return default configuration for now
	return &config.Config{
		GommitConf: &config.GommitLintConfig{
			Subject: &config.SubjectConfig{
				MaxLength: 100,
			},
			ConventionalCommit: &config.ConventionalConfig{
				MaxDescriptionLength: 72,
				Types: []string{
					"feat", "fix", "docs", "style", "refactor",
					"perf", "test", "build", "ci", "chore", "revert",
				},
				Scopes: []string{},
			},
		},
	}, nil
}
