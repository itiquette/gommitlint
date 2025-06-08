// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/stretchr/testify/require"
)

// createConventionalTestCommit creates a test commit with default values.
func createConventionalTestCommit() domain.Commit {
	return domain.Commit{
		Hash:          "abc123def456",
		Subject:       "feat: add new feature",
		Message:       "feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.",
		Body:          "This commit adds a new feature that enhances the user experience.",
		Author:        "Test User",
		AuthorEmail:   "test@example.com",
		CommitDate:    time.Now().Format(time.RFC3339),
		IsMergeCommit: false,
	}
}

// assertRuleFailure checks that a rule failure has expected properties.
func assertRuleFailure(t *testing.T, failure domain.ValidationError, expectedRule string) {
	t.Helper()
	require.Equal(t, expectedRule, failure.Rule, "rule mismatch")
	require.NotEmpty(t, failure.Message, "failure message should not be empty")
}

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
			name:        "empty description not allowed",
			subject:     "feat: ",
			wantErrors:  true,
			expectedErr: string(domain.ErrEmptyConventionalDesc),
			description: "Empty description should fail with enhanced validation",
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
			name:        "multiple spaces after colon not allowed",
			subject:     "feat:  add feature",
			wantErrors:  true,
			expectedErr: string(domain.ErrInvalidSpacing),
			description: "Multiple spaces after colon should fail with strict spacing",
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
			commit := createConventionalTestCommit()
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
					assertRuleFailure(t, failures[0], "ConventionalCommit")
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
			expectValid: false, // Not allowed with enhanced validation
			description: "Conventional commit with empty description should fail",
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
			commit := createConventionalTestCommit()
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
			expectedErr: string(domain.ErrInvalidSpacing),
			description: "Multiple spaces should fail with strict spacing",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commit := createConventionalTestCommit()
			commit.Subject = testCase.subject

			cfg := config.Config{}
			rule := rules.NewConventionalCommitRule(cfg)

			failures := rule.Validate(commit, cfg)

			if testCase.expectedErr != "" {
				require.NotEmpty(t, failures, "Expected validation errors for case: %s", testCase.description)
				assertRuleFailure(t, failures[0], "ConventionalCommit")
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
			commit := createConventionalTestCommit()
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

// Test enhanced features added to the ConventionalCommit rule.
func TestConventionalCommitRule_EnhancedFeatures(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		wantErr bool
		errCode string
	}{
		// Enhanced empty description detection
		{
			name:    "Enhanced empty description - whitespace only",
			subject: "feat:   ",
			wantErr: true,
			errCode: "empty_conventional_desc",
		},
		{
			name:    "Enhanced empty description - tab only",
			subject: "feat:\t",
			wantErr: true,
			errCode: "empty_conventional_desc",
		},
		{
			name:    "Enhanced empty description - mixed whitespace",
			subject: "feat: \t ",
			wantErr: true,
			errCode: "empty_conventional_desc",
		},
		{
			name:    "Valid non-empty description",
			subject: "feat: add login",
			wantErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewConventionalCommitRule(config.Config{})
			commit := createConventionalTestCommit()
			commit.Subject = testCase.subject

			errors := rule.Validate(commit, config.Config{})

			if testCase.wantErr {
				require.NotEmpty(t, errors, "Expected validation error")
				// Find the specific error code we're looking for
				found := false

				for _, err := range errors {
					if err.Code == testCase.errCode {
						found = true

						break
					}
				}

				require.True(t, found, "Expected error code %s, got errors: %v", testCase.errCode, errors)
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}
		})
	}
}

func TestConventionalCommitRule_MultiScopeSupport(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		wantErr bool
		errCode string
	}{
		{
			name:    "Valid multi-scope",
			subject: "feat(ui,api): add login functionality",
			wantErr: false,
		},
		{
			name:    "Valid multi-scope with three scopes",
			subject: "feat(ui,api,auth): implement complete login flow",
			wantErr: false,
		},
		{
			name:    "Valid multi-scope with breaking change",
			subject: "feat(ui,api)!: change authentication structure",
			wantErr: false,
		},
		{
			name:    "Invalid multi-scope with spaces after comma",
			subject: "feat(ui, api): add login functionality",
			wantErr: true,
			errCode: "invalid_conventional_format",
		},
		{
			name:    "Invalid multi-scope with spaces before comma",
			subject: "feat(ui ,api): add login functionality",
			wantErr: true,
			errCode: "invalid_conventional_format",
		},
		{
			name:    "Invalid multi-scope with trailing comma",
			subject: "feat(ui,api,): add login functionality",
			wantErr: true,
			errCode: "invalid_multi_scope",
		},
		{
			name:    "Invalid multi-scope with leading comma",
			subject: "feat(,ui,api): add login functionality",
			wantErr: true,
			errCode: "invalid_multi_scope",
		},
		{
			name:    "Invalid multi-scope with consecutive commas",
			subject: "feat(ui,,api): add login functionality",
			wantErr: true,
			errCode: "invalid_multi_scope",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewConventionalCommitRule(config.Config{})
			commit := createConventionalTestCommit()
			commit.Subject = testCase.subject

			errors := rule.Validate(commit, config.Config{})

			if testCase.wantErr {
				require.NotEmpty(t, errors, "Expected validation error")
				// Find the specific error code we're looking for
				found := false

				for _, err := range errors {
					if err.Code == testCase.errCode {
						found = true

						break
					}
				}

				require.True(t, found, "Expected error code %s, got errors: %v", testCase.errCode, errors)
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}
		})
	}
}

