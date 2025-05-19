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
  - Context-Based Configuration: Rules retrieve configuration from context
  - Functional Options: Rules use the functional options pattern for initialization
  - Pure Functions: Validation logic is side-effect free
  - Consistent Pattern: All rules follow the same structural pattern

# Common Patterns

All rules share these common patterns:

 1. Simple Structure:
    Each rule has a simple struct with a name field.

 2. Functional Options:
    Rules are configured using functional options pattern when needed.

 3. Configuration Access:
    Rules retrieve configuration from context using contextx.GetConfig(ctx).

 4. Validation Interface:
    All rules implement the Validate method that returns []errors.ValidationError.

# Adding New Rules

To add a new rule:

1. Create a new file named after your rule (e.g., myfeature.go)
2. Define a struct with a name field
3. Implement the Name() and Validate() methods
4. Create a constructor that sets the name
5. Access configuration from context, not embedded fields
6. Write comprehensive tests

Example:

	type MyFeatureRule struct {
		name string
	}

	func NewMyFeatureRule() MyFeatureRule {
		return MyFeatureRule{
			name: "MyFeature",
		}
	}

	func (r MyFeatureRule) Name() string {
		return r.name
	}

	func (r MyFeatureRule) Validate(ctx context.Context, commit domain.CommitInfo) []errors.ValidationError {
		cfg := contextx.GetConfig(ctx)
		enabled := cfg.GetBool("myfeature.enabled")
		if !enabled {
			return nil
		}
		// Validation logic
		return errors
	}

# Configuration

Rules access configuration through context using the standard pattern:

	cfg := contextx.GetConfig(ctx)
	maxLength := cfg.GetInt("subject.max_length")
	enabled := cfg.GetBool("myfeature.enabled")

This ensures consistent configuration access and maintains separation of concerns.

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
