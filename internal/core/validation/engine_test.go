// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validation_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/validation"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRule is a mock implementation of the Rule interface for testing.
type MockRule struct {
	name         string
	shouldPass   bool
	errors       []*domain.ValidationError
	validationFn func(*domain.CommitInfo) []*domain.ValidationError
}

func NewMockRule(name string, shouldPass bool) *MockRule {
	rule := &MockRule{
		name:       name,
		shouldPass: shouldPass,
		errors:     make([]*domain.ValidationError, 0),
	}

	if !shouldPass {
		rule.errors = append(rule.errors, domain.NewValidationError(name, "mock_error", "Mock error message"))
	}

	return rule
}

func (r *MockRule) Name() string {
	return r.name
}

func (r *MockRule) Validate(commit *domain.CommitInfo) []*domain.ValidationError {
	if r.validationFn != nil {
		return r.validationFn(commit)
	}

	return r.errors
}

func (r *MockRule) Result() string {
	if r.shouldPass {
		return "Passed"
	}

	return "Failed"
}

func (r *MockRule) VerboseResult() string {
	if r.shouldPass {
		return "Mock rule passed"
	}

	return "Mock rule failed"
}

func (r *MockRule) Help() string {
	return "This is a mock rule for testing"
}

func (r *MockRule) Errors() []*domain.ValidationError {
	return r.errors
}

// MockRuleProvider is a mock implementation of the RuleProvider interface for testing.
type MockRuleProvider struct {
	rules []domain.Rule
}

func NewMockRuleProvider(rules []domain.Rule) *MockRuleProvider {
	return &MockRuleProvider{
		rules: rules,
	}
}

func (p *MockRuleProvider) GetRules() []domain.Rule {
	return p.rules
}

func (p *MockRuleProvider) GetActiveRules() []domain.Rule {
	return p.rules
}

// TestValidationEngine_ValidateCommit tests the validation of a single commit.
func TestValidationEngine_ValidateCommit(t *testing.T) {
	// Create mock rules
	passingRule := NewMockRule("PassingRule", true)
	failingRule := NewMockRule("FailingRule", false)

	// Create mock rule provider
	provider := NewMockRuleProvider([]domain.Rule{passingRule, failingRule})

	// Create validation engine
	engine := validation.NewEngine(provider)

	// Create test commit
	commit := &domain.CommitInfo{
		Hash:    "abcdef",
		Subject: "Test commit",
		Message: "Test commit message",
	}

	// Validate commit
	result := engine.ValidateCommit(context.Background(), commit)

	// Assert result
	require.False(t, result.Passed, "Validation should fail with one failing rule")
	require.Len(t, result.RuleResults, 2, "Should have results for both rules")

	// Verify rule results
	for _, ruleResult := range result.RuleResults {
		if ruleResult.RuleID == "PassingRule" {
			assert.Equal(t, domain.StatusPassed, ruleResult.Status, "Passing rule should have passed status")
			assert.Empty(t, ruleResult.Errors, "Passing rule should have no errors")
		} else if ruleResult.RuleID == "FailingRule" {
			assert.Equal(t, domain.StatusFailed, ruleResult.Status, "Failing rule should have failed status")
			require.NotEmpty(t, ruleResult.Errors, "Failing rule should have errors")
			assert.Equal(t, "Mock error message", ruleResult.Errors[0].Message, "Error message should match")
		}
	}
}

