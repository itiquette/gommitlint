// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestRulePriorityLogic(t *testing.T) {
	// Test domain.IsRuleEnabled implementation via the rules.IsRuleEnabled delegation
	// Tests the canonical rule priority implementation
	tests := []struct {
		name          string
		ruleName      string
		enabledRules  []string
		disabledRules []string
		expected      bool
	}{
		{
			name:          "Explicitly enabled rule should be included",
			ruleName:      "ExplicitlyEnabled",
			enabledRules:  []string{"ExplicitlyEnabled"},
			disabledRules: []string{},
			expected:      true,
		},
		{
			name:          "Explicitly disabled rule should be excluded",
			ruleName:      "ExplicitlyDisabled",
			enabledRules:  []string{},
			disabledRules: []string{"ExplicitlyDisabled"},
			expected:      false,
		},
		{
			name:          "Default-disabled rule should be excluded (JiraReference)",
			ruleName:      "JiraReference",
			enabledRules:  []string{},
			disabledRules: []string{},
			expected:      false,
		},
		{
			name:          "Rule enabled by default and not explicitly disabled should be included (SubjectLength)",
			ruleName:      "SubjectLength",
			enabledRules:  []string{},
			disabledRules: []string{},
			expected:      true,
		},
		{
			name:          "Explicitly enabled rule overrides explicitly disabled rule",
			ruleName:      "ExplicitlyEnabledAndDisabled",
			enabledRules:  []string{"ExplicitlyEnabledAndDisabled"},
			disabledRules: []string{"ExplicitlyEnabledAndDisabled"},
			expected:      true,
		},
		{
			name:          "Explicitly enabled rule overrides default-disabled rule (JiraReference)",
			ruleName:      "JiraReference",
			enabledRules:  []string{"JiraReference"},
			disabledRules: []string{},
			expected:      true,
		},
		{
			name:          "Rule with quotes and whitespace in name should be normalized",
			ruleName:      "RuleWithQuotes",
			enabledRules:  []string{" \"RuleWithQuotes\" "},
			disabledRules: []string{},
			expected:      true,
		},
		{
			name:          "Rule with different capitalization should match",
			ruleName:      "RuleWithDifferentCase",
			enabledRules:  []string{"RuLeWiThDiFfErEnTcAsE"}, // Different case in config
			disabledRules: []string{},
			expected:      true, // Should match regardless of case
		},
	}

	// Run tests
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Call the function directly - no need for indirect logger now
			result := IsRuleEnabled(testCase.ruleName, testCase.enabledRules, testCase.disabledRules)

			// Check the result
			require.Equal(t, testCase.expected, result, "IsRuleEnabled returned unexpected result")
		})
	}
}

func TestMakeRuleMap(t *testing.T) {
	tests := []struct {
		name      string
		ruleNames []string
		expected  map[string]bool
	}{
		{
			name:      "Empty slice",
			ruleNames: []string{},
			expected:  map[string]bool{},
		},
		{
			name:      "Simple rule names",
			ruleNames: []string{"rule1", "rule2", "rule3"},
			expected: map[string]bool{
				"rule1": true,
				"rule2": true,
				"rule3": true,
			},
		},
		{
			name:      "Rule names with quotes and whitespace",
			ruleNames: []string{" \"Rule1\" ", "'Rule2'", "  Rule3  "},
			expected: map[string]bool{
				"rule1": true,
				"rule2": true,
				"rule3": true,
			},
		},
		{
			name:      "Rule names with comments",
			ruleNames: []string{"rule1", "#Rule2", "rule3"},
			expected: map[string]bool{
				"rule1": true,
				"rule3": true,
			},
		},
		{
			name:      "Empty and commented rule names",
			ruleNames: []string{"", "#", "  #Comment  "},
			expected:  map[string]bool{},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Call the domain function
			result := domain.MakeRuleMap(testCase.ruleNames)

			// Filter out empty keys (if any) for consistent testing
			cleanResult := make(map[string]bool)

			for k, v := range result {
				if k != "" {
					cleanResult[k] = v
				}
			}

			// Check the result
			require.Equal(t, testCase.expected, cleanResult, "MakeRuleMap returned unexpected result")
		})
	}
}