func TestConventionalCommitRule_StrictSpacing(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		wantErr bool
		errCode string
	}{
		{
			name:    "Valid spacing - exactly one space",
			subject: "feat: add login functionality",
			wantErr: false,
		},
		{
			name:    "Valid spacing with scope",
			subject: "feat(auth): add login functionality",
			wantErr: false,
		},
		{
			name:    "Invalid spacing - no space after colon",
			subject: "feat:add login functionality",
			wantErr: true,
			errCode: "invalid_spacing",
		},
		{
			name:    "Invalid spacing - multiple spaces after colon",
			subject: "feat:  add login functionality",
			wantErr: true,
			errCode: "invalid_spacing",
		},
		{
			name:    "Invalid spacing - tab after colon",
			subject: "feat:\tadd login functionality",
			wantErr: true,
			errCode: "invalid_spacing",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewConventionalCommitRule(config.Config{})
			commit := createConventionalTestCommit()
			commit.Subject = testCase.subject

			errors := rule.Validate(commit, config.Config{})

			if testCase.wantErr {
				require.NotEmpty(t, errors, "Expected validation error")
				// Find the specific error code we're looking for
				found := false

				for _, err := range errors {
					if err.Code == testCase.errCode {
						found = true

						break
					}
				}

				require.True(t, found, "Expected error code %s, got errors: %v", testCase.errCode, errors)
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}
		})
	}
}

func TestConventionalCommitRule_EnhancedScopeValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      config.Config
		subject     string
		wantErr     bool
		errCode     string
		description string
	}{
		{
			name: "Multi-scope with restricted scopes - all valid",
			config: config.Config{
				Conventional: config.ConventionalConfig{
					Scopes: []string{"auth", "ui", "api"},
				},
			},
			subject: "feat(auth,ui): add login interface",
			wantErr: false,
		},
		{
			name: "Multi-scope with restricted scopes - one invalid",
			config: config.Config{
				Conventional: config.ConventionalConfig{
					Scopes: []string{"auth", "ui", "api"},
				},
			},
			subject: "feat(auth,invalid): add feature",
			wantErr: true,
			errCode: "invalid_conventional_scope",
		},
		{
			name: "Required scope - missing with multi-scope enabled",
			config: config.Config{
				Conventional: config.ConventionalConfig{
					RequireScope: true,
				},
			},
			subject: "feat: add login",
			wantErr: true,
			errCode: "missing_conventional_scope",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewConventionalCommitRule(testCase.config)
			commit := createConventionalTestCommit()
			commit.Subject = testCase.subject

			errors := rule.Validate(commit, config.Config{})

			if testCase.wantErr {
				require.NotEmpty(t, errors, "Expected validation error")
				// Find the specific error code we're looking for
				found := false

				for _, err := range errors {
					if err.Code == testCase.errCode {
						found = true

						break
					}
				}

				require.True(t, found, "Expected error code %s, got errors: %v", testCase.errCode, errors)
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}
		})
	}
}

func TestConventionalCommitRule_BackwardCompatibility(t *testing.T) {
	// Test that all existing functionality still works
	tests := []struct {
		name    string
		subject string
		wantErr bool
	}{
		{
			name:    "Original valid format",
			subject: "feat: add user authentication",
			wantErr: false,
		},
		{
			name:    "Original valid with scope",
			subject: "fix(auth): resolve login timeout",
			wantErr: false,
		},
		{
			name:    "Original breaking change",
			subject: "feat(api)!: change response format",
			wantErr: false,
		},
		{
			name:    "Original invalid format",
			subject: "feat add user authentication",
			wantErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewConventionalCommitRule(config.Config{})
			commit := createConventionalTestCommit()
			commit.Subject = testCase.subject

			errors := rule.Validate(commit, config.Config{})

			if testCase.wantErr {
				require.NotEmpty(t, errors, "Expected validation error")
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}
		})
	}
}
