// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestConventionalCommitRule_Validate(t *testing.T) {
	tests := []struct {
		name        string
		commit      domain.CommitInfo
		options     []ConventionalCommitOption
		wantErrors  bool
		description string
	}{
		{
			name: "valid conventional commit without scope",
			commit: domain.CommitInfo{
				Subject: "feat: add user authentication",
			},
			options:     nil,
			wantErrors:  false,
			description: "Should pass with valid conventional commit format without scope",
		},
		{
			name: "valid conventional commit with scope",
			commit: domain.CommitInfo{
				Subject: "fix(auth): resolve login timeout",
			},
			options:     nil,
			wantErrors:  false,
			description: "Should pass with valid conventional commit format with scope",
		},
		{
			name: "valid conventional commit with breaking change marker",
			commit: domain.CommitInfo{
				Subject: "feat(api)!: change response format",
			},
			options:     nil,
			wantErrors:  false,
			description: "Should pass with valid conventional commit format with breaking change marker",
		},
		{
			name: "invalid format - no colon",
			commit: domain.CommitInfo{
				Subject: "feat add user authentication",
			},
			options:     nil,
			wantErrors:  true,
			description: "Should fail with invalid format (missing colon)",
		},
		{
			name: "invalid format - no description",
			commit: domain.CommitInfo{
				Subject: "feat: ",
			},
			options:     nil,
			wantErrors:  true,
			description: "Should fail with invalid format (missing description)",
		},
		{
			name: "invalid type with allowed types specified",
			commit: domain.CommitInfo{
				Subject: "unknown: add feature",
			},
			options: []ConventionalCommitOption{
				WithAllowedTypes([]string{"feat", "fix"}),
			},
			wantErrors:  true,
			description: "Should fail with invalid commit type when allowed types are specified",
		},
		{
			name: "valid type with allowed types specified",
			commit: domain.CommitInfo{
				Subject: "feat: add feature",
			},
			options: []ConventionalCommitOption{
				WithAllowedTypes([]string{"feat", "fix"}),
			},
			wantErrors:  false,
			description: "Should pass with valid commit type when allowed types are specified",
		},
		{
			name: "invalid scope with allowed scopes specified",
			commit: domain.CommitInfo{
				Subject: "feat(unknown): add feature",
			},
			options: []ConventionalCommitOption{
				WithAllowedScopes([]string{"auth", "api"}),
			},
			wantErrors:  true,
			description: "Should fail with invalid commit scope when allowed scopes are specified",
		},
		{
			name: "missing scope when required",
			commit: domain.CommitInfo{
				Subject: "feat: add feature",
			},
			options: []ConventionalCommitOption{
				WithRequiredScope(),
			},
			wantErrors:  true,
			description: "Should fail when scope is required but not provided",
		},
		{
			name: "description too long",
			commit: domain.CommitInfo{
				Subject: "feat: " + string(make([]rune, 100)),
			},
			options: []ConventionalCommitOption{
				WithMaxDescLength(50),
			},
			wantErrors:  true,
			description: "Should fail when description is longer than the configured max length",
		},
		{
			name: "too many spaces after colon",
			commit: domain.CommitInfo{
				Subject: "feat:  add feature",
			},
			options:     nil,
			wantErrors:  true,
			description: "Should fail when there are too many spaces after the colon",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create rule with options
			rule := NewConventionalCommitRule(testCase.options...)

			ctx := context.Background()
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

func TestConventionalCommitRuleWithConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      config.Config
		commit      domain.CommitInfo
		wantErrors  bool
		description string
	}{
		{
			name: "valid commit with default config",
			config: config.NewConfig().
				WithConventionalRequired(true),
			commit: domain.CommitInfo{
				Subject: "feat: add user authentication",
			},
			wantErrors:  false,
			description: "Should pass with valid conventional commit format and default config",
		},
		{
			name: "invalid type with custom allowed types",
			config: config.NewConfig().
				WithConventionalRequired(true).
				WithConventionalTypes([]string{"custom", "update"}),
			commit: domain.CommitInfo{
				Subject: "feat: add feature",
			},
			wantErrors:  true,
			description: "Should fail with invalid commit type when custom allowed types are configured",
		},
		{
			name: "valid type with custom allowed types",
			config: config.NewConfig().
				WithConventionalRequired(true).
				WithConventionalTypes([]string{"custom", "update"}),
			commit: domain.CommitInfo{
				Subject: "custom: add feature",
			},
			wantErrors:  false,
			description: "Should pass with valid commit type when custom allowed types are configured",
		},
		{
			name: "invalid scope with custom allowed scopes",
			config: config.NewConfig().
				WithConventionalRequired(true).
				WithConventionalScopes([]string{"auth", "api"}),
			commit: domain.CommitInfo{
				Subject: "feat(other): add feature",
			},
			wantErrors:  true,
			description: "Should fail with invalid commit scope when custom allowed scopes are configured",
		},
		{
			name: "description too long with custom max length",
			config: config.NewConfig().
				WithConventionalRequired(true).
				WithConventionalMaxDescriptionLength(20),
			commit: domain.CommitInfo{
				Subject: "feat: this description is too long for the configured maximum",
			},
			wantErrors:  true,
			description: "Should fail when description is longer than the configured max length",
		},
	}

	for _, testCase := range tests {
		ctx := context.Background()

		t.Run(testCase.name, func(t *testing.T) {
			// Create rule with unified config
			rule := NewConventionalCommitRuleWithConfig(testCase.config)

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

func TestConventionalCommitRuleErrorMessages(t *testing.T) {
	// Create a rule for testing error messages
	rule := NewConventionalCommitRule()

	// Invalid format error
	invalidFormatCommit := domain.CommitInfo{
		Subject: "invalid format",
	}
	errors, updatedRule := ValidateConventionalWithState(rule, invalidFormatCommit)
	require.NotEmpty(t, errors, "Expected validation errors for invalid format")

	// Check the error messages methods
	require.Equal(t, "Invalid conventional commit format", updatedRule.Result(), "Expected correct result message")
	require.NotEmpty(t, updatedRule.VerboseResult(), "Expected non-empty verbose result")
	require.NotEmpty(t, updatedRule.Help(), "Expected non-empty help text")
}
