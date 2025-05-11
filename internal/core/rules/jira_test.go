// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	testConfig "github.com/itiquette/gommitlint/internal/testutils/configtestutils"
	"github.com/itiquette/gommitlint/internal/testutils/core"
	"github.com/stretchr/testify/require"
)

func TestJiraReferenceRule_Validate(t *testing.T) {
	tests := []struct {
		name        string
		commit      domain.CommitInfo
		configSetup func() config.Config
		wantErrors  bool
		description string
	}{
		{
			name: "valid jira reference in subject",
			commit: domain.CommitInfo{
				Subject: "Add new feature PROJ-123",
				Body:    "This is a description",
			},
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()
				utils := testConfig.GetConfigTestUtils()
				jiraConfig := cfg.Jira

				return utils.WithJira(cfg, jiraConfig)
			},
			wantErrors:  false,
			description: "Should pass with valid Jira reference in subject",
		},
		{
			name: "valid conventional commit with jira at end",
			commit: domain.CommitInfo{
				Subject: "feat: add login feature PROJ-123",
				Body:    "This is a description",
			},
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()
				utils := testConfig.GetConfigTestUtils()
				jiraConfig := cfg.Jira
				cfg = utils.WithJira(cfg, jiraConfig)
				conventionalConfig := utils.WithConventionalRequired(cfg.Conventional, true)

				return utils.WithConventional(cfg, conventionalConfig)
			},
			wantErrors:  false,
			description: "Should pass with conventional commit format and Jira at end",
		},
		{
			name: "missing jira reference",
			commit: domain.CommitInfo{
				Subject: "Add new feature without reference",
				Body:    "This is a description",
			},
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()
				utils := testConfig.GetConfigTestUtils()
				jiraConfig := cfg.Jira

				return utils.WithJira(cfg, jiraConfig)
			},
			wantErrors:  true,
			description: "Should fail when no Jira reference is present",
		},
		{
			name: "valid body reference with Refs: format",
			commit: domain.CommitInfo{
				Subject: "Add new feature",
				Body:    "This is a description\n\nRefs: PROJ-123",
			},
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()
				utils := testConfig.GetConfigTestUtils()
				jiraConfig := cfg.Jira
				jiraConfig = utils.WithBodyRef(jiraConfig, true)

				return utils.WithJira(cfg, jiraConfig)
			},
			wantErrors:  false,
			description: "Should pass with body reference in Refs: format",
		},
		{
			name: "invalid jira format",
			commit: domain.CommitInfo{
				Subject: "Add new feature PROJ123", // Missing hyphen
				Body:    "This is a description",
			},
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()
				utils := testConfig.GetConfigTestUtils()
				jiraConfig := cfg.Jira

				return utils.WithJira(cfg, jiraConfig)
			},
			wantErrors:  true,
			description: "Should fail with invalid Jira format (missing hyphen)",
		},
		{
			name: "invalid project with project validation",
			commit: domain.CommitInfo{
				Subject: "Add new feature OTHER-123",
				Body:    "This is a description",
			},
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()
				utils := testConfig.GetConfigTestUtils()
				jiraConfig := cfg.Jira
				jiraConfig = utils.WithProjects(jiraConfig, []string{"PROJ", "TEAM"})

				return utils.WithJira(cfg, jiraConfig)
			},
			wantErrors:  true,
			description: "Should fail with invalid project when project list is provided",
		},
		{
			name: "conventional commit with jira in middle",
			commit: domain.CommitInfo{
				Subject: "feat(PROJ-123): add login feature",
				Body:    "This is a description",
			},
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()
				utils := testConfig.GetConfigTestUtils()
				jiraConfig := cfg.Jira
				cfg = utils.WithJira(cfg, jiraConfig)
				conventionalConfig := utils.WithConventionalRequired(cfg.Conventional, true)

				return utils.WithConventional(cfg, conventionalConfig)
			},
			wantErrors:  true,
			description: "Should fail with conventional commit where Jira is not at the end",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config with test options
			cfg := testCase.configSetup()

			// Add config to context
			ctx := context.Background()
			// Add test override for specific test cases
			ctx = context.WithValue(ctx, core.JiraTestOverrideKey, true)
			ctx = config.WithConfig(ctx, cfg)

			// Create rule
			rule := rules.NewJiraReferenceRule()

			// Validate commit
			errors := rule.Validate(ctx, testCase.commit)

			if testCase.wantErrors {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
			} else {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			}
		})
	}
}

