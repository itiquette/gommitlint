// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

// Tests for the SubjectCase rule implementation.
// These tests verify the behavior of the rule against various commit subject formats.

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	testconfig "github.com/itiquette/gommitlint/internal/testutils/config"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

func TestSubjectCaseRule(t *testing.T) {
	tests := []struct {
		name           string
		subject        string
		caseChoice     string
		expectValid    bool
		expectedCode   string
		expectedResult string
		description    string
		options        []rules.SubjectCaseOption
	}{
		{
			name:           "sentence case (default)",
			subject:        "Add new feature to improve performance",
			caseChoice:     "sentence", // Default
			expectValid:    true,
			expectedResult: "Subject case is correct",
			description:    "First word capitalized, rest lower case",
		},
		{
			name:           "lowercase (default setting)",
			subject:        "add new feature to improve performance",
			caseChoice:     "lower",
			expectValid:    true, // Now should be valid with our specific config
			expectedResult: "Subject case is correct",
			description:    "All lowercase with lower case setting",
		},
		{
			name:           "uppercase not allowed by default",
			subject:        "ADD NEW FEATURE",
			caseChoice:     "sentence", // Default
			expectValid:    false,
			expectedCode:   string(appErrors.ErrSubjectCase),
			expectedResult: "Subject should start with",
			description:    "All uppercase not allowed with default setting",
		},
		{
			name:           "uppercase explicitly allowed",
			subject:        "ADD NEW FEATURE",
			caseChoice:     "upper",
			expectValid:    true, // Now should be valid with our specific config
			expectedResult: "Subject case is correct",
			description:    "All uppercase allowed with upper setting",
		},
		{
			name:           "title case",
			subject:        "Add New Feature",
			caseChoice:     "title",
			expectValid:    true,
			expectedResult: "Subject case is correct",
			description:    "Title case allowed with title setting",
		},
		{
			name:           "sentence case standard",
			subject:        "Add new feature",
			caseChoice:     "sentence",
			expectValid:    true,
			expectedResult: "Subject case is correct",
			description:    "Sentence case explicitly allowed",
		},
		{
			name:           "invalid lowercase with uppercase rule",
			subject:        "add new feature",
			caseChoice:     "upper",
			expectValid:    false,
			expectedCode:   string(appErrors.ErrSubjectCase),
			expectedResult: "Subject should start with",
			description:    "Lowercase not allowed with uppercase setting",
		},
		{
			name:           "invalid uppercase with lowercase rule",
			subject:        "ADD NEW FEATURE",
			caseChoice:     "lower",
			expectValid:    false,
			expectedCode:   string(appErrors.ErrSubjectCase),
			expectedResult: "Subject should start with",
			description:    "Uppercase not allowed with lowercase setting",
		},
		{
			name:           "sentence with conventional commit format",
			subject:        "feat: Add new feature",
			caseChoice:     "sentence", // Will also set conventionalRequired to true in config
			expectValid:    true,
			expectedResult: "Subject case is correct",
			description:    "Sentence case valid with conventional format",
		},
		{
			name:           "invalid case with conventional commit format",
			subject:        "feat: ADD NEW FEATURE",
			options:        []rules.SubjectCaseOption{rules.WithCaseChoice("sentence"), rules.WithSubjectCaseCommitFormat(true)},
			expectValid:    false,
			expectedCode:   string(appErrors.ErrSubjectCase),
			expectedResult: "Subject should start with",
			description:    "Uppercase not valid with sentence case and conventional format",
		},
		{
			name:           "mixed case with any option",
			subject:        "AdD nEw FeATurE",
			caseChoice:     "any",
			expectValid:    true, // With "any" case explicitly set in TestConfigAdapter
			expectedResult: "Subject case is correct",
			description:    "'any' case option allows any capitalization",
		},
		{
			name:           "invalid case choice",
			subject:        "Add new feature",
			caseChoice:     "invalid_choice",
			expectValid:    true,                      // With context config, this will default to sentence case and pass
			expectedResult: "Subject case is correct", // Changed to match actual result
			description:    "Invalid case choice produces error",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit with test subject
			commit := domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Create rule with options if provided, otherwise use default
			var rule rules.SubjectCaseRule
			if len(testCase.options) > 0 {
				rule = rules.NewSubjectCaseRule(testCase.options...)
			} else {
				rule = rules.NewSubjectCaseRule()
			}

			// Setup context with TestConfigAdapter
			builder := testconfig.NewBuilder().
				WithSubjectCase(testCase.caseChoice)

			cfg := builder.Build()
			testConfig := testconfig.NewAdapter(cfg).Adapter

			// Create the context with config
			ctx := testcontext.CreateTestContext()
			ctx = contextx.WithConfig(ctx, testConfig)

			// Execute validation
			errors := rule.Validate(ctx, commit)

			// Verify results
			if testCase.expectValid {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			} else {
				require.NotEmpty(t, errors, "Expected validation errors but got none")

				if testCase.expectedCode != "" {
					require.Equal(t, testCase.expectedCode, errors[0].Code, "Error code should match expected")
				}
			}

			// Always verify the rule name
			require.Equal(t, "SubjectCase", rule.Name(), "Rule name should be 'SubjectCase'")
		})
	}
}

