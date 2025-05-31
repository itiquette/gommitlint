// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	configTestdata "github.com/itiquette/gommitlint/internal/config/testdata"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/itiquette/gommitlint/internal/domain/testdata"
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
			name: "jira required in subject",
			configSetup: func() config.Config {
				return configTestdata.NewConfigBuilder().Build()
			},
			commit: func() domain.CommitInfo {
				commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
				commit.Subject = "Add new feature PROJ-123"
				commit.Body = "This is a description"

				return commit
			}(),
			wantErrors:  false,
			description: "Should pass with valid Jira reference in subject",
		},
		{
			name: "jira required in body with body ref enabled",
			configSetup: func() config.Config {
				return configTestdata.NewConfigBuilder().
					WithJiraCheckBody(true).
					Build()
			},
			commit: func() domain.CommitInfo {
				commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
				commit.Subject = "Add new feature"
				commit.Body = "This is a description\n\nRefs: PROJ-123"

				return commit
			}(),
			wantErrors:  false,
			description: "Should pass with valid Jira reference in body",
		},
		{
			name: "jira required but missing",
			configSetup: func() config.Config {
				return configTestdata.NewConfigBuilder().Build()
			},
			commit: func() domain.CommitInfo {
				commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
				commit.Subject = "Add new feature"
				commit.Body = "This is a description"

				return commit
			}(),
			wantErrors:  true,
			description: "Should fail when Jira is required but missing",
		},
		{
			name: "jira with conventional and valid projects",
			configSetup: func() config.Config {
				return configTestdata.NewConfigBuilder().
					WithJiraProjects([]string{"PROJ", "TEAM"}).
					Build()
			},
			commit: func() domain.CommitInfo {
				commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
				commit.Subject = "feat: add login feature PROJ-123"
				commit.Body = "This is a description"

				return commit
			}(),
			wantErrors:  false,
			description: "Should pass with conventional format and valid project",
		},
		{
			name: "jira with conventional and invalid project",
			configSetup: func() config.Config {
				return configTestdata.NewConfigBuilder().
					WithJiraProjects([]string{"PROJ", "TEAM"}).
					Build()
			},
			commit: func() domain.CommitInfo {
				commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
				commit.Subject = "feat: add login feature OTHER-123"
				commit.Body = "This is a description"

				return commit
			}(),
			wantErrors:  true,
			description: "Should fail with conventional format and invalid project",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup config
			cfg := testCase.configSetup()
			ctx := context.Background()

			// Create rule with configuration
			// For this test, assume conventional commit is enabled
			cfg.Rules.Enabled = []string{"conventional"}

			rule := rules.NewJiraReferenceRule(cfg)

			// Execute
			errors := rule.Validate(ctx, testCase.commit)

			// Assert
			if testCase.wantErrors {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
			} else {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			}
		})
	}
}
