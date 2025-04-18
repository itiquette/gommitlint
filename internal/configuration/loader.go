// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package configuration provides centralized configuration management.
package configuration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itiquette/gommitlint/internal/defaults"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	stderrors "errors"
)

// ConfigLoader handles loading configuration from files.
type ConfigLoader struct {
	validator ConfigValidator
	paths     []string // Prioritized list of configuration paths
}

// NewConfigLoader creates a new configuration loader.
func NewConfigLoader(validator ConfigValidator, paths ...string) *ConfigLoader {
	// If no paths provided, use default search paths
	if len(paths) == 0 {
		paths = getDefaultConfigPaths()
	}

	return &ConfigLoader{
		validator: validator,
		paths:     paths,
	}
}

// LoadConfiguration loads and validates configuration from files.
func (l *ConfigLoader) LoadConfiguration() (*AppConf, error) {
	// Start with default configuration
	config := DefaultConfiguration()

	// Try to load from each path until successful
	loaded := false

	for _, path := range l.paths {
		err := l.loadFromFile(path, config)
		if err == nil {
			loaded = true

			break
		}

		// Log error but continue trying other paths
		fmt.Fprintf(os.Stderr, "Failed to load config from %s: %v\n", path, err)
	}

	// If no configuration files were found, use defaults
	if !loaded && len(l.paths) > 0 {
		// This is just a debug message, not an error
		fmt.Fprintf(os.Stderr, "No configuration files found, using defaults\n")
	}

	// Validate configuration
	if l.validator != nil {
		if errs := l.validator.Validate(config); len(errs) > 0 {
			// Format all validation errors
			var errMsgs []string
			for _, err := range errs {
				errMsgs = append(errMsgs, err.Error())
			}

			return nil, appErrors.NewConfigError("Configuration validation failed: "+strings.Join(errMsgs, "; "), nil)
		}
	}

	return config, nil
}

// LoadFromPath loads configuration from a specific path.
func (l *ConfigLoader) LoadFromPath(path string) (*AppConf, error) {
	// Start with default configuration
	config := DefaultConfiguration()

	// Load from the specified path
	err := l.loadFromFile(path, config)
	if err != nil {
		return nil, appErrors.NewConfigError("Failed to load configuration from "+path, err)
	}

	// Validate configuration
	if l.validator != nil {
		if errs := l.validator.Validate(config); len(errs) > 0 {
			// Format all validation errors
			var errMsgs []string
			for _, err := range errs {
				errMsgs = append(errMsgs, err.Error())
			}

			return nil, appErrors.NewConfigError("Configuration validation failed: "+strings.Join(errMsgs, "; "), nil)
		}
	}

	return config, nil
}

// loadFromFile loads configuration from a specific file.
func (l *ConfigLoader) loadFromFile(path string, config *AppConf) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return stderrors.New("configuration file does not exist")
	}

	// Using koanf for configuration parsing
	konfig := koanf.New(".")

	// Handle different file formats based on extension
	switch {
	case strings.HasSuffix(path, ".yaml"), strings.HasSuffix(path, ".yml"):
		if err := konfig.Load(file.Provider(path), yaml.Parser()); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}
	case strings.HasSuffix(path, ".json"):
		if err := konfig.Load(file.Provider(path), nil); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		return stderrors.New("unsupported configuration file format")
	}

	// Unmarshal into config struct
	if err := konfig.Unmarshal("", config); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return nil
}

// No longer needed as we're using file.Provider directly

// getDefaultConfigPaths returns the default configuration search paths.
func getDefaultConfigPaths() []string {
	// Check for XDG config home
	var xdgConfigPath string
	if xdgHome, exists := os.LookupEnv("XDG_CONFIG_HOME"); exists {
		xdgConfigPath = filepath.Join(xdgHome, defaults.XDGConfigPath)
	} else {
		// Default XDG config path
		home, err := os.UserHomeDir()
		if err == nil {
			xdgConfigPath = filepath.Join(home, ".config", defaults.XDGConfigPath)
		}
	}

	paths := []string{}

	// Add system-wide configuration if it exists
	if _, err := os.Stat("/etc/gommitlint/config.yaml"); err == nil {
		paths = append(paths, "/etc/gommitlint/config.yaml")
	}

	// Add XDG configuration if it exists
	if xdgConfigPath != "" {
		paths = append(paths, xdgConfigPath)
	}

	// Add project-level configuration
	paths = append(paths, defaults.ConfigFileName)

	return paths
}

// DefaultConfiguration returns a default configuration.
func DefaultConfiguration() *AppConf {
	// Initialize with defaults
	imperativeVal := defaults.SubjectImperativeDefault
	signOff := defaults.SignOffRequiredDefault
	jiraRequired := defaults.JIRARequiredDefault
	conventional := defaults.ConventionalCommitRequiredDefault
	ignoreCommits := defaults.IgnoreMergeCommitsDefault
	nCommitsAhead := defaults.NCommitsAheadDefault

	return &AppConf{
		GommitConf: &GommitLintConfig{
			Subject: &SubjectRule{
				Case:            defaults.SubjectCaseDefault,
				Imperative:      &imperativeVal,
				InvalidSuffixes: defaults.SubjectInvalidSuffixesDefault,
				MaxLength:       defaults.SubjectMaxLengthDefault,
				Jira: &JiraRule{
					Required: jiraRequired,
					Pattern:  defaults.JIRAPatternDefault,
				},
			},
			Body: &BodyRule{
				Required: defaults.BodyRequiredDefault,
			},
			ConventionalCommit: &ConventionalRule{
				Types:                defaults.ConventionalCommitTypesDefault,
				MaxDescriptionLength: defaults.ConventionalCommitMaxDescLengthDefault,
				Required:             conventional,
			},
			SpellCheck: &SpellingRule{
				Locale:  defaults.SpellcheckLocaleDefault,
				Enabled: defaults.SpellcheckEnabledDefault,
			},
			Signature: &SignatureRule{
				Required: defaults.SignatureRequiredDefault,
			},
			SignOffRequired:    &signOff,
			NCommitsAhead:      &nCommitsAhead,
			IgnoreMergeCommits: &ignoreCommits,
			Reference:          "main",
		},
	}
}
