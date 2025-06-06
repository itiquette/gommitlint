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
  - Direct Configuration: Rules receive configuration directly through constructors
  - Pure Functions: Validation logic is side-effect free
  - Consistent Pattern: All rules follow the same structural pattern

# Common Patterns

All rules share these common patterns:

 1. Simple Structure:
    Each rule has a simple struct with a name field and configuration.

 2. Direct Configuration:
    Rules receive configuration directly through their constructors.

 3. Explicit Dependencies:
    Configuration is passed as a parameter, not retrieved from context.

 4. Validation Interface:
    All rules implement the Validate method that returns []domain.RuleFailure.

# Adding New Rules

To add a new rule:

1. Create a new file named after your rule (e.g., myfeature.go)
2. Define a struct with a name field and configuration fields
3. Implement the Name() and Validate() methods
4. Create a constructor that accepts configuration directly
6. Write comprehensive tests

Example:

	type MyFeatureRule struct {
		name      string
	}

	func NewMyFeatureRule(cfg config.Config) MyFeatureRule {
		return MyFeatureRule{
			name:      "MyFeature",
			maxItems:  cfg.MyFeature.MaxItems,
			checkBody: cfg.MyFeature.CheckBody,
		}
	}

	func (r MyFeatureRule) Name() string {
		return r.name
	}

	func (r MyFeatureRule) Validate(commit domain.Commit, repo domain.Repository, cfg config.Config) []domain.RuleFailure {
		// Use injected configuration from constructor
		if r.maxItems <= 0 {
			return nil
		}
		// Validation logic using r.maxItems and r.checkBody
		return failures
	}

# Configuration

Rules receive configuration directly through their constructors.
The rule factory is responsible for creating rules with appropriate configuration:

	// In domain/rules/factory.go
	func createSubjectRule(cfg config.Config) domain.Rule {
		return NewSubjectRule(cfg)
	}

This ensures explicit dependencies and maintains separation of concerns.

# Available Rules

The package provides these validation rules:

  - CommitBodyRule: Validates commit body requirements
  - BranchAheadRule: Validates the number of commits ahead of reference branch
  - ConventionalCommitRule: Validates Conventional Commits format
  - JiraReferenceRule: Validates JIRA ticket references
  - SignatureRule: Validates commit signatures (GPG/SSH)
  - IdentityRule: Validates commit signatures match committer identity
  - SignOffRule: Validates Developer Certificate of Origin
  - SpellRule: Validates spelling in commit messages
  - SubjectRule: Validates subject length, case, suffix, and imperative mood (consolidated rule)

Each rule focuses on a specific aspect of commit message validation and can be
independently enabled, disabled, and configured.
*/
package rules
