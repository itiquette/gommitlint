// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package configtestutils contains test utilities for the config package.
// This package is intended for testing purposes only.
package configtestutils

import (
	"github.com/itiquette/gommitlint/internal/config/types"
)

// WithSubject returns a new Config with the updated subject configuration.
func (t *TestUtils) WithSubject(c types.Config, subject types.SubjectConfig) types.Config {
	result := c
	result.Subject = subject

	return result
}

// WithBody returns a new Config with the updated body configuration.
func (t *TestUtils) WithBody(c types.Config, body types.BodyConfig) types.Config {
	result := c
	result.Body = body

	return result
}

// WithConventional returns a new Config with the updated conventional commit configuration.
func (t *TestUtils) WithConventional(c types.Config, conventional types.ConventionalConfig) types.Config {
	result := c
	result.Conventional = conventional

	return result
}

// WithRules returns a new Config with the updated rules configuration.
func (t *TestUtils) WithRules(c types.Config, rules types.RulesConfig) types.Config {
	result := c
	result.Rules = rules

	return result
}

// WithSecurity returns a new Config with the updated security configuration.
func (t *TestUtils) WithSecurity(c types.Config, security types.SecurityConfig) types.Config {
	result := c
	result.Security = security

	return result
}

// WithOutput returns a new Config with the updated output configuration.
func (t *TestUtils) WithOutput(c types.Config, output types.OutputConfig) types.Config {
	result := c
	result.Output = output

	return result
}

// WithSpellCheck returns a new Config with the updated spell check configuration.
func (t *TestUtils) WithSpellCheck(c types.Config, spellcheck types.SpellCheckConfig) types.Config {
	result := c
	result.SpellCheck = spellcheck

	return result
}

// WithJira returns a new Config with the updated JIRA configuration.
func (t *TestUtils) WithJira(c types.Config, jira types.JiraConfig) types.Config {
	result := c
	result.Jira = jira

	return result
}

// SubjectConfig transformations

// WithSubjectMaxLength returns a new SubjectConfig with the updated max length value.
func WithSubjectMaxLength(c types.SubjectConfig, maxLength int) types.SubjectConfig {
	result := c
	result.MaxLength = maxLength

	return result
}

// WithSubjectCase returns a new SubjectConfig with the updated case value.
func WithSubjectCase(c types.SubjectConfig, caseStyle string) types.SubjectConfig {
	result := c
	result.Case = caseStyle

	return result
}

// WithSubjectRequireImperative returns a new SubjectConfig with the updated require imperative flag.
func WithSubjectRequireImperative(c types.SubjectConfig, require bool) types.SubjectConfig {
	result := c
	result.RequireImperative = require

	return result
}

// WithSubjectDisallowedSuffixes returns a new SubjectConfig with the updated disallowed suffixes.
func WithSubjectDisallowedSuffixes(c types.SubjectConfig, suffixes []string) types.SubjectConfig {
	result := c
	result.DisallowedSuffixes = suffixes

	return result
}

// BodyConfig transformations

// WithBodyRequired returns a new BodyConfig with the updated required flag.
func WithBodyRequired(c types.BodyConfig, required bool) types.BodyConfig {
	result := c
	result.Required = required

	return result
}

// WithBodyMinLength returns a new BodyConfig with the updated min length value.
func WithBodyMinLength(c types.BodyConfig, minLength int) types.BodyConfig {
	result := c
	result.MinLength = minLength

	return result
}

// WithBodyMinimumLines returns a new BodyConfig with the updated minimum lines value.
func WithBodyMinimumLines(c types.BodyConfig, minLines int) types.BodyConfig {
	result := c
	result.MinimumLines = minLines

	return result
}

// WithBodyAllowSignOffOnly returns a new BodyConfig with the updated allow sign-off only flag.
func WithBodyAllowSignOffOnly(c types.BodyConfig, allow bool) types.BodyConfig {
	result := c
	result.AllowSignOffOnly = allow

	return result
}

// ConventionalConfig transformations

// WithConventionalRequired returns a new ConventionalConfig with the updated required flag.
func WithConventionalRequired(c types.ConventionalConfig, required bool) types.ConventionalConfig {
	result := c
	result.Required = required

	return result
}

// WithConventionalRequireScope returns a new ConventionalConfig with the updated require scope flag.
func WithConventionalRequireScope(c types.ConventionalConfig, require bool) types.ConventionalConfig {
	result := c
	result.RequireScope = require

	return result
}

