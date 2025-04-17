// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package configuration provides centralized configuration management.
package configuration

import (
	"errors"
	"fmt"
	"strings"
)

// ConfigValidator provides validation for configuration values.
type ConfigValidator interface {
	// Validate validates the configuration and returns any errors.
	Validate(config *AppConf) []error
}

// DefaultConfigValidator validates configuration values with sane defaults.
type DefaultConfigValidator struct{}

// NewConfigValidator creates a new DefaultConfigValidator.
func NewConfigValidator() *DefaultConfigValidator {
	return &DefaultConfigValidator{}
}

// Validate checks configuration values and returns validation errors.
func (v *DefaultConfigValidator) Validate(config *AppConf) []error {
	var errs []error

	// Basic null check
	if config == nil || config.GommitConf == nil {
		return []error{errors.New("configuration is nil or empty")}
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

	// Body validation
	if config.GommitConf.Body == nil {
		errs = append(errs, errors.New("body configuration is required"))
	}

	// Signature validation
	if config.GommitConf.Signature != nil && config.GommitConf.Signature.Required {
		// Check if we have an identity configuration when signature verification is required
		if config.GommitConf.Signature.Identity != nil && config.GommitConf.Signature.Identity.PublicKeyURI == "" {
			errs = append(errs, errors.New("signature.identity.public-key-uri must be specified when identity verification is enabled"))
		}
	}

	// Spell check validation
	if config.GommitConf.SpellCheck != nil && config.GommitConf.SpellCheck.Enabled {
		if config.GommitConf.SpellCheck.Locale == "" {
			errs = append(errs, errors.New("spellcheck.locale must be specified when spell checking is enabled"))
		}
	}

	return errs
}

// mapKeys extracts keys from a map.
func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}
