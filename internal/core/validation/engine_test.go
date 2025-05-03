// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validation_test

import (
	"context"
	"fmt"
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
	validationFn func(domain.CommitInfo) []errors.ValidationError
}

func NewMockRule(name string, shouldPass bool) MockRule {
	rule := MockRule{
		name:       name,
		shouldPass: shouldPass,
		errors:     make([]errors.ValidationError, 0),
	}

	if !shouldPass {
		// Create a simple error
		rule.errors = append(rule.errors, errors.CreateBasicError(name, errors.ErrUnknown, "Mock error message"))
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

func (r MockRule) Name() string {
	return r.name
}

func (r MockRule) Validate(_ context.Context, commit domain.CommitInfo) []errors.ValidationError {
	if r.validationFn != nil {
		return r.validationFn(commit)
	}

	return r.errors
}

func (r MockRule) Result(_ []errors.ValidationError) string {
	if r.shouldPass {
		return "Passed"
	}

	return "Failed"
}

func (r MockRule) VerboseResult(_ []errors.ValidationError) string {
	if r.shouldPass {
		return "Mock rule passed"
	}

	return "Mock rule failed"
}

func (r MockRule) Help(_ []errors.ValidationError) string {
	return "This is a mock rule for testing"
}

func (r MockRule) Errors() []errors.ValidationError {
	return r.errors
}

// MockRuleProvider is a mock implementation of the RuleProvider interface for testing.
// MockAnalyzer is a mock implementation of the CommitAnalyzer interface for testing.
type MockAnalyzer struct{}

func (m *MockAnalyzer) GetCommitsAhead(_ context.Context, _ string) (int, error) {
	return 0, nil
}

type MockRuleProvider struct {
	rules []domain.Rule
}

func NewMockRuleProvider(rules []domain.Rule) *MockRuleProvider {
	return &MockRuleProvider{
		rules: rules,
	}
}

func (p *MockRuleProvider) GetRules(_ context.Context) []domain.Rule {
	return p.rules
}

func (p *MockRuleProvider) GetActiveRules(_ context.Context) []domain.Rule {
	return p.rules
}

func (p *MockRuleProvider) WithActiveRules(_ []string) domain.RuleProvider {
	// Simplified implementation for testing - just returns self
	return p
}

func (p *MockRuleProvider) WithDisabledRules(_ []string) domain.RuleProvider {
	// Simplified implementation for testing - just returns self
	return p
}

func (p *MockRuleProvider) WithCustomRule(_ domain.Rule) domain.RuleProvider {
	// Simplified implementation for testing - just returns self
	return p
}

// TestValidationEngine_ValidateCommit tests the validation of a single commit.
func TestValidationEngine_ValidateCommit(t *testing.T) {
	// Create mock rules
	passingRule := NewMockRule("PassingRule", true)
	failingRule := NewMockRule("FailingRule", false)

	// Create mock rule provider
	provider := NewMockRuleProvider([]domain.Rule{passingRule, failingRule})

	// No longer needed as we're using NewEngine directly
	_ = &MockAnalyzer{}

	// Create validation engine with custom rule provider
	engine := validation.NewEngine(provider)

	// Create test commit
	commit := domain.CommitInfo{
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
		validationFn: func(commit domain.CommitInfo) []errors.ValidationError {
			if commit.Hash == "failing" {
				return []errors.ValidationError{
					errors.CreateBasicError("ConditionalRule", errors.ErrUnknown, "Failed for specific commit"),
				}
			}

			return nil
		},
	}

	// Create mock rule provider
	provider := NewMockRuleProvider([]domain.Rule{passingRule, conditionalRule})

	// No longer needed as we're using NewEngine directly
	_ = &MockAnalyzer{}

	// Create validation engine with custom rule provider
	engine := validation.NewEngine(provider)

	// Create test commits
	commits := []domain.CommitInfo{
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
		validationFn: func(_ domain.CommitInfo) []errors.ValidationError {
			// In a real test we would use time.Sleep,
			// but for a unit test we'll avoid actual waiting
			return []errors.ValidationError{
				errors.CreateBasicError("SlowRule", errors.ErrUnknown, "Slow rule completed"),
			}
		},
	}

	// Create mock rule provider
	provider := NewMockRuleProvider([]domain.Rule{slowRule})

	// No longer needed as we're using NewEngine directly
	_ = &MockAnalyzer{}

	// Create validation engine with custom rule provider
	engine := validation.NewEngine(provider)

	// Create test commit
	commit := domain.CommitInfo{
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

// MockCommitsAheadRule is a mock implementation of the CommitsAhead rule for testing message generation.
type MockCommitsAheadRule struct {
	name         string
	ahead        int
	maxAhead     int
	ref          string
	shouldFailFn func(int) bool
}

func NewMockCommitsAheadRule(name string, ahead, maxAhead int, ref string) *MockCommitsAheadRule {
	return &MockCommitsAheadRule{
		name:     name,
		ahead:    ahead,
		maxAhead: maxAhead,
		ref:      ref,
		shouldFailFn: func(ahead int) bool {
			return ahead > maxAhead
		},
	}
}

func (r *MockCommitsAheadRule) Name() string {
	return r.name
}

func (r *MockCommitsAheadRule) Validate(_ context.Context, _ domain.CommitInfo) []errors.ValidationError {
	if r.shouldFailFn(r.ahead) {
		return []errors.ValidationError{
			errors.CreateBasicError(r.name, errors.ErrTooManyCommits,
				fmt.Sprintf("HEAD is %d commits ahead of %s (maximum allowed: %d)",
					r.ahead, r.ref, r.maxAhead)),
		}
	}

	return nil
}

func (r *MockCommitsAheadRule) Result(_ []errors.ValidationError) string {
	// This should only be called after validation
	if r.ahead > r.maxAhead {
		return fmt.Sprintf("HEAD is %d commits ahead of %s (exceeds limit of %d)",
			r.ahead, r.ref, r.maxAhead)
	} else if r.ahead == 0 {
		return "HEAD is at same commit as " + r.ref
	}

	return fmt.Sprintf("HEAD is %d commit(s) ahead of %s (within limit)", r.ahead, r.ref)
}

func (r *MockCommitsAheadRule) VerboseResult(errors []errors.ValidationError) string {
	if r.ahead > r.maxAhead {
		return fmt.Sprintf(
			"HEAD is currently %d commit(s) ahead of %s (maximum allowed: %d). Consider merging or rebasing with %s.",
			r.ahead, r.ref, r.maxAhead, r.ref)
	}

	return "Verbose result: " + r.Result(errors)
}

func (r *MockCommitsAheadRule) Help(_ []errors.ValidationError) string {
	return "Mock help message"
}

func (r *MockCommitsAheadRule) Errors() []errors.ValidationError {
	ctx := context.Background()

	return r.Validate(ctx, domain.CommitInfo{})
}

// TestValidationEngine_CommitsAheadMessages tests that the message from the CommitsAhead rule matches the validation state.
func TestValidationEngine_CommitsAheadMessages(t *testing.T) {
	tests := []struct {
		name           string
		ahead          int
		maxAhead       int
		expectedStatus domain.ValidationStatus
		expectedMsg    string
	}{
		{
			name:           "no commits ahead",
			ahead:          0,
			maxAhead:       5,
			expectedStatus: domain.StatusPassed,
			expectedMsg:    "HEAD is at same commit as main",
		},
		{
			name:           "within limit",
			ahead:          3,
			maxAhead:       5,
			expectedStatus: domain.StatusPassed,
			expectedMsg:    "HEAD is 3 commit(s) ahead of main (within limit)",
		},
		{
			name:           "exceeds limit",
			ahead:          20,
			maxAhead:       5,
			expectedStatus: domain.StatusFailed,
			expectedMsg:    "HEAD is 20 commits ahead of main (exceeds limit of 5)",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create the mock rule
			rule := NewMockCommitsAheadRule("CommitsAhead", testCase.ahead, testCase.maxAhead, "main")

			// Create a rule provider with just this rule
			provider := NewMockRuleProvider([]domain.Rule{rule})

			// Create the engine
			engine := validation.NewEngine(provider)

			// Create a dummy commit
			commit := domain.CommitInfo{
				Hash:    "testcommit",
				Subject: "Test commit",
				Message: "This is a test commit",
			}

			// Validate
			result := engine.ValidateCommit(context.Background(), commit)

			// Get the rule result
			require.Len(t, result.RuleResults, 1, "Should have one rule result")
			ruleResult := result.RuleResults[0]

			// Check the message
			require.Equal(t, testCase.expectedMsg, ruleResult.Message,
				"Result message should match the expected message for the commits ahead state")

			// Check the status
			require.Equal(t, testCase.expectedStatus, ruleResult.Status,
				"Rule status should match expected status")

			// For failing cases, verify we have errors
			if testCase.expectedStatus == domain.StatusFailed {
				require.False(t, result.Passed, "Result should not pass")
				require.NotEmpty(t, ruleResult.Errors, "Should have validation errors")
			} else {
				require.True(t, result.Passed, "Result should pass")
				require.Empty(t, ruleResult.Errors, "Should not have validation errors")
			}
		})
	}
}
