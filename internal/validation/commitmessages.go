// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package validation

import (
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal/model"
)

// getCommitMessages retrieves commit messages based on options.
func (v *Validator) getCommitMessages() ([]model.CommitInfo, error) {
	switch {
	case v.options.CommitMsgFile != nil:
		return v.getCommitFromFile()
	case v.options.RevisionRange != "":
		return v.getCommitFromRange()
	default:
		return v.getCurrentCommit()
	}
}

func (v *Validator) getCommitFromFile() ([]model.CommitInfo, error) {
	contents, err := os.ReadFile(*v.options.CommitMsgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit message file: %w", err)
	}

	return []model.CommitInfo{{Message: string(contents)}}, nil
}

func (v *Validator) getCommitFromRange() ([]model.CommitInfo, error) {
	revs, err := v.parseRevisionRange()
	if err != nil {
		return nil, err
	}

	msgs, err := v.git.Messages(revs[0], revs[1])
	if err != nil {
		return nil, fmt.Errorf("failed to get commit messages: %w", err)
	}

	return msgs, nil
}

func (v *Validator) getCurrentCommit() ([]model.CommitInfo, error) {
	msg, err := v.git.Messages("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit message: %w", err)
	}

	return []model.CommitInfo{msg[0]}, nil
}
