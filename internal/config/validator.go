// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration management for gommitlint.
package config

import (
	stderrors "errors"
	"fmt"
	"regexp"
	"strings"
)

// Maximum allowed lengths to prevent resource exhaustion.
const (
	MaxStringLength  = 1024 // Maximum length for string values
	MaxSliceLength   = 1000 // Maximum length for slices
	MaxMapEntries    = 1000 // Maximum entries in maps
	MaxSubjectLength = 200  // Maximum reasonable subject length
	MaxDescLength    = 1000 // Maximum reasonable description length
	MaxPathLength    = 4096 // Maximum path length
)

// ValidateConfig performs comprehensive validation of the configuration.
// It checks all fields for security issues and returns validation errors.
func ValidateConfig(config Config) []error {
	var errs []error

	// Validate subject configuration
	errs = append(errs, validateSubjectConfig(config.Subject)...)

	// Validate body configuration
	errs = append(errs, validateBodyConfig(config.Body)...)

	// Validate conventional configuration
	errs = append(errs, validateConventionalConfig(config.Conventional)...)

	// Validate spell check configuration
	errs = append(errs, validateSpellCheckConfig(config.SpellCheck)...)

	// Validate security configuration
	errs = append(errs, validateSecurityConfig(config.Security)...)

	// Validate repository configuration
	errs = append(errs, validateRepositoryConfig(config.Repository)...)

	// Validate rules configuration
	errs = append(errs, validateRulesConfig(config.Rules)...)

	return errs
}

// validateSubjectConfig validates subject configuration values.
func validateSubjectConfig(config SubjectConfig) []error {
	var errs []error

	// Validate subject max length
	if config.MaxLength <= 0 {
		errs = append(errs, stderrors.New("subject.max-length must be positive"))
	} else if config.MaxLength > MaxSubjectLength {
		errs = append(errs, fmt.Errorf("subject.max-length exceeds maximum allowed value (%d)", MaxSubjectLength))
	}

	// Validate subject case
	validCases := map[string]bool{"upper": true, "lower": true, "sentence": true, "ignore": true}
	if _, valid := validCases[config.Case]; !valid && config.Case != "" {
		errs = append(errs, fmt.Errorf("subject.case must be one of: %s", strings.Join(mapKeys(validCases), ", ")))
	}

	// Validate invalid suffixes (prevent regex injection)
	if len(config.InvalidSuffixes) > MaxStringLength {
		errs = append(errs, fmt.Errorf("subject.invalid-suffixes exceeds maximum length (%d)", MaxStringLength))
	}

	// Validate Jira configuration
	errs = append(errs, validateJiraConfig(config.Jira)...)

	return errs
}

// validateJiraConfig validates Jira configuration values.
func validateJiraConfig(config JiraConfig) []error {
	var errs []error

	// Jira pattern validation (prevent regex injection)
	if len(config.Pattern) > MaxStringLength {
		errs = append(errs, fmt.Errorf("jira.pattern exceeds maximum length (%d)", MaxStringLength))
	}

	// When Jira references are required, pattern must be provided
	if config.Required && config.Pattern == "" {
		errs = append(errs, stderrors.New("jira.pattern must not be empty when jira references are required"))
	}

	// Validate pattern is a valid regex (if provided)
	if config.Pattern != "" {
		if _, err := regexp.Compile(config.Pattern); err != nil {
			errs = append(errs, fmt.Errorf("jira.pattern is not a valid regular expression: %w", err))
		}
	}

	// Validate projects slice length
	if len(config.Projects) > MaxSliceLength {
		errs = append(errs, fmt.Errorf("jira.projects exceeds maximum allowed entries (%d)", MaxSliceLength))
	}

	// Validate individual project names
	for i, project := range config.Projects {
		if len(project) > MaxStringLength {
			errs = append(errs, fmt.Errorf("jira.projects[%d] exceeds maximum length (%d)", i, MaxStringLength))
		}
	}

	return errs
}

