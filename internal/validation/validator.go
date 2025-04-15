// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package validation

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal"
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
		return nil, internal.NewGitError(fmt.Errorf("failed to open git repo: %w", err))
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
	return v.ValidateWithContext(context.Background())
}

// ValidateWithContext performs commit message validation with context propagation.
func (v *Validator) ValidateWithContext(ctx context.Context) (*model.CommitRules, error) {
	// Check for early cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	commits, err := v.GetCommitsToValidateWithContext(ctx)
	if err != nil {
		return nil, err
	}

	// Create overall rules container
	commitRules := model.NewCommitRules()

	// Validate all commits and add results to the same rule container
	for _, commit := range commits {
		// Check for cancellation between iterations
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		rules, err := v.ValidateCommitWithContext(ctx, commit)
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
// This is kept for backward compatibility.
func (v *Validator) GetCommitsToValidate() ([]model.CommitInfo, error) {
	return v.GetCommitsToValidateWithContext(context.Background())
}

// GetCommitsToValidateWithContext retrieves all commit infos with context support.
func (v *Validator) GetCommitsToValidateWithContext(ctx context.Context) ([]model.CommitInfo, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return v.getCommitInfos(ctx)
}

// ValidateCommit validates a single commit and returns its rules.
// This is kept for backward compatibility.
func (v *Validator) ValidateCommit(commitInfo model.CommitInfo) (*model.CommitRules, error) {
	return v.ValidateCommitWithContext(context.Background(), commitInfo)
}

// ValidateCommitWithContext validates a single commit with context support.
func (v *Validator) ValidateCommitWithContext(ctx context.Context, commitInfo model.CommitInfo) (*model.CommitRules, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Create a new rule container for this commit
	commitRules := model.NewCommitRules()

	// Check validity of this specific commit
	v.checkValidity(ctx, commitRules, commitInfo)

	return commitRules, nil
}