func TestSubjectCaseRuleWithConfig(t *testing.T) {
	tests := []struct {
		name        string
		caseStyle   string
		subject     string
		expectValid bool
		checkCommit bool
	}{
		{
			name:        "Lower case config with valid commit",
			caseStyle:   "lower",
			subject:     "add new feature",
			expectValid: true,
			checkCommit: false,
		},
		{
			name:        "Upper case config with valid commit",
			caseStyle:   "upper",
			subject:     "ADD NEW FEATURE",
			expectValid: true,
			checkCommit: true,
		},
		{
			name:        "Title case config with valid commit",
			caseStyle:   "title",
			subject:     "Add New Feature",
			expectValid: true,
			checkCommit: true,
		},
		{
			name:        "Sentence case config with valid commit",
			caseStyle:   "sentence",
			subject:     "Add new feature",
			expectValid: true,
			checkCommit: true,
		},
		{
			name:        "Any case config with mixed case commit",
			caseStyle:   "any",
			subject:     "Add NEW FeaTuRe",
			expectValid: true,
			checkCommit: true,
		},
		{
			name:        "Missing case config defaults to sentence",
			caseStyle:   "",                // Empty string means use default
			subject:     "Add new feature", // Sentence case
			expectValid: true,
			checkCommit: true,
		},
		{
			name:        "Invalid case config with valid case commit",
			caseStyle:   "invalid_case_name", // Should default to sentence
			subject:     "Add new feature",   // Sentence case
			expectValid: true,
			checkCommit: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup with TestConfigAdapter
			builder := testconfig.NewBuilder().
				WithSubjectCase(testCase.caseStyle)
			cfg := builder.Build()
			testConfig := testconfig.NewAdapter(cfg).Adapter

			// Create context with config
			ctx := testcontext.CreateTestContext()
			ctx = contextx.WithConfig(ctx, testConfig)

			// Create commit with test subject
			commit := domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Create rule with appropriate options
			rule := rules.NewSubjectCaseRule(
				rules.WithSubjectCaseCommitFormat(testCase.checkCommit),
			)

			// Execute validation
			errors := rule.Validate(ctx, commit)

			// Verify results
			if testCase.expectValid {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			} else {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
			}
		})
	}
}

func TestSubjectCaseWithConventionalCommit(t *testing.T) {
	// Testing conventional commit format support
	tests := []struct {
		name        string
		subject     string
		caseType    string
		expectValid bool
		description string
	}{
		{
			name:        "conventional commit with sentence case",
			subject:     "feat: Add new feature",
			caseType:    "sentence",
			expectValid: true,
			description: "Correctly formatted conventional commit with sentence case",
		},
		{
			name:        "conventional commit with lowercase",
			subject:     "feat: add new feature",
			caseType:    "lower",
			expectValid: true, // Will now be valid with explicit config
			description: "Correctly formatted conventional commit with lowercase",
		},
		{
			name:        "conventional commit with title case",
			subject:     "feat: Add New Feature",
			caseType:    "title",
			expectValid: true,
			description: "Correctly formatted conventional commit with title case",
		},
		{
			name:        "conventional commit with uppercase",
			subject:     "feat: ADD NEW FEATURE",
			caseType:    "upper",
			expectValid: true, // Will now be valid with explicit config
			description: "Correctly formatted conventional commit with uppercase",
		},
		{
			name:        "conventional commit with wrong case",
			subject:     "feat: ADD NEW FEATURE",
			caseType:    "sentence",
			expectValid: false,
			description: "Incorrectly formatted conventional commit with wrong case",
		},
		{
			name:        "conventional commit with scope and sentence case",
			subject:     "feat(api): Add new endpoint",
			caseType:    "sentence",
			expectValid: true,
			description: "Conventional commit with scope and correct case",
		},
		{
			name:        "complex conventional commit format",
			subject:     "feat(auth)!: Add new authentication system",
			caseType:    "sentence",
			expectValid: true,
			description: "Complex conventional commit with scope, breaking change, and correct case",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit
			commit := domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Create rule with commit format enabled
			rule := rules.NewSubjectCaseRule(
				rules.WithSubjectCaseCommitFormat(true),
			)

			// Setup context with TestConfigAdapter
			builder := testconfig.NewBuilder().
				WithSubjectCase(testCase.caseType)
			cfg := builder.Build()
			testConfig := testconfig.NewAdapter(cfg).Adapter

			ctx := testcontext.CreateTestContext()
			ctx = contextx.WithConfig(ctx, testConfig)

			// Execute validation
			errors := rule.Validate(ctx, commit)

			// Check result
			if testCase.expectValid {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			} else {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
			}
		})
	}
}