// validateBodyConfig validates body configuration values.
func validateBodyConfig(_ BodyConfig) []error {
	// No additional validation currently needed
	return nil
}

// validateConventionalConfig validates conventional configuration values.
func validateConventionalConfig(config ConventionalConfig) []error {
	var errs []error

	// Validate max description length
	if config.MaxDescriptionLength < 0 {
		errs = append(errs, stderrors.New("conventional.max-description-length cannot be negative"))
	} else if config.MaxDescriptionLength > MaxDescLength {
		errs = append(errs, fmt.Errorf("conventional.max-description-length exceeds maximum allowed value (%d)", MaxDescLength))
	}

	// When conventional commits are required, types should be provided
	if config.Required && len(config.Types) == 0 {
		errs = append(errs, stderrors.New("conventional.types cannot be empty when conventional commits are required"))
	}

	// Validate types slice length
	if len(config.Types) > MaxSliceLength {
		errs = append(errs, fmt.Errorf("conventional.types exceeds maximum allowed entries (%d)", MaxSliceLength))
	}

	// Validate scopes slice length
	if len(config.Scopes) > MaxSliceLength {
		errs = append(errs, fmt.Errorf("conventional.scopes exceeds maximum allowed entries (%d)", MaxSliceLength))
	}

	// Validate individual type values
	for idx, typeName := range config.Types {
		if len(typeName) > MaxStringLength {
			errs = append(errs, fmt.Errorf("conventional.types[%d] exceeds maximum length (%d)", idx, MaxStringLength))
		}
		// Only allow alphanumeric characters and hyphens
		if !isValidIdentifier(typeName) {
			errs = append(errs, fmt.Errorf("conventional.types[%d] contains invalid characters", idx))
		}
	}

	// Validate individual scope values
	for idx, scope := range config.Scopes {
		if len(scope) > MaxStringLength {
			errs = append(errs, fmt.Errorf("conventional.scopes[%d] exceeds maximum length (%d)", idx, MaxStringLength))
		}
		// Only allow alphanumeric characters and hyphens
		if !isValidIdentifier(scope) {
			errs = append(errs, fmt.Errorf("conventional.scopes[%d] contains invalid characters", idx))
		}
	}

	return errs
}

// validateSpellCheckConfig validates spell check configuration values.
func validateSpellCheckConfig(config SpellCheckConfig) []error {
	var errs []error

	// Validate locale
	if len(config.Locale) > MaxStringLength {
		errs = append(errs, fmt.Errorf("spellcheck.locale exceeds maximum length (%d)", MaxStringLength))
	}

	// Validate max errors
	if config.MaxErrors < 0 {
		errs = append(errs, stderrors.New("spellcheck.max-errors cannot be negative"))
	}

	// Validate ignore words slice length
	if len(config.IgnoreWords) > MaxSliceLength {
		errs = append(errs, fmt.Errorf("spellcheck.ignore-words exceeds maximum allowed entries (%d)", MaxSliceLength))
	}

	// Validate custom words map size
	if len(config.CustomWords) > MaxMapEntries {
		errs = append(errs, fmt.Errorf("spellcheck.custom-words exceeds maximum allowed entries (%d)", MaxMapEntries))
	}

	// Validate individual ignore words
	for i, word := range config.IgnoreWords {
		if len(word) > MaxStringLength {
			errs = append(errs, fmt.Errorf("spellcheck.ignore-words[%d] exceeds maximum length (%d)", i, MaxStringLength))
		}
	}

	// Validate custom words entries
	for word, replacement := range config.CustomWords {
		if len(word) > MaxStringLength {
			errs = append(errs, fmt.Errorf("spellcheck.custom-words key exceeds maximum length (%d)", MaxStringLength))
		}

		if len(replacement) > MaxStringLength {
			errs = append(errs, fmt.Errorf("spellcheck.custom-words value exceeds maximum length (%d)", MaxStringLength))
		}
	}

	return errs
}

