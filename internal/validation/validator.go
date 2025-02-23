// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/itiquette/gommitlint/internal/git"
	"github.com/itiquette/gommitlint/internal/model"
)

// Validator handles commit message validation logic.
type Validator struct {
	git     *git.Git
	options *model.Options
	config  *configuration.Gommit
}

// NewValidator creates a new Validator instance.
func NewValidator(options *model.Options, config *configuration.Gommit) (*Validator, error) {
	gitHandler, err := git.NewGit()
	if err != nil {
		return nil, fmt.Errorf("failed to open git repo: %w", err)
	}

	return &Validator{
		git:     gitHandler,
		options: options,
		config:  config,
	}, nil
}

// Validate performs commit message validation based on configured rules.
func (v *Validator) Validate() (*model.Report, error) {
	msgs, err := v.getCommitMessages()
	if err != nil {
		return nil, err
	}

	report := &model.Report{}

	for _, msg := range msgs {
		v.config.Message = msg.Message
		v.checkValidity(report)
	}

	return report, nil
}
