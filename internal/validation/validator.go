// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package validation

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/itiquette/gommitlint/internal/model"
)

// Validator handles commit message validation logic.
type Validator struct {
	repo    *model.Repository
	options *model.Options
	config  *configuration.GommitLintConfig
}

// NewValidator creates a new Validator instance.
func NewValidator(options *model.Options, config *configuration.GommitLintConfig) (*Validator, error) {
	repo, err := model.NewRepository("")
	if err != nil {
		return nil, fmt.Errorf("failed to open git repo: %w", err)
	}

	return &Validator{
		repo:    repo,
		options: options,
		config:  config,
	}, nil
}

// Validate performs commit message validation based on configured rules.
// This is kept for backward compatibility.
func (v *Validator) Validate() (*model.CommitRules, error) {
	commits, err := v.GetCommitsToValidate()
	if err != nil {
		return nil, err
	}

	// Create overall rules container
	commitRules := model.NewCommitRules()

	// Validate all commits and add results to the same rule container
	for _, commit := range commits {
		rules, err := v.ValidateCommit(commit)
		if err != nil {
			return nil, err
		}

		// Add all rules from this commit to the overall container
		for _, rule := range rules.All() {
			commitRules.Add(rule)
		}
	}

	return commitRules, nil
}

// GetCommitsToValidate retrieves all commit infos that need validation based on options.
func (v *Validator) GetCommitsToValidate() ([]model.CommitInfo, error) {
	return v.getCommitInfos()
}

// ValidateCommit validates a single commit and returns its rules.
func (v *Validator) ValidateCommit(commitInfo model.CommitInfo) (*model.CommitRules, error) {
	// Create a new rule container for this commit
	commitRules := model.NewCommitRules()

	// Check validity of this specific commit
	v.checkValidity(commitRules, commitInfo)

	return commitRules, nil
}
