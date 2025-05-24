// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain_test

import (
	"context"
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

// TestRule is a test implementation of the domain.Rule interface specifically for RuleRegistry tests.
type TestRule struct {
	ruleName string
	result   []errors.ValidationError
}

func (r TestRule) Name() string {
	return r.ruleName
}

func (r TestRule) Validate(_ context.Context, _ domain.CommitInfo) []errors.ValidationError {
	return r.result
}

// TestRuleRegistry_InitializeAndGetActiveRules tests the rule initialization and active rule filtering.
func TestRuleRegistry_InitializeAndGetActiveRules(t *testing.T) {
	// Create test context
	ctx := testcontext.CreateTestContext()

	// Test cases
	tests := []struct {
		name            string
		ruleNames       []string
		defaultDisabled map[string]bool
		enabledRules    []string
		disabledRules   []string
		expectedRules   []string
	}{
		{
			name:            "All rules enabled by default",
			ruleNames:       []string{"Rule1", "Rule2", "Rule3"},
			defaultDisabled: map[string]bool{},
			enabledRules:    []string{},
			disabledRules:   []string{},
			expectedRules:   []string{"Rule1", "Rule2", "Rule3"},
		},
		{
			name:            "Some rules default disabled",
			ruleNames:       []string{"Rule1", "Rule2", "Rule3"},
			defaultDisabled: map[string]bool{"Rule2": true},
			enabledRules:    []string{},
			disabledRules:   []string{},
			expectedRules:   []string{"Rule1", "Rule3"},
		},
		{
			name:            "Explicitly enabled overrides default disabled",
			ruleNames:       []string{"Rule1", "Rule2", "Rule3"},
			defaultDisabled: map[string]bool{"Rule2": true},
			enabledRules:    []string{"Rule2"},
			disabledRules:   []string{},
			expectedRules:   []string{"Rule1", "Rule2", "Rule3"},
		},
		{
			name:            "Explicitly disabled overrides default enabled",
			ruleNames:       []string{"Rule1", "Rule2", "Rule3"},
			defaultDisabled: map[string]bool{},
			enabledRules:    []string{},
			disabledRules:   []string{"Rule2"},
			expectedRules:   []string{"Rule1", "Rule3"},
		},
		{
			name:            "Explicitly enabled overrides explicitly disabled",
			ruleNames:       []string{"Rule1", "Rule2", "Rule3"},
			defaultDisabled: map[string]bool{},
			enabledRules:    []string{"Rule2"},
			disabledRules:   []string{"Rule1", "Rule2"},
			expectedRules:   []string{"Rule2", "Rule3"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a new registry with custom default disabled rules
			// Convert rule names to lowercase for the priority service
			lowerDefaultDisabled := make(map[string]bool)
			for rule, disabled := range testCase.defaultDisabled {
				lowerDefaultDisabled[strings.ToLower(rule)] = disabled
			}

			priorityService := domain.NewRulePriorityService(lowerDefaultDisabled)
			registry := domain.NewRuleRegistry(domain.WithPriorityService(priorityService))

			// Register the rules
			for _, name := range testCase.ruleNames {
				// Capture for closure
				ruleName := name // Capture loop variable
				factory := func(_ context.Context) domain.Rule {
					return TestRule{ruleName: ruleName, result: nil}
				}
				registry.RegisterWithContext(ctx, name, factory)
			}

			// Initialize rules
			registry.InitializeRules(ctx)

			// Get active rules
			activeRules := registry.GetActiveRules(ctx, testCase.enabledRules, testCase.disabledRules)

			// Check that we got the expected rules
			require.Equal(t, len(testCase.expectedRules), len(activeRules),
				"Expected %d active rules, got %d", len(testCase.expectedRules), len(activeRules))

			// Verify each expected rule is in the active rules
			activeRuleNames := make(map[string]bool)
			for _, rule := range activeRules {
				activeRuleNames[rule.Name()] = true
			}

			for _, expectedName := range testCase.expectedRules {
				require.True(t, activeRuleNames[expectedName],
					"Expected rule %s to be active", expectedName)
			}

			// Verify there are no unexpected rules
			for _, rule := range activeRules {
				found := false

				for _, expected := range testCase.expectedRules {
					if rule.Name() == expected {
						found = true

						break
					}
				}

				require.True(t, found, "Unexpected rule %s was active", rule.Name())
			}
		})
	}
}

// TestRuleRegistry_CreationDeduplication tests that rules are only created once.
func TestRuleRegistry_CreationDeduplication(t *testing.T) {
	// Create a test context
	ctx := testcontext.CreateTestContext()

	// Create a registry
	registry := domain.NewRuleRegistry()

	// Track rule creation
	creationCount := make(map[string]int)

	ruleNames := []string{"Rule1", "Rule2", "Rule3"}
	for _, name := range ruleNames {
		// Capture for closure
		factory := func(_ context.Context) domain.Rule {
			creationCount[name]++

			return TestRule{ruleName: name, result: nil}
		}
		registry.RegisterWithContext(ctx, name, factory)
	}

	// Initialize rules once
	registry.InitializeRules(ctx)

	// Verify each rule was created exactly once
	for _, name := range ruleNames {
		require.Equal(t, 1, creationCount[name],
			"Rule %s should be created exactly once during initialization", name)
	}

	// Get active rules multiple times
	registry.GetActiveRules(ctx, []string{}, []string{})
	registry.GetActiveRules(ctx, []string{"Rule1"}, []string{})
	registry.GetActiveRules(ctx, []string{}, []string{"Rule2"})

	// Verify the rules weren't created again
	for _, name := range ruleNames {
		require.Equal(t, 1, creationCount[name],
			"Rule %s should not be recreated when filtering active rules", name)
	}
}

// TestRuleRegistry_WithContext tests that the context is correctly passed to the rule.
func TestRuleRegistry_WithContext(t *testing.T) {
	// Create a test context with a logger
	ctx := testcontext.CreateTestContext()

	// Create a registry
	registry := domain.NewRuleRegistry()

	// Track whether context was correctly passed
	contextReceived := false

	// Register a rule factory that checks the context
	registry.RegisterWithContext(ctx, "ContextTestRule", func(ruleCtx context.Context) domain.Rule {
		// Verify that the context has a logger
		logger := contextx.GetLogger(ruleCtx)
		if logger != nil {
			contextReceived = true
		}

		return TestRule{ruleName: "ContextTestRule", result: nil}
	})

	// Initialize the rules
	registry.InitializeRules(ctx)

	// Verify context was correctly passed
	require.True(t, contextReceived, "Context should be passed to rule factory")

	// Get active rules
	activeRules := registry.GetActiveRules(ctx, []string{}, []string{})

	// Verify rule was returned and is active
	require.Len(t, activeRules, 1, "One rule should be active")
	require.Equal(t, "ContextTestRule", activeRules[0].Name(), "Rule name should match")
}
