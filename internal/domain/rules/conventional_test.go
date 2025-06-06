// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/itiquette/gommitlint/internal/domain/testdata"
	"github.com/stretchr/testify/require"
)

func TestConventionalCommitRule_Validate(t *testing.T) {
	tests := []struct {
		name        string
		subject     string
		types       []string
		scopes      []string
		wantErrors  bool
		expectedErr string
		description string
	}{
		{
			name:        "valid conventional commit without scope",
			subject:     "feat: add user authentication",
			wantErrors:  false,
			description: "Should pass with valid conventional commit format without scope",
		},
		{
			name:        "valid conventional commit with scope",
			subject:     "fix(auth): resolve login timeout",
			wantErrors:  false,
			description: "Should pass with valid conventional commit format with scope",
		},
		{
			name:        "valid conventional commit with breaking change marker",
			subject:     "feat(api)!: change response format",
			wantErrors:  false,
			description: "Should pass with valid conventional commit format with breaking change marker",
		},
		{
			name:        "invalid format - no colon",
			subject:     "feat add user authentication",
			wantErrors:  true,
			expectedErr: string(domain.ErrInvalidConventionalFormat),
			description: "Should fail with invalid format (missing colon)",
		},
		{
			name:        "empty description allowed",
			subject:     "feat: ",
			wantErrors:  false,
			description: "Empty description is actually allowed by this rule",
		},
		{
			name:        "invalid type with allowed types specified",
			subject:     "unknown: add feature",
			types:       []string{"feat", "fix"},
			wantErrors:  true,
			expectedErr: string(domain.ErrInvalidConventionalType),
			description: "Should fail with invalid commit type when allowed types are specified",
		},
		{
			name:        "valid type with allowed types specified",
			subject:     "feat: add feature",
			types:       []string{"feat", "fix"},
			wantErrors:  false,
			description: "Should pass with valid commit type when allowed types are specified",
		},
		{
			name:        "invalid scope with allowed scopes specified",
			subject:     "feat(unknown): add feature",
			scopes:      []string{"auth", "api"},
			wantErrors:  true,
			expectedErr: string(domain.ErrInvalidConventionalScope),
			description: "Should fail with invalid scope when allowed scopes are specified",
		},
		{
			name:        "valid scope with allowed scopes specified",
			subject:     "feat(auth): add feature",
			scopes:      []string{"auth", "api"},
			wantErrors:  false,
			description: "Should pass with valid scope when allowed scopes are specified",
		},
		{
			name:        "multiple spaces after colon allowed",
			subject:     "feat:  add feature",
			wantErrors:  false,
			description: "Multiple spaces after colon are actually allowed",
		},
		{
			name:        "empty type not allowed",
			subject:     ": add feature",
			wantErrors:  true,
			description: "Empty type is actually not allowed",
		},
		{
			name:        "case sensitive type matching",
			subject:     "FEAT: add feature",
			types:       []string{"feat", "fix"},
			wantErrors:  true,
			description: "Type matching is actually case sensitive",
		},
		// Additional test cases from original
		{
			name:        "valid scope with breaking change",
			subject:     "feat(scope)!: breaking change",
			scopes:      []string{"scope"},
			wantErrors:  false,
			description: "Valid scope with breaking change marker",
		},
		{
			name:        "valid without scope, only type",
			subject:     "docs: update readme",
			wantErrors:  false,
			description: "Valid conventional commit with just type",
		},
		{
			name:        "multiple word description",
			subject:     "feat: add multiple word feature description",
			wantErrors:  false,
			description: "Valid with multi-word description",
		},
		{
			name:        "numeric type not allowed",
			subject:     "v1: initial version",
			wantErrors:  true,
			description: "Numeric types are actually not allowed by this rule",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
			commit.Subject = testCase.subject

			// Create config with options
			cfg := config.Config{
				Conventional: config.ConventionalConfig{
					Types:  testCase.types,
					Scopes: testCase.scopes,
				},
			}

			rule := rules.NewConventionalCommitRule(cfg)

			failures := rule.Validate(commit, cfg)

			// Verify results
			if testCase.wantErrors {
				require.NotEmpty(t, failures, "Expected validation errors but got none for case: %s", testCase.description)

				if testCase.expectedErr != "" {
					// Only check error code if we specified one
					testdata.AssertRuleFailure(t, failures[0], "ConventionalCommit")
				}
			} else {
				require.Empty(t, failures, "Expected no validation errors but got: %v for case: %s", failures, testCase.description)
			}

			// Always verify the rule name
			require.Equal(t, "ConventionalCommit", rule.Name(), "Rule name should be 'ConventionalCommit'")
		})
	}
}