// TestValidationEngine_ValidateCommits tests the validation of multiple commits.
func TestValidationEngine_ValidateCommits(t *testing.T) {
	// Create mock rules
	passingRule := NewMockRule("PassingRule", true)

	// Create a conditional rule that fails only for specific commits
	conditionalRule := &MockRule{
		name: "ConditionalRule",
		validationFn: func(commit *domain.CommitInfo) []*domain.ValidationError {
			if commit.Hash == "failing" {
				return []*domain.ValidationError{
					domain.NewValidationError("ConditionalRule", "conditional_error", "Failed for specific commit"),
				}
			}

			return nil
		},
	}

	// Create mock rule provider
	provider := NewMockRuleProvider([]domain.Rule{passingRule, conditionalRule})

	// Create validation engine
	engine := validation.NewEngine(provider)

	// Create test commits
	commits := []*domain.CommitInfo{
		{
			Hash:    "passing1",
			Subject: "Passing commit 1",
			Message: "This commit should pass",
		},
		{
			Hash:    "failing",
			Subject: "Failing commit",
			Message: "This commit should fail",
		},
		{
			Hash:    "passing2",
			Subject: "Passing commit 2",
			Message: "This commit should also pass",
		},
	}

	// Validate commits
	results := engine.ValidateCommits(context.Background(), commits)

	// Assert results
	require.Equal(t, 3, results.TotalCommits, "Should have validated all commits")
	require.Equal(t, 2, results.PassedCommits, "Two commits should have passed")
	require.False(t, results.AllPassed(), "Not all commits passed")

	// Verify commit results
	for _, commitResult := range results.CommitResults {
		if commitResult.CommitInfo.Hash == "failing" {
			assert.False(t, commitResult.Passed, "Failing commit should have failed")
		} else {
			assert.True(t, commitResult.Passed, "Passing commit should have passed")
		}
	}

	// Verify rule summary
	assert.Equal(t, 1, results.RuleSummary["ConditionalRule"], "ConditionalRule should have failed once")
	assert.Equal(t, 0, results.RuleSummary["PassingRule"], "PassingRule should not have failed")
}

// TestValidationEngine_ContextCancellation tests that validation can be cancelled via context.
func TestValidationEngine_ContextCancellation(t *testing.T) {
	// Create mock rules
	slowRule := &MockRule{
		name: "SlowRule",
		validationFn: func(_ *domain.CommitInfo) []*domain.ValidationError {
			return nil // We'll use a cancelled context, so this won't matter
		},
	}

	// Create mock rule provider
	provider := NewMockRuleProvider([]domain.Rule{slowRule})

	// Create validation engine
	engine := validation.NewEngine(provider)

	// Create test commit
	commit := &domain.CommitInfo{
		Hash:    "test",
		Subject: "Test commit",
		Message: "Test commit message",
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Validate commit with cancelled context
	result := engine.ValidateCommit(ctx, commit)

	// The implementation should handle cancellation gracefully
	assert.NotNil(t, result, "Should still return a result even with cancelled context")
}

// TestDefaultRuleProvider tests the default rule provider implementation.
func TestDefaultRuleProvider(t *testing.T) {
	// Create default configuration
	config := validation.DefaultConfiguration()

	// Create provider
	provider := validation.NewDefaultRuleProvider(config)

	// Verify rules are initialized
	allRules := provider.GetRules()
	require.NotEmpty(t, allRules, "Should have at least one rule")

	// Verify active rules match all rules by default
	activeRules := provider.GetActiveRules()
	assert.Equal(t, len(allRules), len(activeRules), "All rules should be active by default")

	// Test enabling specific rules
	provider.SetActiveRules([]string{"SubjectLength"})
	activeRules = provider.GetActiveRules()
	assert.Len(t, activeRules, 1, "Only specified rule should be active")
	assert.Equal(t, "SubjectLength", activeRules[0].Name(), "Active rule should match specified name")

	// Test disabling specific rules
	provider.SetActiveRules([]string{}) // Reset to all rules
	provider.DisableRules([]string{"ConventionalCommit"})

	activeRules = provider.GetActiveRules()
	for _, rule := range activeRules {
		assert.NotEqual(t, "ConventionalCommit", rule.Name(), "Disabled rule should not be active")
	}

	// Test getting rule by name
	rule := provider.GetRuleByName("SubjectLength")
	assert.NotNil(t, rule, "Should find rule by name")
	assert.Equal(t, "SubjectLength", rule.Name(), "Rule name should match")
}
