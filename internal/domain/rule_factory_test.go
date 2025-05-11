// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain_test

import (
	"context"
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

// MockRule is a mock implementation of the Rule interface for testing purposes.
type MockRule struct {
	name string
}

// Name returns the rule name.
func (r MockRule) Name() string {
	return r.name
}

// Validate simulates rule validation based on rule name and commit content.
func (r MockRule) Validate(_ context.Context, commit domain.CommitInfo) []errors.ValidationError {
	// Check for specific validation scenarios based on rule type
	switch r.name {
	case "ConventionalCommit":
		// Check if this follows conventional commit format (type: subject)
		if commit.Subject != "" && !containsConventionalFormat(commit.Subject) {
			return []errors.ValidationError{
				errors.CreateBasicError(
					r.Name(),
					"conventional_format",
					"commit message does not follow the conventional format",
				).WithContext("subject", commit.Subject),
			}
		}
	case "SubjectLength":
		// Check if subject is too short (empty or just 1 char)
		if len(commit.Subject) <= 1 {
			return []errors.ValidationError{
				errors.CreateBasicError(
					r.Name(),
					"subject_too_short",
					"commit subject is too short",
				).WithContext("subject", commit.Subject),
			}
		}
		// Check if subject is too long (> 80 chars)
		if len(commit.Subject) > 80 {
			return []errors.ValidationError{
				errors.CreateBasicError(
					r.Name(),
					"subject_too_long",
					"commit subject is too long",
				).WithContext("subject", commit.Subject),
			}
		}
	}
	// Default case - no errors
	return []errors.ValidationError{}
}

// containsConventionalFormat checks if a commit message follows conventional format.
func containsConventionalFormat(subject string) bool {
	// Simple check for "type: message" or "type(scope): message" pattern
	for _, prefix := range []string{"feat:", "fix:", "docs:", "style:", "refactor:", "test:", "chore:"} {
		if strings.HasPrefix(subject, prefix) {
			return true
		}
	}

	// Check for scoped format
	if strings.Contains(subject, "): ") && strings.Contains(subject, "(") {
		return true
	}

	return false
}

// Result returns a simple result message.
func (r MockRule) Result(errors []errors.ValidationError) string {
	if len(errors) > 0 {
		return "Invalid: " + errors[0].Message
	}

	return "Valid"
}

// VerboseResult returns a more detailed result message.
func (r MockRule) VerboseResult(errors []errors.ValidationError) string {
	if len(errors) > 0 {
		return "❌ " + errors[0].Message
	}

	return "✓ Validation passed"
}

// Help returns guidance for the rule.
func (r MockRule) Help(errors []errors.ValidationError) string {
	if len(errors) == 0 {
		return "No errors to fix"
	}

	switch r.name {
	case "ConventionalCommit":
		return "Your commit doesn't follow the conventional commit format.\n" +
			"Use the format: type(scope): description\n" +
			"Example: feat(auth): add login feature"
	case "SubjectLength":
		return "Your commit subject should be between 2 and 80 characters."
	default:
		return "Fix the validation errors to proceed."
	}
}

// HasErrors returns whether this rule has errors.
func (r MockRule) HasErrors() bool {
	return len(r.Errors()) > 0
}

// Errors returns errors for mock rules.
func (r MockRule) Errors() []errors.ValidationError {
	return []errors.ValidationError{}
}

// MockRuleProvider is a testing RuleProvider implementation.
type MockRuleProvider struct {
	rules []domain.Rule
}

// NewMockRuleProvider creates a new mock rule provider for testing.
func NewMockRuleProvider() *MockRuleProvider {
	// Create standard set of mock rules
	rules := []domain.Rule{
		&MockRule{name: "SubjectLength"},
		&MockRule{name: "SubjectCase"},
		&MockRule{name: "CommitBody"},
		&MockRule{name: "ConventionalCommit"},
		&MockRule{name: "SignOff"},
		&MockRule{name: "Signature"},
		&MockRule{name: "JiraReference"},
		&MockRule{name: "ImperativeVerb"},
		&MockRule{name: "Spell"},
		&MockRule{name: "CommitsAhead"},
		&MockRule{name: "SubjectSuffix"},
	}

	return &MockRuleProvider{
		rules: rules,
	}
}

// GetRules returns all rules in the provider.
func (p *MockRuleProvider) GetRules(_ context.Context) []domain.Rule {
	return p.rules
}

// GetActiveRules returns the active rules based on configuration.
func (p *MockRuleProvider) GetActiveRules(ctx context.Context) []domain.Rule {
	// In a real implementation, this would filter rules based on configuration
	return p.GetRules(ctx)
}

// WithActiveRules returns a new provider with the specified active rules.
func (p *MockRuleProvider) WithActiveRules(_ []string) domain.RuleProvider {
	// Create a copy of the provider
	newProvider := &MockRuleProvider{
		rules: make([]domain.Rule, len(p.rules)),
	}
	copy(newProvider.rules, p.rules)

	return newProvider
}

// WithDisabledRules returns a new provider with the specified rules disabled.
func (p *MockRuleProvider) WithDisabledRules(_ []string) domain.RuleProvider {
	// Create a copy of the provider
	newProvider := &MockRuleProvider{
		rules: make([]domain.Rule, len(p.rules)),
	}
	copy(newProvider.rules, p.rules)

	return newProvider
}

// WithCustomRule returns a new provider with the custom rule added.
func (p *MockRuleProvider) WithCustomRule(rule domain.Rule) domain.RuleProvider {
	// Create a copy of the provider
	newProvider := &MockRuleProvider{
		rules: make([]domain.Rule, len(p.rules), len(p.rules)+1),
	}
	copy(newProvider.rules, p.rules)

	// Add the new rule
	newProvider.rules = append(newProvider.rules, rule)

	return newProvider
}

func TestMockRuleProvider(t *testing.T) {
	// A simple test to verify our mock provider works correctly
	provider := NewMockRuleProvider()

	// Check we got the expected rules
	rules := provider.GetRules(context.Background())
	require.NotEmpty(t, rules, "Should have rules")
	require.GreaterOrEqual(t, len(rules), 11, "Should have all expected rules")

	// Verify at least a couple of the expected rule names are present
	var foundSubjectLength, foundConventional bool

	for _, rule := range rules {
		if rule.Name() == "SubjectLength" {
			foundSubjectLength = true
		}

		if rule.Name() == "ConventionalCommit" {
			foundConventional = true
		}
	}

	require.True(t, foundSubjectLength, "Should have SubjectLength rule")
	require.True(t, foundConventional, "Should have ConventionalCommit rule")
}
