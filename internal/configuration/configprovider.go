// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package configuration provides centralized configuration management.
package configuration

import (
	"fmt"
	"os"
	"strings"

	"github.com/itiquette/gommitlint/internal/core/validation"
	"github.com/itiquette/gommitlint/internal/defaults"
	"github.com/itiquette/gommitlint/internal/errors"
)

// AppConfigProvider provides application configuration from various sources.
type AppConfigProvider struct {
	loader         ConfigLoader
	envProvider    EnvProvider
	defaultsGetter DefaultsGetter
	configCache    *AppConf
}

// EnvProvider defines methods for getting configuration from environment variables.
type EnvProvider interface {
	GetString(key string, defaultValue string) string
	GetBool(key string, defaultValue bool) bool
	GetInt(key string, defaultValue int) int
	GetStringSlice(key string, defaultValue []string) []string
}

// DefaultsGetter defines methods for getting default configuration values.
type DefaultsGetter interface {
	GetDefaultConfig() *AppConf
}

// DefaultEnvProvider implements the EnvProvider interface.
type DefaultEnvProvider struct{}

// NewAppConfigProvider creates a new AppConfigProvider.
func NewAppConfigProvider(loader ConfigLoader) *AppConfigProvider {
	return &AppConfigProvider{
		loader:         loader,
		envProvider:    &DefaultEnvProvider{},
		defaultsGetter: &DefaultDefaults{},
	}
}

// GetString gets a string value from environment or returns default.
func (p *DefaultEnvProvider) GetString(key string, defaultValue string) string {
	// Convert key to environment variable format
	envKey := "GOMMITLINT_" + strings.ToUpper(key)

	if value, exists := os.LookupEnv(envKey); exists {
		return value
	}

	return defaultValue
}

// GetBool gets a boolean value from environment or returns default.
func (p *DefaultEnvProvider) GetBool(key string, defaultValue bool) bool {
	value := p.GetString(key, "")

	if value == "" {
		return defaultValue
	}

	return strings.ToLower(value) == "true" || value == "1" || value == "yes"
}

// GetInt gets an integer value from environment or returns default.
func (p *DefaultEnvProvider) GetInt(key string, defaultValue int) int {
	value := p.GetString(key, "")

	if value == "" {
		return defaultValue
	}

	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return defaultValue
	}

	return result
}

// GetStringSlice gets a string slice value from environment or returns default.
func (p *DefaultEnvProvider) GetStringSlice(key string, defaultValue []string) []string {
	value := p.GetString(key, "")

	if value == "" {
		return defaultValue
	}

	return strings.Split(value, ",")
}

// DefaultDefaults implements the DefaultsGetter interface.
type DefaultDefaults struct{}

// GetDefaultConfig returns the default application configuration.
func (d *DefaultDefaults) GetDefaultConfig() *AppConf {
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
		},
	}
}

// GetConfiguration returns the application configuration.
func (p *AppConfigProvider) GetConfiguration() (*AppConf, error) {
	// Return cached configuration if available
	if p.configCache != nil {
		return p.configCache, nil
	}

	// Start with default configuration
	config := p.defaultsGetter.GetDefaultConfig()

	// Try to load configuration file
	fileConfig, err := p.loader.LoadConfiguration()
	if err != nil {
		// Log warning but continue with defaults
		fmt.Fprintf(os.Stderr, "Warning: Failed to load configuration file: %v\n", err)
	} else if fileConfig != nil {
		config = fileConfig
	}

	// Override with environment variables if provided
	p.applyEnvironmentOverrides(config)

	// Cache the configuration
	p.configCache = config

	return config, nil
}

// ConvertToRuleConfiguration converts the AppConf to RuleConfiguration.
func (p *AppConfigProvider) ConvertToRuleConfiguration(appConf *AppConf) (*validation.RuleConfiguration, error) {
	if appConf == nil || appConf.GommitConf == nil {
		return nil, errors.NewConfigError("invalid configuration", nil)
	}

	// Extract configuration values
	conf := appConf.GommitConf

	// Create rule configuration
	ruleConfig := &validation.RuleConfiguration{
		// Subject configuration
		MaxSubjectLength: conf.Subject.MaxLength,

		// Conventional commit configuration
		ConventionalTypes:    conf.ConventionalCommit.Types,
		MaxDescLength:        conf.ConventionalCommit.MaxDescriptionLength,
		IsConventionalCommit: conf.ConventionalCommit.Required,

		// Jira configuration
		JiraBodyRef: false,

		// SignOff configuration
		RequireSignOff: *conf.SignOffRequired,

		// SubjectCase configuration
		SubjectCaseChoice: conf.Subject.Case,

		// SubjectSuffix configuration
		SubjectInvalidSuffixes: conf.Subject.InvalidSuffixes,

		// CommitBody configuration
		RequireBody: conf.Body.Required,

		// CommitsAhead configuration
		MaxCommitsAhead: 5, // Default value, since NCommitsAhead is a bool in AppConf
		Reference:       "main",
	}

	if conf.Subject.Jira != nil {
		ruleConfig.JiraValidProjects = []string{}
	}

	if conf.Signature != nil {
		ruleConfig.RequireSignature = conf.Signature.Required
		ruleConfig.AllowedSignatureTypes = []string{"gpg", "ssh"}
	}

	return ruleConfig, nil
}

// GetRuleConfiguration returns the rule configuration.
func (p *AppConfigProvider) GetRuleConfiguration() (*validation.RuleConfiguration, error) {
	// Get application configuration
	appConf, err := p.GetConfiguration()
	if err != nil {
		return nil, err
	}

	// Convert to rule configuration
	return p.ConvertToRuleConfiguration(appConf)
}

// applyEnvironmentOverrides applies environment variable overrides to the configuration.
func (p *AppConfigProvider) applyEnvironmentOverrides(config *AppConf) {
	if config == nil || config.GommitConf == nil {
		return
	}

	conf := config.GommitConf

	// Override subject configuration
	conf.Subject.MaxLength = p.envProvider.GetInt("SUBJECT_MAX_LENGTH", conf.Subject.MaxLength)
	conf.Subject.Case = p.envProvider.GetString("SUBJECT_CASE", conf.Subject.Case)
	conf.Subject.InvalidSuffixes = p.envProvider.GetString("SUBJECT_INVALID_SUFFIXES", conf.Subject.InvalidSuffixes)

	// Override conventional commit configuration
	imperativeRequired := p.envProvider.GetBool("SUBJECT_IMPERATIVE_REQUIRED", *conf.Subject.Imperative)
	conf.Subject.Imperative = &imperativeRequired

	conventionalRequired := p.envProvider.GetBool("CONVENTIONAL_COMMIT_REQUIRED", conf.ConventionalCommit.Required)
	conf.ConventionalCommit.Required = conventionalRequired

	// Override signature configuration
	if conf.Signature != nil {
		signatureRequired := p.envProvider.GetBool("SIGNATURE_REQUIRED", conf.Signature.Required)
		conf.Signature.Required = signatureRequired
	}

	// Override sign-off configuration
	signOffRequired := p.envProvider.GetBool("SIGNOFF_REQUIRED", *conf.SignOffRequired)
	conf.SignOffRequired = &signOffRequired

	// Override commits ahead configuration
	nCommitsAheadBool := p.envProvider.GetBool("N_COMMITS_AHEAD", *conf.NCommitsAhead)
	conf.NCommitsAhead = &nCommitsAheadBool
}