// WithConventionalTypes returns a new ConventionalConfig with the updated types.
func WithConventionalTypes(c types.ConventionalConfig, types []string) types.ConventionalConfig {
	result := c
	result.Types = types

	return result
}

// WithConventionalScopes returns a new ConventionalConfig with the updated scopes.
func WithConventionalScopes(c types.ConventionalConfig, scopes []string) types.ConventionalConfig {
	result := c
	result.Scopes = scopes

	return result
}

// WithConventionalAllowBreakingChanges returns a new ConventionalConfig with the updated allow breaking changes flag.
func WithConventionalAllowBreakingChanges(c types.ConventionalConfig, allow bool) types.ConventionalConfig {
	result := c
	result.AllowBreakingChanges = allow

	return result
}

// WithConventionalMaxDescriptionLength returns a new ConventionalConfig with the updated max description length.
func WithConventionalMaxDescriptionLength(c types.ConventionalConfig, maxLength int) types.ConventionalConfig {
	result := c
	result.MaxDescriptionLength = maxLength

	return result
}

// RulesConfig transformations

// WithEnabledRules returns a new RulesConfig with the updated enabled rules.
func WithEnabledRules(c types.RulesConfig, rules []string) types.RulesConfig {
	result := c
	result.EnabledRules = rules

	return result
}

// WithDisabledRules returns a new RulesConfig with the updated disabled rules.
func WithDisabledRules(c types.RulesConfig, rules []string) types.RulesConfig {
	result := c
	result.DisabledRules = rules

	return result
}

// EnableRule adds a rule to the enabled rules list and removes it from disabled rules if present.
func EnableRule(c types.RulesConfig, rule string) types.RulesConfig {
	result := c

	// Check if rule is in the enabled list
	for _, r := range result.EnabledRules {
		if r == rule {
			return result // Already enabled
		}
	}

	// Add to enabled rules
	result.EnabledRules = append(result.EnabledRules, rule)

	// Remove from disabled rules if present
	var newDisabled []string

	for _, r := range result.DisabledRules {
		if r != rule {
			newDisabled = append(newDisabled, r)
		}
	}

	result.DisabledRules = newDisabled

	return result
}

// DisableRule adds a rule to the disabled rules list and removes it from enabled rules if present.
func DisableRule(c types.RulesConfig, rule string) types.RulesConfig {
	result := c

	// Check if rule is in the disabled list
	for _, r := range result.DisabledRules {
		if r == rule {
			return result // Already disabled
		}
	}

	// Add to disabled rules
	result.DisabledRules = append(result.DisabledRules, rule)

	// Remove from enabled rules if present
	var newEnabled []string

	for _, r := range result.EnabledRules {
		if r != rule {
			newEnabled = append(newEnabled, r)
		}
	}

	result.EnabledRules = newEnabled

	return result
}

// SecurityConfig transformations

// WithSignOffRequired returns a new SecurityConfig with the updated sign-off required flag.
func WithSignOffRequired(c types.SecurityConfig, required bool) types.SecurityConfig {
	result := c
	result.SignOffRequired = required

	return result
}

// WithGPGRequired returns a new SecurityConfig with the updated GPG required flag.
func WithGPGRequired(c types.SecurityConfig, required bool) types.SecurityConfig {
	result := c
	result.GPGRequired = required

	return result
}

// OutputConfig transformations

// WithOutputFormat returns a new OutputConfig with the updated format.
func WithOutputFormat(c types.OutputConfig, format string) types.OutputConfig {
	result := c
	result.Format = format

	return result
}

// JiraConfig transformations

// WithJiraPattern returns a new JiraConfig with the updated pattern.
func WithJiraPattern(c types.JiraConfig, pattern string) types.JiraConfig {
	result := c
	result.Pattern = pattern

	return result
}

// WithJiraProjects returns a new JiraConfig with the updated projects.
func WithJiraProjects(c types.JiraConfig, projects []string) types.JiraConfig {
	result := c
	result.Projects = projects

	return result
}

// WithJiraBodyRef returns a new JiraConfig with the updated body reference flag.
func WithJiraBodyRef(c types.JiraConfig, ref bool) types.JiraConfig {
	result := c
	result.BodyRef = ref

	return result
}