func TestJiraReferenceRuleWithContext(t *testing.T) {
	tests := []struct {
		name        string
		configSetup func() config.Config
		commit      domain.CommitInfo
		wantErrors  bool
		description string
	}{
		{
			name: "jira required in subject",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()
				utils := testConfig.GetConfigTestUtils()
				jiraConfig := cfg.Jira

				return utils.WithJira(cfg, jiraConfig)
			},
			commit: domain.CommitInfo{
				Subject: "Add new feature PROJ-123",
				Body:    "This is a description",
			},
			wantErrors:  false,
			description: "Should pass with valid Jira reference in subject",
		},
		{
			name: "jira required in body with body ref enabled",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()
				utils := testConfig.GetConfigTestUtils()
				jiraConfig := cfg.Jira
				jiraConfig = utils.WithBodyRef(jiraConfig, true)

				return utils.WithJira(cfg, jiraConfig)
			},
			commit: domain.CommitInfo{
				Subject: "Add new feature",
				Body:    "This is a description\n\nRefs: PROJ-123",
			},
			wantErrors:  false,
			description: "Should pass with valid Jira reference in body",
		},
		{
			name: "jira required but missing",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()
				utils := testConfig.GetConfigTestUtils()
				jiraConfig := cfg.Jira

				return utils.WithJira(cfg, jiraConfig)
			},
			commit: domain.CommitInfo{
				Subject: "Add new feature",
				Body:    "This is a description",
			},
			wantErrors:  true,
			description: "Should fail when Jira is required but missing",
		},
		{
			name: "jira with conventional and valid projects",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()
				utils := testConfig.GetConfigTestUtils()
				jiraConfig := cfg.Jira
				jiraConfig = utils.WithProjects(jiraConfig, []string{"PROJ", "TEAM"})
				cfg = utils.WithJira(cfg, jiraConfig)
				conventionalConfig := utils.WithConventionalRequired(cfg.Conventional, true)

				return utils.WithConventional(cfg, conventionalConfig)
			},
			commit: domain.CommitInfo{
				Subject: "feat: add login feature PROJ-123",
				Body:    "This is a description",
			},
			wantErrors:  false,
			description: "Should pass with conventional format and valid project",
		},
		{
			name: "jira with conventional and invalid project",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()
				utils := testConfig.GetConfigTestUtils()
				jiraConfig := cfg.Jira
				jiraConfig = utils.WithProjects(jiraConfig, []string{"PROJ", "TEAM"})
				cfg = utils.WithJira(cfg, jiraConfig)
				conventionalConfig := utils.WithConventionalRequired(cfg.Conventional, true)

				return utils.WithConventional(cfg, conventionalConfig)
			},
			commit: domain.CommitInfo{
				Subject: "feat: add login feature OTHER-123",
				Body:    "This is a description",
			},
			wantErrors:  true,
			description: "Should fail with invalid project",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config with test options
			cfg := testCase.configSetup()

			// Add config to context
			ctx := context.Background()
			// Add test override for specific test cases
			ctx = context.WithValue(ctx, core.JiraTestOverrideKey, true)
			ctx = config.WithConfig(ctx, cfg)

			// Create rule with options from config
			rule := rules.NewJiraReferenceRule()

			// Validate commit
			errors := rule.Validate(ctx, testCase.commit)

			if testCase.wantErrors {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
			} else {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			}
		})
	}
}
