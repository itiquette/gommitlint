// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration management for gommitlint.
package config

// This file is reserved for future config validation

/* // Commented out until needed
// validateConfig checks configuration values and returns validation errors.
func validateConfig(config Config) []error {
	var errs []error

	// Subject validation
	if config.Subject.MaxLength <= 0 {
		errs = append(errs, errors.New("subject.max-length must be positive"))
	}

	// Subject case validation
	validCases := map[string]bool{"upper": true, "lower": true, "ignore": true}
	if _, valid := validCases[config.Subject.Case]; !valid && config.Subject.Case != "" {
		errs = append(errs, fmt.Errorf("subject.case must be one of: %s", strings.Join(debugMapKeys(validCases), ", ")))
	}

	// Jira configuration validation
	if config.Subject.Jira.Required {
		if config.Subject.Jira.Pattern == "" {
			errs = append(errs, errors.New("subject.jira.pattern must not be empty when jira references are required"))
		}
	}

	// Conventional commit validation
	if config.Conventional.Required {
		if len(config.Conventional.Types) == 0 {
			errs = append(errs, errors.New("conventional.types cannot be empty when conventional commits are required"))
		}

		if config.Conventional.MaxDescriptionLength <= 0 {
			errs = append(errs, errors.New("conventional.max-description-length must be positive"))
		}
	}

	return errs
}

// mapKeys extracts keys from a map.
func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	return keys
}
*/
