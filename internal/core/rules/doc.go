// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package rules implements the core validation rules for commit messages in gommitlint.

This package follows hexagonal architecture principles, with each rule being an independent
component that implements the domain.Rule interface. Rules are designed to be pure functions
with value semantics, focusing on validating specific aspects of commit messages.

# Architecture

The rules package follows these design principles:

  - Value Semantics: All rules use value receivers and immutable data structures
  - Explicit Dependencies: Rules receive configuration through constructors
  - Functional Options: Rules use the functional options pattern for initialization
  - Pure Functions: Validation logic is side-effect free
  - Consistent Pattern: All rules follow the same structural pattern

# Common Patterns

All rules share these common patterns:

 1. Simple Structure:
    Each rule has a simple struct with a name field and configuration.

 2. Functional Options:
    Rules are configured using functional options pattern for initialization.

 3. Explicit Dependencies:
    Configuration is injected through constructors, not retrieved from context.

 4. Validation Interface:
    All rules implement the Validate method that returns []errors.ValidationError.

# Adding New Rules

To add a new rule:

1. Create a new file named after your rule (e.g., myfeature.go)
2. Define a struct with a name field and configuration fields
3. Create functional option types and functions for configuration
4. Implement the Name() and Validate() methods
5. Create a constructor that accepts functional options
6. Write comprehensive tests

Example:

	type MyFeatureRule struct {
		name      string
		maxItems  int
		checkBody bool
	}

	type MyFeatureOption func(MyFeatureRule) MyFeatureRule

	func WithMaxItems(max int) MyFeatureOption {
		return func(r MyFeatureRule) MyFeatureRule {
			r.maxItems = max
			return r
		}
	}

	func WithCheckBody(check bool) MyFeatureOption {
		return func(r MyFeatureRule) MyFeatureRule {
			r.checkBody = check
			return r
		}
	}

	func NewMyFeatureRule(options ...MyFeatureOption) MyFeatureRule {
		rule := MyFeatureRule{
			name:      "MyFeature",
			maxItems:  10,        // default
			checkBody: true,      // default
		}

		for _, option := range options {
			rule = option(rule)
		}

		return rule
	}

	func (r MyFeatureRule) Name() string {
		return r.name
	}

	func (r MyFeatureRule) Validate(ctx context.Context, commit domain.CommitInfo) []errors.ValidationError {
		// Use injected configuration, not context
		if r.maxItems <= 0 {
			return nil
		}
		// Validation logic using r.maxItems and r.checkBody
		return errors
	}

# Configuration

Rules receive configuration through their constructors via functional options.
The factory (SimpleRuleFactory) is responsible for reading configuration and
passing it to rules during initialization:

	// In the factory
	func (f SimpleRuleFactory) createSubjectLengthRule() domain.Rule {
		options := []rules.SubjectLengthOption{}

		if f.config != nil {
			maxLength := f.config.Message.Subject.MaxLength
			if maxLength > 0 {
				options = append(options, rules.WithMaxLength(maxLength))
			}
		}

		return rules.NewSubjectLengthRule(options...)
	}

This ensures explicit dependencies and maintains separation of concerns.

# Available Rules

The package provides these validation rules:

  - CommitBodyRule: Validates commit body requirements
  - BranchAheadRule: Validates the number of commits ahead of reference branch
  - ConventionalCommitRule: Validates Conventional Commits format
  - ImperativeVerbRule: Validates imperative mood in subject
  - JiraReferenceRule: Validates JIRA ticket references
  - SignatureRule: Validates commit signatures (GPG/SSH)
  - IdentityRule: Validates commit signatures match committer identity
  - SignOffRule: Validates Developer Certificate of Origin
  - SpellRule: Validates spelling in commit messages
  - SubjectCaseRule: Validates subject line casing
  - SubjectLengthRule: Validates subject line length
  - SubjectSuffixRule: Validates subject line doesn't end with certain suffixes

Each rule focuses on a specific aspect of commit message validation and can be
independently enabled, disabled, and configured.
*/
package rules
