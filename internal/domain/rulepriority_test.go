// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/errors"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

// MockLogger and LogEvent have been removed as part of hexagonal architecture migration.
// Logger is now accessed via context using contextx.GetLogger().

// MockRule implements the Rule interface for testing.
type MockRule struct {
	name    string
	errors  []errors.ValidationError
	hasErrs bool
}

func NewMockRule(name string) *MockRule {
	return &MockRule{name: name}
}

func (r *MockRule) Name() string {
	return r.name
}

func (r *MockRule) Validate(_ context.Context, _ CommitInfo) []errors.ValidationError {
	return r.errors
}

func (r *MockRule) HasErrors() bool {
	return r.hasErrs
}

func (r *MockRule) Errors() []errors.ValidationError {
	return r.errors
}

// For test compatibility.
func (r *MockRule) Result(_ []errors.ValidationError) string {
	return r.name
}

func (r *MockRule) VerboseResult(_ []errors.ValidationError) string {
	return r.name
}

func (r *MockRule) Help(_ []errors.ValidationError) string {
	return r.name
}

func TestRulePriorityService_CleanRuleName(t *testing.T) {
	service := NewRulePriorityService(DefaultDisabledRuleMap{})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "RuleName",
			expected: "RuleName",
		},
		{
			name:     "name with whitespace",
			input:    "  RuleName  ",
			expected: "RuleName",
		},
		{
			name:     "name with double quotes",
			input:    "\"RuleName\"",
			expected: "RuleName",
		},
		{
			name:     "name with single quotes",
			input:    "'RuleName'",
			expected: "RuleName",
		},
		{
			name:     "name with quotes and whitespace",
			input:    " \"RuleName\" ",
			expected: "RuleName",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := service.CleanRuleName(testCase.input)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestRulePriorityService_MakeRuleMap(t *testing.T) {
	service := NewRulePriorityService(DefaultDisabledRuleMap{})

	tests := []struct {
		name      string
		ruleNames []string
		expected  map[string]bool
	}{
		{
			name:      "empty slice",
			ruleNames: []string{},
			expected:  map[string]bool{},
		},
		{
			name:      "simple rule names",
			ruleNames: []string{"Rule1", "Rule2", "Rule3"},
			expected: map[string]bool{
				"Rule1": true,
				"Rule2": true,
				"Rule3": true,
			},
		},
		{
			name:      "rule names with quotes and whitespace",
			ruleNames: []string{" \"Rule1\" ", "'Rule2'", "  Rule3  "},
			expected: map[string]bool{
				"Rule1": true,
				"Rule2": true,
				"Rule3": true,
			},
		},
		{
			name:      "rule names with comments",
			ruleNames: []string{"Rule1", "#Rule2", "Rule3"},
			expected: map[string]bool{
				"Rule1": true,
				"Rule3": true,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := service.MakeRuleMap(testCase.ruleNames)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestRulePriorityService_IsRuleEnabled(t *testing.T) {
	// Setup test context
	ctx := testcontext.CreateTestContext()

	// Create default disabled rules for testing
	defaultDisabled := DefaultDisabledRuleMap{
		"DefaultDisabled": true,
	}

	service := NewRulePriorityService(defaultDisabled)

	tests := []struct {
		name          string
		ruleName      string
		enabledRules  []string
		disabledRules []string
		expected      bool
	}{
		{
			name:          "explicitly enabled rule should be included",
			ruleName:      "ExplicitlyEnabled",
			enabledRules:  []string{"ExplicitlyEnabled"},
			disabledRules: []string{},
			expected:      true,
		},
		{
			name:          "explicitly disabled rule should be excluded",
			ruleName:      "ExplicitlyDisabled",
			enabledRules:  []string{},
			disabledRules: []string{"ExplicitlyDisabled"},
			expected:      false,
		},
		{
			name:          "default-disabled rule should be excluded",
			ruleName:      "DefaultDisabled",
			enabledRules:  []string{},
			disabledRules: []string{},
			expected:      false,
		},
		{
			name:          "rule enabled by default and not explicitly disabled should be included",
			ruleName:      "DefaultEnabled",
			enabledRules:  []string{},
			disabledRules: []string{},
			expected:      true,
		},
		{
			name:          "explicitly disabled rule overrides explicitly enabled rule",
			ruleName:      "ExplicitlyEnabledAndDisabled",
			enabledRules:  []string{"ExplicitlyEnabledAndDisabled"},
			disabledRules: []string{"ExplicitlyEnabledAndDisabled"},
			expected:      false,
		},
		{
			name:          "explicitly enabled rule overrides default-disabled rule",
			ruleName:      "DefaultDisabled",
			enabledRules:  []string{"DefaultDisabled"},
			disabledRules: []string{},
			expected:      true,
		},
		{
			name:          "rule with quotes and whitespace in name should be normalized",
			ruleName:      "RuleWithQuotes",
			enabledRules:  []string{" \"RuleWithQuotes\" "},
			disabledRules: []string{},
			expected:      true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := service.IsRuleEnabled(ctx, testCase.ruleName, testCase.enabledRules, testCase.disabledRules)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestRulePriorityService_FilterRules(t *testing.T) {
	// Setup test context
	ctx := testcontext.CreateTestContext()

	// Create default disabled rules for testing
	defaultDisabled := DefaultDisabledRuleMap{
		"DefaultDisabled": true,
	}

	service := NewRulePriorityService(defaultDisabled)

	// Create test rules
	allRules := []Rule{
		NewMockRule("DefaultEnabled"),
		NewMockRule("DefaultDisabled"),
		NewMockRule("ExplicitlyEnabled"),
		NewMockRule("ExplicitlyDisabled"),
		NewMockRule("ExplicitlyEnabledAndDisabled"),
	}

	tests := []struct {
		name          string
		enabledRules  []string
		disabledRules []string
		expectedNames []string
	}{
		{
			name:          "default behavior",
			enabledRules:  []string{},
			disabledRules: []string{},
			expectedNames: []string{"DefaultEnabled", "ExplicitlyEnabled", "ExplicitlyDisabled", "ExplicitlyEnabledAndDisabled"},
		},
		{
			name:          "explicitly enable default-disabled rule",
			enabledRules:  []string{"DefaultDisabled"},
			disabledRules: []string{},
			expectedNames: []string{"DefaultEnabled", "DefaultDisabled", "ExplicitlyEnabled", "ExplicitlyDisabled", "ExplicitlyEnabledAndDisabled"},
		},
		{
			name:          "explicitly disable some rules",
			enabledRules:  []string{},
			disabledRules: []string{"ExplicitlyDisabled", "DefaultEnabled"},
			expectedNames: []string{"ExplicitlyEnabled", "ExplicitlyEnabledAndDisabled"},
		},
		{
			name:          "disabled takes precedence over enabled",
			enabledRules:  []string{"ExplicitlyEnabledAndDisabled", "DefaultDisabled"},
			disabledRules: []string{"ExplicitlyEnabledAndDisabled"},
			expectedNames: []string{"DefaultEnabled", "DefaultDisabled", "ExplicitlyEnabled", "ExplicitlyDisabled"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			filtered := service.FilterRules(ctx, allRules, testCase.enabledRules, testCase.disabledRules)

			// We'll verify rule presence directly without extracting names first

			// Check that we have the expected number of rules
			require.Len(t, filtered, len(testCase.expectedNames), "Filtered rules count doesn't match expected")

			// Check that all expected rule names are present
			for _, expectedName := range testCase.expectedNames {
				found := false

				for _, rule := range filtered {
					if rule.Name() == expectedName {
						found = true

						break
					}
				}

				require.True(t, found, "Expected rule %s not found in filtered rules", expectedName)
			}
		})
	}
}
