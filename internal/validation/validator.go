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
	git     *model.Repository
	options *model.Options
	config  *configuration.GommitLintConfig
}

// NewValidator creates a new Validator instance.
func NewValidator(options *model.Options, config *configuration.GommitLintConfig) (*Validator, error) {
	gitHandler, err := model.NewRepository("")
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
		v.checkValidity(report, msg)
	}

	return report, nil
}
