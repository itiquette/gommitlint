// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Example usage of the new architecture.
// This file is for demonstration purposes only and is not intended to be used in production.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/core/validation"
	"github.com/itiquette/gommitlint/internal/domain"
)

func main() {
	// Create context
	ctx := context.Background()

	// For demonstration, use a manually created commit
	commit := &domain.CommitInfo{
		Hash:    "example",
		Subject: "feat(core): implement hexagonal architecture",
		Body:    "This commit restructures the application to follow a hexagonal architecture pattern. The new architecture provides better separation of concerns, improved testability, and more maintainable code.",
		Message: "feat(core): implement hexagonal architecture\n\nThis commit restructures the application to follow a hexagonal architecture pattern. The new architecture provides better separation of concerns, improved testability, and more maintainable code.",
	}

	// Create rule configuration
	config := &validation.RuleConfiguration{
		// Configure subject validation
		MaxSubjectLength: 72,

		// Configure conventional commit validation
		ConventionalTypes: []string{
			"feat", "fix", "docs", "style", "refactor",
			"perf", "test", "build", "ci", "chore", "revert",
		},
		ConventionalScopes: []string{}, // Empty means allow any scope
		MaxDescLength:      100,
	}

	// Create rule provider with rules
	ruleProvider := validation.NewDefaultRuleProvider(config)

	// Create validation engine
	engine := validation.NewEngine(ruleProvider)

	// Validate commit
	result := engine.ValidateCommit(ctx, commit)

	// Print validation results
	printResults(result)

	// Exit with appropriate status code
	if !result.Passed {
		os.Exit(1)
	}
}

// printResults prints the validation results to the console.
func printResults(result domain.CommitResult) {
	fmt.Printf("Commit: %s\n", result.CommitInfo.Hash)
	fmt.Printf("Subject: %s\n", result.CommitInfo.Subject)
	fmt.Println()

	// Print validation results for each rule
	for _, ruleResult := range result.RuleResults {
		status := "✅"
		if ruleResult.Status == domain.StatusFailed {
			status = "❌"
		}

		fmt.Printf("%s %s: %s\n", status, ruleResult.RuleName, ruleResult.Message)

		// Print errors for failed rules
		if ruleResult.Status == domain.StatusFailed {
			for _, err := range ruleResult.Errors {
				fmt.Printf("   - %s\n", err.Message)
			}

			// Print help message
			if ruleResult.HelpMessage != "" {
				fmt.Printf("   Help: %s\n", ruleResult.HelpMessage)
			}
		}
	}

	// Print overall status
	fmt.Println()

	if result.Passed {
		fmt.Println("✅ Commit validation passed")
	} else {
		fmt.Println("❌ Commit validation failed")
	}
}

// Example of manually creating and validating a commit
func exampleManualValidation() {
	// Create a commit info object
	commit := &domain.CommitInfo{
		Hash:    "example",
		Subject: "feat(core): add new validation rule",
		Body:    "This commit adds a new validation rule for...",
		Message: "feat(core): add new validation rule\n\nThis commit adds a new validation rule for...",
	}

	// Create validation rules
	subjectRule := rules.NewSubjectLengthRule(72)
	conventionalRule := rules.NewConventionalCommitRule(
		[]string{"feat", "fix", "docs"},
		[]string{"core", "ui"},
		100,
	)

	// Create context
	ctx := context.Background()

	// Validate subject length
	subjectErrors := subjectRule.Validate(commit)
	fmt.Printf("Subject length validation: %d errors\n", len(subjectErrors))

	// Validate conventional commit format
	conventionalErrors := conventionalRule.Validate(commit)
	fmt.Printf("Conventional commit validation: %d errors\n", len(conventionalErrors))

	// Create a rule provider with these rules
	ruleProvider := &customRuleProvider{
		rules: []domain.Rule{subjectRule, conventionalRule},
	}

	// Create validation engine
	engine := validation.NewEngine(ruleProvider)

	// Validate commit
	result := engine.ValidateCommit(ctx, commit)

	// Print results
	if result.Passed {
		fmt.Println("Validation passed")
	} else {
		fmt.Println("Validation failed")
	}
}

// customRuleProvider is a simple implementation of the RuleProvider interface.
type customRuleProvider struct {
	rules []domain.Rule
}

func (p *customRuleProvider) GetRules() []domain.Rule {
	return p.rules
}

func (p *customRuleProvider) GetActiveRules() []domain.Rule {
	return p.rules
}
