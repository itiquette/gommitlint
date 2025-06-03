// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import "github.com/itiquette/gommitlint/internal/domain/config"

// Rule defines the interface for all validation rules.
// Rules are pure validators that receive all dependencies and use only what they need.
type Rule interface {
	// Name returns the rule's name.
	Name() string

	// Validate performs validation with explicit dependencies.
	// Rules check for nil repository/config if they need them.
	Validate(commit Commit, repo Repository, config *config.Config) []RuleFailure
}

// IsRepositoryLevelRule checks if a rule operates at repository level.
// Currently only BranchAhead is a repository-level rule.
func IsRepositoryLevelRule(rule Rule) bool {
	return rule.Name() == "BranchAhead"
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