func TestConventionalCommitRuleWithContextConfig(t *testing.T) {
	// Test cases that exercise different config scenarios
	tests := []struct {
		name        string
		subject     string
		types       []string
		scopes      []string
		expectValid bool
		description string
	}{
		{
			name:        "valid conventional commit",
			subject:     "feat: add feature",
			expectValid: true,
			description: "Standard valid conventional commit",
		},
		{
			name:        "invalid conventional commit",
			subject:     "invalid format",
			expectValid: false,
			description: "Invalid conventional commit format",
		},
		{
			name:        "empty description",
			subject:     "feat: ",
			expectValid: true, // Actually allowed
			description: "Conventional commit with empty description",
		},
		{
			name:        "custom types validation",
			subject:     "custom: add feature",
			types:       []string{"custom", "feat"},
			expectValid: true,
			description: "Custom type should be allowed when specified",
		},
		{
			name:        "custom scopes validation",
			subject:     "feat(custom): add feature",
			scopes:      []string{"custom", "api"},
			expectValid: true,
			description: "Custom scope should be allowed when specified",
		},
		{
			name:        "invalid custom type",
			subject:     "badtype: add feature",
			types:       []string{"feat", "fix"},
			expectValid: false,
			description: "Invalid type should fail when types are restricted",
		},
		{
			name:        "invalid custom scope",
			subject:     "feat(badscope): add feature",
			scopes:      []string{"api", "auth"},
			expectValid: false,
			description: "Invalid scope should fail when scopes are restricted",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
			commit.Subject = testCase.subject

			// Create config with options
			cfg := config.Config{
				Conventional: config.ConventionalConfig{
					Types:  testCase.types,
					Scopes: testCase.scopes,
				},
			}

			rule := rules.NewConventionalCommitRule(cfg)

			failures := rule.Validate(commit, cfg)

			if testCase.expectValid {
				require.Empty(t, failures, "Expected no validation errors but got: %v for case: %s", failures, testCase.description)
			} else {
				require.NotEmpty(t, failures, "Expected validation errors but got none for case: %s", testCase.description)
			}
		})
	}
}

func TestConventionalCommitRuleErrorMessages(t *testing.T) {
	// Test specific error message scenarios
	tests := []struct {
		name        string
		subject     string
		expectedErr string
		description string
	}{
		{
			name:        "invalid format produces correct error",
			subject:     "invalid format",
			expectedErr: string(domain.ErrInvalidConventionalFormat),
			description: "Invalid format should produce specific error code",
		},
		{
			name:        "empty description is valid",
			subject:     "feat: x",
			description: "Non-empty description to test valid case",
		},
		{
			name:        "no colon produces error",
			subject:     "feat add feature",
			expectedErr: string(domain.ErrInvalidConventionalFormat),
			description: "Missing colon should produce format error",
		},
		{
			name:        "multiple spaces after colon",
			subject:     "feat:  add feature",
			description: "Multiple spaces are allowed",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
			commit.Subject = testCase.subject

			cfg := config.Config{}
			rule := rules.NewConventionalCommitRule(cfg)

			failures := rule.Validate(commit, cfg)

			if testCase.expectedErr != "" {
				require.NotEmpty(t, failures, "Expected validation errors for case: %s", testCase.description)
				testdata.AssertRuleFailure(t, failures[0], "ConventionalCommit")
			} else {
				// No expected error means we expect it to be valid
				require.Empty(t, failures, "Expected no validation errors for case: %s but got: %v", testCase.description, failures)
			}
		})
	}
}

// Additional edge case tests.
func TestConventionalCommitRuleEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		subject     string
		expectValid bool
		description string
	}{
		{
			name:        "very long type not in default list",
			subject:     "verylongtypename: description",
			expectValid: false,
			description: "Long type names not in default list are rejected",
		},
		{
			name:        "special characters in scope",
			subject:     "feat(api-v2): add feature",
			expectValid: true,
			description: "Scopes with special characters",
		},
		{
			name:        "uppercase breaking change not allowed",
			subject:     "FEAT!: breaking change",
			expectValid: false,
			description: "Uppercase types are not allowed",
		},
		{
			name:        "nested scopes",
			subject:     "feat(api/v2): nested scope format",
			expectValid: true,
			description: "Nested scope notation",
		},
		{
			name:        "numbers in type not allowed",
			subject:     "feat2: add feature",
			expectValid: false,
			description: "Numbers in type names are not allowed",
		},
		{
			name:        "minimal commit with invalid type",
			subject:     "f: x",
			expectValid: false,
			description: "Short type 'f' is not in default allowed list",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
			commit.Subject = testCase.subject

			cfg := config.Config{}
			rule := rules.NewConventionalCommitRule(cfg)

			failures := rule.Validate(commit, cfg)

			if testCase.expectValid {
				require.Empty(t, failures, "Expected no validation errors but got: %v for case: %s", failures, testCase.description)
			} else {
				require.NotEmpty(t, failures, "Expected validation errors but got none for case: %s", testCase.description)
			}
		})
	}
}
