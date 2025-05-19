// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package incoming defines the driver port interfaces for external adapters
package incoming

import (
	"context"

	"github.com/itiquette/gommitlint/internal/domain"
)

// ValidationService is the main driver port used by CLI and other adapters.
type ValidationService interface {
	// ValidateCommit validates a single commit by its hash
	ValidateCommit(ctx context.Context, ref string) (domain.CommitResult, error)

	// ValidateCommits validates multiple commits by their hashes
	ValidateCommits(ctx context.Context, commitHashes []string) (domain.ValidationResults, error)

	// ValidateCommitRange validates commits in a range
	ValidateCommitRange(ctx context.Context, fromHash, toHash string, skipMergeCommits bool) (domain.ValidationResults, error)

	// ValidateMessage validates a commit message directly
	ValidateMessage(ctx context.Context, message string) (domain.ValidationResults, error)

	// ValidateMessageFile validates a commit message from a file
	ValidateMessageFile(ctx context.Context, filepath string) (domain.ValidationResults, error)
}
