// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestJiraReferenceRule_Validate(t *testing.T) {
	tests := []struct {
		name        string
		commit      domain.CommitInfo
		options     []JiraReferenceOption
		wantErrors  bool
		description string
	}{
		{
			name: "valid jira reference in subject",
			commit: domain.CommitInfo{
				Subject: "Add new feature PROJ-123",
				Body:    "This is a description",
			},
			options:     nil,
			wantErrors:  false,
			description: "Should pass with valid Jira reference in subject",
		},
		{
			name: "valid conventional commit with jira at end",
			commit: domain.CommitInfo{
				Subject: "feat: add login feature PROJ-123",
				Body:    "This is a description",
			},
			options: []JiraReferenceOption{
				WithConventionalCommit(),
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
			options:     nil,
			wantErrors:  true,
			description: "Should fail when no Jira reference is present",
		},
		{
			name: "valid body reference with Refs: format",
			commit: domain.CommitInfo{
				Subject: "Add new feature",
				Body:    "This is a description\n\nRefs: PROJ-123",
			},
			options: []JiraReferenceOption{
				WithBodyRefChecking(),
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
			options:     nil,
			wantErrors:  true,
			description: "Should fail with invalid Jira format (missing hyphen)",
		},
		{
			name: "invalid project with project validation",
			commit: domain.CommitInfo{
				Subject: "Add new feature OTHER-123",
				Body:    "This is a description",
			},
			options: []JiraReferenceOption{
				WithValidProjects([]string{"PROJ", "TEAM"}),
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
			options: []JiraReferenceOption{
				WithConventionalCommit(),
			},
			wantErrors:  true,
			description: "Should fail with conventional commit where Jira is not at the end",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create rule with options
			rule := NewJiraReferenceRule(testCase.options...)

			// Validate commit
			errors := rule.Validate(testCase.commit)

			if testCase.wantErrors {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
			} else {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			}
		})
	}
}

func TestJiraReferenceRuleWithConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      config.Config
		commit      domain.CommitInfo
		wantErrors  bool
		description string
	}{
		{
			name: "jira required in subject",
			config: config.NewConfig().
				WithJiraRequired(true),
			commit: domain.CommitInfo{
				Subject: "Add new feature PROJ-123",
				Body:    "This is a description",
			},
			wantErrors:  false,
			description: "Should pass with valid Jira reference in subject",
		},
		{
			name: "jira required in body with body ref enabled",
			config: config.NewConfig().
				WithJiraRequired(true).
				WithJiraBodyRef(true),
			commit: domain.CommitInfo{
				Subject: "Add new feature",
				Body:    "This is a description\n\nRefs: PROJ-123",
			},
			wantErrors:  false,
			description: "Should pass with valid Jira reference in body",
		},
		{
			name: "jira required but missing",
			config: config.NewConfig().
				WithJiraRequired(true),
			commit: domain.CommitInfo{
				Subject: "Add new feature",
				Body:    "This is a description",
			},
			wantErrors:  true,
			description: "Should fail when Jira is required but missing",
		},
		{
			name: "jira with conventional and valid projects",
			config: config.NewConfig().
				WithJiraRequired(true).
				WithJiraProjects([]string{"PROJ", "TEAM"}).
				WithConventionalRequired(true),
			commit: domain.CommitInfo{
				Subject: "feat: add login feature PROJ-123",
				Body:    "This is a description",
			},
			wantErrors:  false,
			description: "Should pass with conventional format and valid project",
		},
		{
			name: "jira with conventional and invalid project",
			config: config.NewConfig().
				WithJiraRequired(true).
				WithJiraProjects([]string{"PROJ", "TEAM"}).
				WithConventionalRequired(true),
			commit: domain.CommitInfo{
				Subject: "feat: add login feature OTHER-123",
				Body:    "This is a description",
			},
			wantErrors:  true,
			description: "Should fail with conventional format but invalid project",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create rule with options
			options := []JiraReferenceOption{}

			// Check if conventional commit format is required
			if testCase.config.ConventionalRequired() {
				options = append(options, WithConventionalCommit())
			}

			// Check if body reference checking is enabled
			if testCase.config.JiraBodyRef() {
				options = append(options, WithBodyRefChecking())
			}

			// Add valid projects if provided
			if projects := testCase.config.JiraProjects(); len(projects) > 0 {
				options = append(options, WithValidProjects(projects))
			}

			rule := NewJiraReferenceRule(options...)

			// Validate commit
			errors := rule.Validate(testCase.commit)

			if testCase.wantErrors {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
			} else {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			}
		})
	}
}
