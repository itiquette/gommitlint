// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package validation

import (
	"context"
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal"
	"github.com/itiquette/gommitlint/internal/model"
)

// getCommitInfos retrieves commit messages based on options.
func (v *Validator) getCommitInfos(ctx context.Context) ([]model.CommitInfo, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	switch {
	case v.options.MsgFromFile != nil:
		return v.getCommitInfosFromFile(ctx)
	case v.options.RevisionRange != "":
		return v.getCommitInfosFromRange(ctx)
	default:
		return v.getCurrentCommit(ctx)
	}
}

func (v *Validator) getCommitInfosFromFile(ctx context.Context) ([]model.CommitInfo, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	contents, err := os.ReadFile(*v.options.MsgFromFile)
	if err != nil {
		return nil, internal.NewInputError(
			fmt.Errorf("failed to read commit message file: %w", err),
			map[string]string{"file": *v.options.MsgFromFile})
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

func (v *Validator) getCommitInfosFromRange(ctx context.Context) ([]model.CommitInfo, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	revs, err := v.parseRevisionRange()
	if err != nil {
		return nil, err
	}

	msgs, err := v.repo.CommitInfos(revs[0], revs[1])
	if err != nil {
		return nil, internal.NewGitError(
			fmt.Errorf("failed to get commit messages: %w", err),
			map[string]string{
				"from_revision": revs[0],
				"to_revision":   revs[1],
			})
	}

	return msgs, nil
}

func (v *Validator) getCurrentCommit(ctx context.Context) ([]model.CommitInfo, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	msg, err := v.repo.CommitInfos("", "")
	if err != nil {
		return nil, internal.NewGitError(
			fmt.Errorf("failed to get current commit message: %w", err))
	}

	return []model.CommitInfo{msg[0]}, nil
}
