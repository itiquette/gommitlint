// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"testing"

	infraConfig "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	testConfig "github.com/itiquette/gommitlint/internal/testutils/config"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	testdomain "github.com/itiquette/gommitlint/internal/testutils/domain"
	"github.com/stretchr/testify/require"
)

// COMPLETE FILE WITH ALL TESTS MIGRATED TO BUILDER PATTERN

func TestJiraReferenceRule_Validate(t *testing.T) {
	tests := []struct {
		name        string
		commit      domain.CommitInfo
		configSetup func() types.Config
		wantErrors  bool
		description string
	}{
		{
			name: "jira required in subject",
			configSetup: func() types.Config {
				return testConfig.NewBuilder().Build()
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
			configSetup: func() types.Config {
				return testConfig.NewBuilder().
					WithJiraBodyRef(true).
					Build()
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
			configSetup: func() types.Config {
				return testConfig.NewBuilder().Build()
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
			configSetup: func() types.Config {
				return testConfig.NewBuilder().
					WithJiraProjects([]string{"PROJ", "TEAM"}).
					WithConventionalRequired(true).
					Build()
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
			configSetup: func() types.Config {
				return testConfig.NewBuilder().
					WithJiraProjects([]string{"PROJ", "TEAM"}).
					WithConventionalRequired(true).
					Build()
			},
			commit: domain.CommitInfo{
				Subject: "feat: add login feature OTHER-123",
				Body:    "This is a description",
			},
			wantErrors:  true,
			description: "Should fail with conventional format and invalid project",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup config
			cfg := testCase.configSetup()
			ctx := createJiraTestContext(cfg)

			// Execute
			rule := rules.NewJiraReferenceRule()
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

// Helper functions remain the same.
func createJiraTestContext(cfg types.Config) context.Context {
	ctx := testcontext.CreateTestContext()
	ctx = context.WithValue(ctx, testdomain.JiraTestOverrideKey, true)

	// Use the config directly with the adapter
	adapter := infraConfig.NewAdapter(cfg)

	return contextx.WithConfig(ctx, adapter)
}
