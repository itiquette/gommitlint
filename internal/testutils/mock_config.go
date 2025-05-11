// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package testutils provides common utilities for testing across the application.
// It is intended to be used only by tests and should not be imported by production code.
package testutils

import (
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/config/types"
)

// Config related test helpers
// ==========================

// MockConfig creates a default configuration for testing.
func MockConfig() types.Config {
	return config.DefaultConfig()
}

// WithBody returns a new Config with the updated body configuration.
func WithBody(c types.Config, body types.BodyConfig) types.Config {
	result := c
	result.Body = body

	return result
}

// WithBodyConfig creates a body configuration for testing with common options.
func WithBodyConfig(required bool, allowSignOffOnly bool) types.BodyConfig {
	result := types.BodyConfig{
		Required:         required,
		AllowSignOffOnly: allowSignOffOnly,
		MinLength:        10,
		MinimumLines:     3,
	}

	return result
}

// WithJira returns a new Config with the updated jira configuration.
func WithJira(c types.Config, jira types.JiraConfig) types.Config {
	result := c
	result.Jira = jira

	return result
}

// WithJiraConfig creates a jira configuration for testing with common options.
func WithJiraConfig(_ bool, bodyRef bool, projects []string) types.JiraConfig {
	result := types.JiraConfig{
		Pattern: "[A-Z]+-\\d+",
		BodyRef: bodyRef,
	}

	if len(projects) > 0 {
		result.Projects = make([]string, len(projects))
		copy(result.Projects, projects)
	}

	return result
}

// WithConventional returns a new Config with the updated conventional configuration.
func WithConventional(c types.Config, conventional types.ConventionalConfig) types.Config {
	result := c
	result.Conventional = conventional

	return result
}

// WithConventionalConfig creates a conventional configuration for testing with common options.
func WithConventionalConfig(required bool, requireScope bool) types.ConventionalConfig {
	result := types.ConventionalConfig{
		Required:     required,
		RequireScope: requireScope,
		Types:        []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"},
	}

	return result
}
