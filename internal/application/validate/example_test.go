// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validate_test

import (
	"fmt"
	"regexp"

	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
)

// This example demonstrates how to create and register a custom rule.
func Example_customRule() {
	// A custom rule that checks if commit messages contain a reference to a JIRA ticket
	customRule := &MyCustomRule{
		name:            "MyJiraRule",
		jiraProjectKeys: []string{"PROJ", "ENG"},
	}

	// Assume service is created elsewhere
	var service validate.ValidationService // This would be properly initialized in real code

	// Register the rule with the validation service
	_ = service.RegisterCustomRule(customRule)

	// Now when validating commits, the custom rule will be included
	fmt.Println("Custom rule registered successfully")

	// Output: Custom rule registered successfully
}

// This example demonstrates how to create a rule factory for conditional rule creation.
func Example_customRuleFactory() {
	// Assume service is created elsewhere
	var service validate.ValidationService // This would be properly initialized in real code

	// Register a factory for creating the rule conditionally
	_ = service.RegisterCustomRuleFactory(
		"ConditionalJiraRule",
		// Factory function to create the rule
		func(config validate.ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return &MyCustomRule{
				name:            "ConditionalJiraRule",
				jiraProjectKeys: config.JiraProjects(), // Use config from validation service
			}
		},
		false, // Doesn't require commit analyzer
		// Condition function to determine when to create the rule
		func(config validate.ValidationConfig) bool {
			// Only create this rule if Jira validation is enabled
			return config.JiraRequired()
		},
	)

	fmt.Println("Custom rule factory registered successfully")

	// Output: Custom rule factory registered successfully
}

// MyCustomRule is an example of a custom validation rule.
type MyCustomRule struct {
	name            string
	jiraProjectKeys []string
	violations      []errors.ValidationError
}

// Name returns the rule's name.
func (r *MyCustomRule) Name() string {
	return r.name
}

// Validate checks a commit message for a JIRA reference.
func (r *MyCustomRule) Validate(commit domain.CommitInfo) []errors.ValidationError {
	r.violations = nil

	// Reset violations
	hasJiraReference := false

	// Check for any of the configured JIRA keys
	for _, key := range r.jiraProjectKeys {
		if containsJiraReference(commit.Message, key) {
			hasJiraReference = true

			break
		}
	}

	// Record a violation if no JIRA reference was found
	if !hasJiraReference {
		r.violations = append(r.violations, errors.ValidationError{
			Code:    "missing_jira_reference",
			Message: "Commit message must reference a JIRA ticket",
		})
	}

	return r.violations
}

// Result returns a concise result message.
func (r *MyCustomRule) Result() string {
	if len(r.violations) > 0 {
		return "Missing JIRA reference"
	}

	return "JIRA reference found"
}

// VerboseResult returns a detailed result message.
func (r *MyCustomRule) VerboseResult() string {
	if len(r.violations) > 0 {
		projects := ""

		for i, key := range r.jiraProjectKeys {
			if i > 0 {
				projects += ", "
			}

			projects += key
		}

		return "No JIRA reference found. Expected one of: " + projects
	}

	return "Found a valid JIRA reference"
}

// Help returns guidance for fixing violations.
func (r *MyCustomRule) Help() string {
	return "Include a JIRA ticket reference (e.g., PROJ-123) in your commit message"
}

// Errors returns all validation errors.
func (r *MyCustomRule) Errors() []errors.ValidationError {
	return r.violations
}

// Example: "PROJ-123" or "[PROJ-123]".
func containsJiraReference(s, projectKey string) bool {
	pattern := regexp.QuoteMeta(projectKey) + "-\\d+"
	re := regexp.MustCompile(pattern)

	return re.MatchString(s)
}
