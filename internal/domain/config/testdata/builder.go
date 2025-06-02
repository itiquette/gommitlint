// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package testdata provides test fixtures and builders for config tests.
package testdata

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/stretchr/testify/require"
)

// ConfigBuilder provides a functional builder for creating test configurations.
type ConfigBuilder struct {
	config config.Config
}

// NewConfigBuilder creates a new config builder with sensible defaults.
func NewConfigBuilder() ConfigBuilder {
	return ConfigBuilder{
		config: config.Config{
			Rules: config.RulesConfig{
				Enabled:  []string{},
				Disabled: []string{},
			},
			Jira: config.JiraConfig{
				ProjectPrefixes:      []string{},
				RequireInBody:        false,
				RequireInSubject:     false,
				IgnoreTicketPatterns: []string{},
			},
			Signing: config.SigningConfig{
				AllowedSigners: []string{},
			},
		},
	}
}

// WithJiraProjects sets the allowed Jira project prefixes.
func (b ConfigBuilder) WithJiraProjects(projects []string) ConfigBuilder {
	b.config.Jira.ProjectPrefixes = projects

	return b
}

// WithJiraCheckBody enables checking the commit body for Jira references.
func (b ConfigBuilder) WithJiraCheckBody(checkBody bool) ConfigBuilder {
	b.config.Jira.RequireInBody = checkBody

	return b
}

// WithJiraProjectPrefixes sets the Jira project prefixes (replaces pattern-based approach).
func (b ConfigBuilder) WithJiraProjectPrefixes(prefixes []string) ConfigBuilder {
	b.config.Jira.ProjectPrefixes = prefixes

	return b
}

// WithEnabledRules sets the list of enabled rules.
func (b ConfigBuilder) WithEnabledRules(rules []string) ConfigBuilder {
	b.config.Rules.Enabled = rules

	return b
}

// WithDisabledRules sets the list of disabled rules.
func (b ConfigBuilder) WithDisabledRules(rules []string) ConfigBuilder {
	b.config.Rules.Disabled = rules

	return b
}

// WithAllowedSigners sets the allowed signers for identity verification.
func (b ConfigBuilder) WithAllowedSigners(signers []string) ConfigBuilder {
	b.config.Signing.AllowedSigners = signers

	return b
}

// Build returns the constructed config.
func (b ConfigBuilder) Build() config.Config {
	return b.config
}

// CreateTestConfig creates a test configuration file.
func CreateTestConfig(t *testing.T, dir, content string) string {
	t.Helper()

	configPath := filepath.Join(dir, ".gommitlint.yaml")
	err := os.WriteFile(configPath, []byte(content), 0600)
	require.NoError(t, err)

	return configPath
}