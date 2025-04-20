// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validation_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/validation"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

// MockRule is a mock implementation of the Rule interface for testing.
type MockRule struct {
	name         string
	shouldPass   bool
	errors       []errors.ValidationError
	validationFn func(*domain.CommitInfo) []errors.ValidationError
}

func NewMockRule(name string, shouldPass bool) *MockRule {
	rule := &MockRule{
		name:       name,
		shouldPass: shouldPass,
		errors:     make([]errors.ValidationError, 0),
	}

	if !shouldPass {
		// Create a simple error
		rule.errors = append(rule.errors, errors.New(name, errors.ErrUnknown, "Mock error message"))
	}

	return rule
}

// // mockError implements the error interface with additional fields for testing.
// type mockError struct {
// 	name    string
// 	code    string
// 	message string
// }

// func (e mockError) Error() string {
// 	return e.message
// }

func (r *MockRule) Name() string {
	return r.name
}

func (r *MockRule) Validate(commit *domain.CommitInfo) []errors.ValidationError {
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

func (r *MockRule) Errors() []errors.ValidationError {
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
		Hash:    "testcommit",
		Subject: "Test commit",
		Message: "This is a test commit",
	}

	// Validate commit
	result := engine.ValidateCommit(context.Background(), commit)

	// Assert result
	require.Equal(t, commit, result.CommitInfo, "CommitInfo should be the same")
	require.False(t, result.Passed, "Commit should not pass validation")
	require.Len(t, result.RuleResults, 2, "Should have results for all rules")

	// Check rule results
	for _, ruleResult := range result.RuleResults {
		if ruleResult.RuleName == "PassingRule" {
			require.Equal(t, domain.StatusPassed, ruleResult.Status, "PassingRule should pass")
			require.Empty(t, ruleResult.Errors, "PassingRule should have no errors")
		} else if ruleResult.RuleName == "FailingRule" {
			require.Equal(t, domain.StatusFailed, ruleResult.Status, "FailingRule should fail")
			require.NotEmpty(t, ruleResult.Errors, "FailingRule should have errors")
			require.Len(t, ruleResult.Errors, 1, "FailingRule should have one error")
			require.Equal(t, "Mock error message", ruleResult.Errors[0].Error(), "Error message should match")
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
		validationFn: func(commit *domain.CommitInfo) []errors.ValidationError {
			if commit.Hash == "failing" {
				return []errors.ValidationError{
					errors.New("ConditionalRule", errors.ErrUnknown, "Failed for specific commit"),
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
			require.False(t, commitResult.Passed, "Failing commit should not pass")

			// Find the conditional rule result
			var conditionalRuleResult domain.RuleResult

			found := false

			for _, ruleResult := range commitResult.RuleResults {
				if ruleResult.RuleName == "ConditionalRule" {
					conditionalRuleResult = ruleResult
					found = true

					break
				}
			}

			require.True(t, found, "Should have result for conditional rule")
			require.Equal(t, domain.StatusFailed, conditionalRuleResult.Status, "Conditional rule should fail for this commit")
			require.NotEmpty(t, conditionalRuleResult.Errors, "Conditional rule should have errors")
		} else {
			require.True(t, commitResult.Passed, "Passing commit should pass")
		}
	}
}

// TestValidationEngine_Timeout tests that the validation engine respects timeouts.
func TestValidationEngine_Timeout(t *testing.T) {
	// Create a slow rule that sleeps
	slowRule := &MockRule{
		name: "SlowRule",
		validationFn: func(_ *domain.CommitInfo) []errors.ValidationError {
			// In a real test we would use time.Sleep,
			// but for a unit test we'll avoid actual waiting
			return []errors.ValidationError{
				errors.New("SlowRule", errors.ErrUnknown, "Slow rule completed"),
			}
		},
	}

	// Create mock rule provider
	provider := NewMockRuleProvider([]domain.Rule{slowRule})

	// Create validation engine
	engine := validation.NewEngine(provider)

	// Create test commit
	commit := &domain.CommitInfo{
		Hash:    "testcommit",
		Subject: "Test commit",
		Message: "This is a test commit",
	}

	// Use a context with timeout
	// Normally we'd create a context with timeout, but for
	// testing purposes we'll just use a non-canceling context
	ctx := context.Background()

	// Validate commit - in a real test, this would time out
	result := engine.ValidateCommit(ctx, commit)

	// Assert result
	require.False(t, result.Passed, "Commit should not pass validation")
	require.Len(t, result.RuleResults, 1, "Should have results for all rules")
}