// validateSecurityConfig validates security configuration values.
func validateSecurityConfig(config SecurityConfig) []error {
	var errs []error

	// Validate allowed signature types
	if len(config.AllowedSignatureTypes) > MaxSliceLength {
		errs = append(errs, fmt.Errorf("security.allowed-signature-types exceeds maximum allowed entries (%d)", MaxSliceLength))
	}

	validTypes := map[string]bool{"gpg": true, "ssh": true}

	for idx, sigType := range config.AllowedSignatureTypes {
		if len(sigType) > MaxStringLength {
			errs = append(errs, fmt.Errorf("security.allowed-signature-types[%d] exceeds maximum length (%d)", idx, MaxStringLength))
		}

		if _, valid := validTypes[sigType]; !valid {
			errs = append(errs, fmt.Errorf("security.allowed-signature-types[%d] must be one of: %s", idx, strings.Join(mapKeys(validTypes), ", ")))
		}
	}

	// Validate identity configuration
	errs = append(errs, validateIdentityConfig(config.Identity)...)

	return errs
}

// validateIdentityConfig validates identity configuration values.
func validateIdentityConfig(config IdentityConfig) []error {
	var errs []error

	// Validate public key URI
	if len(config.PublicKeyURI) > MaxPathLength {
		errs = append(errs, fmt.Errorf("identity.public-key-uri exceeds maximum length (%d)", MaxPathLength))
	}

	return errs
}

// validateRepositoryConfig validates repository configuration values.
func validateRepositoryConfig(config RepositoryConfig) []error {
	var errs []error

	// Validate reference branch
	if len(config.Reference) > MaxStringLength {
		errs = append(errs, fmt.Errorf("repository.reference exceeds maximum length (%d)", MaxStringLength))
	}

	// Validate max commits ahead
	if config.MaxCommitsAhead < 0 {
		errs = append(errs, stderrors.New("repository.max-commits-ahead cannot be negative"))
	}

	return errs
}

// validateRulesConfig validates rules configuration values.
func validateRulesConfig(config RulesConfig) []error {
	var errs []error

	// Validate enabled rules slice length
	if len(config.EnabledRules) > MaxSliceLength {
		errs = append(errs, fmt.Errorf("rules.enabled exceeds maximum allowed entries (%d)", MaxSliceLength))
	}

	// Validate disabled rules slice length
	if len(config.DisabledRules) > MaxSliceLength {
		errs = append(errs, fmt.Errorf("rules.disabled exceeds maximum allowed entries (%d)", MaxSliceLength))
	}

	// Validate individual enabled rule names
	for idx, rule := range config.EnabledRules {
		if len(rule) > MaxStringLength {
			errs = append(errs, fmt.Errorf("rules.enabled[%d] exceeds maximum length (%d)", idx, MaxStringLength))
		}

		if !isValidRuleName(rule) {
			errs = append(errs, fmt.Errorf("rules.enabled[%d] contains invalid characters", idx))
		}
	}

	// Validate individual disabled rule names
	for idx, rule := range config.DisabledRules {
		if len(rule) > MaxStringLength {
			errs = append(errs, fmt.Errorf("rules.disabled[%d] exceeds maximum length (%d)", idx, MaxStringLength))
		}

		if !isValidRuleName(rule) {
			errs = append(errs, fmt.Errorf("rules.disabled[%d] contains invalid characters", idx))
		}
	}

	return errs
}

// isValidIdentifier checks if a string contains only valid identifier characters (alphanumeric and hyphen).
func isValidIdentifier(s string) bool {
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

	return validPattern.MatchString(s)
}

// isValidRuleName checks if a rule name is valid (allows alphanumeric, camelCase).
func isValidRuleName(s string) bool {
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	return validPattern.MatchString(s)
}

// mapKeys extracts keys from a map.
func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	return keys
}
