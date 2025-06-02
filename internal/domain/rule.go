// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import "github.com/itiquette/gommitlint/internal/domain/config"

// ValidationContext provides all data needed for rule validation.
// This is an immutable value passed to rules containing the commit,
// optional repository access, and configuration.
type ValidationContext struct {
	Commit     Commit         // Always present - the commit to validate
	Repository Repository     // Optional - for rules that need repository access
	Config     *config.Config // Rule configuration
}

// Rule defines the interface for all validation rules.
// Rules are pure validators that check against provided context.
// A single interface is used for all rules - repository-aware rules
// simply check if Repository is available in the context.
type Rule interface {
	// Name returns the rule's name.
	Name() string

	// Validate performs validation using the provided context.
	// This is a pure function that returns failures without side effects.
	// Rules that need repository access check if ctx.Repository is non-nil.
	Validate(ctx ValidationContext) []RuleFailure
}

// RepositoryRule identifies rules that operate at repository level.
// Repository-level rules validate repository state, not individual commits.
type RepositoryRule interface {
	Rule
	// IsRepositoryLevel returns true if this rule operates at repository level.
	IsRepositoryLevel() bool
}

// IsRepositoryLevelRule checks if a rule operates at repository level.
func IsRepositoryLevelRule(rule Rule) bool {
	if repoRule, ok := rule.(RepositoryRule); ok {
		return repoRule.IsRepositoryLevel()
	}

	return false
}

// SeparateRules separates rules into commit-level and repository-level rules.
func SeparateRules(rules []Rule) ([]Rule, []Rule) {
	var commitRules []Rule

	var repositoryRules []Rule

	for _, rule := range rules {
		if IsRepositoryLevelRule(rule) {
			repositoryRules = append(repositoryRules, rule)
		} else {
			commitRules = append(commitRules, rule)
		}
	}

	return commitRules, repositoryRules
}
