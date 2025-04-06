// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package validation

import (
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal/model"
)

// getCommitInfos retrieves commit messages based on options.
func (v *Validator) getCommitInfos() ([]model.CommitInfo, error) {
	switch {
	case v.options.MsgFromFile != nil:
		return v.getCommitInfosFromFile()
	case v.options.RevisionRange != "":
		return v.getCommitInfosFromRange()
	default:
		return v.getCurrentCommit()
	}
}

func (v *Validator) getCommitInfosFromFile() ([]model.CommitInfo, error) {
	contents, err := os.ReadFile(*v.options.MsgFromFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit message file: %w", err)
	}

	subject, body := model.SplitCommitMessage(string(contents))
	commits := []model.CommitInfo{{
		Message:   string(contents),
		Subject:   subject,
		Body:      body,
		Signature: "",
		RawCommit: nil,
	}}

	return commits, nil
}

func (v *Validator) getCommitInfosFromRange() ([]model.CommitInfo, error) {
	revs, err := v.parseRevisionRange()
	if err != nil {
		return nil, err
	}

	msgs, err := v.repo.CommitInfos(revs[0], revs[1])
	if err != nil {
		return nil, fmt.Errorf("failed to get commit messages: %w", err)
	}

	return msgs, nil
}

func (v *Validator) getCurrentCommit() ([]model.CommitInfo, error) {
	msg, err := v.repo.CommitInfos("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit message: %w", err)
	}

	return []model.CommitInfo{msg[0]}, nil
}
