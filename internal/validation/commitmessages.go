// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal/git"
)

// getCommitMessages retrieves commit messages based on options.
func (v *Validator) getCommitMessages() ([]git.CommitInfo, error) {
	switch {
	case v.options.CommitMsgFile != nil:
		return v.getCommitFromFile()
	case v.options.RevisionRange != "":
		return v.getCommitFromRange()
	default:
		return v.getCurrentCommit()
	}
}

func (v *Validator) getCommitFromFile() ([]git.CommitInfo, error) {
	contents, err := os.ReadFile(*v.options.CommitMsgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit message file: %w", err)
	}

	return []git.CommitInfo{{Message: string(contents)}}, nil
}

func (v *Validator) getCommitFromRange() ([]git.CommitInfo, error) {
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

func (v *Validator) getCurrentCommit() ([]git.CommitInfo, error) {
	msg, err := v.git.Message()
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit message: %w", err)
	}

	return []git.CommitInfo{msg}, nil
}
