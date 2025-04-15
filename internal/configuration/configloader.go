// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/itiquette/gommitlint/internal"
	"github.com/itiquette/gommitlint/internal/defaults"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// DefaultConfigLoader implements the ConfigLoader interface with default behavior.
type DefaultConfigLoader struct{}

// LoadConfiguration loads the application configuration from various sources.
// It reads from the configuration file and returns the populated AppConf.
func (DefaultConfigLoader) LoadConfiguration() (*AppConf, error) {
	// Create AppConf with default values
	ignoreCommits := defaults.IgnoreMergeCommitsDefault
	nCommitsAhead := defaults.NCommitsAheadDefault
	signOff := defaults.SignOffRequiredDefault
	imperativeVal := defaults.SubjectImperativeDefault

	jiraRequired := defaults.JIRARequiredDefault
	conventional := defaults.ConventionalCommitRequiredDefault

	// Initialize with defaults
	appConfig := &AppConf{
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
		},
	}

	if err := ReadConfigurationFile(appConfig, defaults.ConfigFileName); err != nil {
		return nil, internal.NewConfigError(fmt.Errorf("failed to read configuration file: %w", err))
	}

	return appConfig, nil
}

// ReadConfigurationFile loads configuration from XDG config directory or local file.
// It populates the provided appConfiguration with values from the found config files.
// The function follows the XDG Base Directory Specification for configuration file locations.
func ReadConfigurationFile(appConfiguration *AppConf, configfile string) error {
	koanfConf := koanf.New(".")
	xdgConfigfileExists, xdgConfigFilePath := hasXDGConfigFile("XDG_CONFIG_HOME", defaults.XDGConfigPath)
	localConfigfileExists := hasLocalConfigFile(configfile)

	// Load XDG config file if it exists
	if xdgConfigfileExists {
		if err := koanfConf.Load(file.Provider(xdgConfigFilePath), yaml.Parser()); err != nil {
			return internal.NewConfigError(fmt.Errorf("error loading xdg_config_home configuration: %w", err),
				map[string]string{"config_path": xdgConfigFilePath})
		}
	}

	// Load local config file if it exists
	if localConfigfileExists {
		if err := koanfConf.Load(file.Provider(configfile), yaml.Parser()); err != nil {
			return internal.NewConfigError(fmt.Errorf("error loading config: %w", err),
				map[string]string{"config_file": configfile})
		}
	}

	// Return early if no configuration files found
	if !localConfigfileExists && !xdgConfigfileExists {
		return nil
	}

	// Unmarshal the YAML data into the config struct
	if err := koanfConf.Unmarshal("", appConfiguration); err != nil {
		return internal.NewConfigError(fmt.Errorf("error unmarshalling yaml config: %w", err))
	}

	return nil
}

// hasXDGConfigFile checks if a configuration file exists in the XDG config directory.
// Returns whether the file exists and, if so, its full path.
func hasXDGConfigFile(xdgconfighome string, xdgconfighomeconfigpath string) (bool, string) {
	xdgConfigfileExists := false

	var xdgConfigFilePath string

	envValue, xdgHomeIsSet := os.LookupEnv(xdgconfighome)
	if xdgHomeIsSet {
		xdgConfigFilePath = filepath.Join(envValue, xdgconfighomeconfigpath)
		if _, err := os.Stat(xdgConfigFilePath); err == nil {
			xdgConfigfileExists = true
		}
	}

	return xdgConfigfileExists, xdgConfigFilePath
}

// hasLocalConfigFile checks if a configuration file exists in the current directory.
// Returns whether the file exists.
func hasLocalConfigFile(configFile string) bool {
	if _, err := os.Stat(configFile); err == nil {
		return true
	}

	return false
}
