// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
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
func (v *Validator) Validate() (*model.CommitRules, error) {
	msgs, err := v.getCommitInfos()
	if err != nil {
		return nil, err
	}

	commitRules := &model.CommitRules{}

	for _, msg := range msgs {
		v.checkValidity(commitRules, msg)
	}

	return commitRules, nil
}
