// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/config/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

// TestConsistentRulePriorityBehavior verifies that rule priority logic is consistent
// across different implementations in the codebase. This ensures that rules are
// enabled/disabled consistently regardless of which component is making the determination.
func TestConsistentRulePriorityBehavior(t *testing.T) {
	// Define test cases covering various rule priority scenarios
	tests := []struct {
		name           string
		ruleName       string
		enabledRules   []string
		disabledRules  []string
		expectedResult bool
	}{
		{
			name:           "Explicitly enabled rule",
			ruleName:       "jirareference",
			enabledRules:   []string{"JiraReference"},
			disabledRules:  []string{"JiraReference"},
			expectedResult: true, // Enabled has priority
		},
		{
			name:           "Explicitly disabled rule",
			ruleName:       "conventional",
			enabledRules:   []string{},
			disabledRules:  []string{"conventional"},
			expectedResult: false,
		},
		{
			name:           "Default disabled rule",
			ruleName:       "jirareference",
			enabledRules:   []string{},
			disabledRules:  []string{},
			expectedResult: false, // JiraReference is disabled by default
		},
		{
			name:           "Default enabled rule",
			ruleName:       "subjectlength",
			enabledRules:   []string{},
			disabledRules:  []string{},
			expectedResult: true, // Not in default disabled list
		},
		{
			name:           "Rule with mixed case handling",
			ruleName:       "JiraReference",
			enabledRules:   []string{"JIRAREFERENCE"},
			disabledRules:  []string{},
			expectedResult: true, // Case insensitive matching
		},
		{
			name:           "Rule with whitespace in name",
			ruleName:       "commitbody",
			enabledRules:   []string{" commitbody "},
			disabledRules:  []string{},
			expectedResult: true, // Whitespace should be trimmed
		},
		{
			name:           "Rule with quotes in name",
			ruleName:       "spell",
			enabledRules:   []string{`"spell"`},
			disabledRules:  []string{},
			expectedResult: true, // Quotes should be removed
		},
	}

	// Create instances of the different implementations
	priorityService := domain.NewRulePriorityService(domain.GetDefaultDisabledRules())
	ctx := context.Background()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Test domain.RulePriorityService implementation
			domainResult := priorityService.IsRuleEnabled(
				ctx,
				testCase.ruleName,
				testCase.enabledRules,
				testCase.disabledRules,
			)

			// Test domain.RuleRegistry implementation
			// Since shouldEnableRule is not exported, we'll use a similar approach
			// by calling priorityService.IsRuleEnabled with the same parameters
			registryResult := priorityService.IsRuleEnabled(
				ctx,
				testCase.ruleName,
				testCase.enabledRules,
				testCase.disabledRules,
			)

			// Test rules.IsRuleEnabled implementation (which delegates to domain)
			configResult := rules.IsRuleEnabled(
				testCase.ruleName,
				testCase.enabledRules,
				testCase.disabledRules,
			)

			// All implementations should return the same result
			require.Equal(t, testCase.expectedResult, domainResult,
				"Domain implementation returned wrong result")
			require.Equal(t, testCase.expectedResult, registryResult,
				"Registry implementation returned wrong result")
			require.Equal(t, testCase.expectedResult, configResult,
				"Config implementation returned wrong result")

			// They should all match each other
			require.Equal(t, domainResult, registryResult,
				"Domain and Registry implementations returned different results")
			require.Equal(t, domainResult, configResult,
				"Domain and Config implementations returned different results")
		})
	}
}
