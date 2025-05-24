// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"

	"github.com/itiquette/gommitlint/internal/errors"
)

// Rule defines the interface for all validation rules.
// Rules are pure validators that check commits against specific criteria.
type Rule interface {
	// Name returns the rule's name.
	Name() string

	// Validate performs validation against a commit.
	// This should be a pure function that returns errors without side effects.
	Validate(ctx context.Context, commit CommitInfo) []errors.ValidationError
}
