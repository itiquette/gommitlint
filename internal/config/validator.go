// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration management for gommitlint.
package config

import (
	"errors"
	"fmt"
	"strings"
)

// validateConfig checks configuration values and returns validation errors.
func validateConfig(config *AppConf) []error {
	var errs []error

	// Basic null check
	if config == nil {
		return []error{errors.New("configuration is nil")}
	}

	// Check for GommitConf
	if config.GommitConf == nil {
		return []error{errors.New("configuration is empty or missing gommitlint section")}
	}

	// Subject validation
	if config.GommitConf.Subject != nil {
		// Subject length validation
		if config.GommitConf.Subject.MaxLength <= 0 {
			errs = append(errs, errors.New("subject.max-length must be positive"))
		}

		// Subject case validation
		validCases := map[string]bool{"upper": true, "lower": true, "ignore": true}
		if _, valid := validCases[config.GommitConf.Subject.Case]; !valid {
			errs = append(errs, fmt.Errorf("subject.case must be one of: %s", strings.Join(mapKeys(validCases), ", ")))
		}

		// Jira configuration validation
		if config.GommitConf.Subject.Jira != nil && config.GommitConf.Subject.Jira.Required {
			if config.GommitConf.Subject.Jira.Pattern == "" {
				errs = append(errs, errors.New("subject.jira.pattern must not be empty when jira references are required"))
			}
		}
	} else {
		errs = append(errs, errors.New("subject configuration is required"))
	}

	// Body validation - Critical check that must be enforced
	if config.GommitConf.Body == nil {
		errs = append(errs, errors.New("body configuration is required"))
	}

	// Conventional commit validation
	if config.GommitConf.ConventionalCommit != nil {
		if config.GommitConf.ConventionalCommit.Required {
			if len(config.GommitConf.ConventionalCommit.Types) == 0 {
				errs = append(errs, errors.New("conventional-commit.types cannot be empty when conventional commits are required"))
			}

			if config.GommitConf.ConventionalCommit.MaxDescriptionLength <= 0 {
				errs = append(errs, errors.New("conventional-commit.max-description-length must be positive"))
			}
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
