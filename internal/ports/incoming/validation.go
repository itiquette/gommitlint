// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package incoming defines the driver port interfaces for external adapters
package incoming

import (
	"context"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/ports"
)

// ValidationService is the main driver port used by CLI and other adapters.
type ValidationService interface {
	// ValidateCommit validates a single commit by its hash
	ValidateCommit(ctx context.Context, ref string, skipMergeCommits bool) (domain.CommitResult, error)

	// ValidateCommits validates multiple commits by their hashes
	ValidateCommits(ctx context.Context, commitHashes []string, skipMergeCommits bool) (domain.ValidationResults, error)

	// ValidateCommitRange validates commits in a range
	ValidateCommitRange(ctx context.Context, fromHash, toHash string, skipMergeCommits bool) (domain.ValidationResults, error)

	// ValidateMessage validates a commit message directly
	ValidateMessage(ctx context.Context, message string) (domain.ValidationResults, error)
}

// ValidationOrchestrator orchestrates validation and report generation.
// This is the primary interface that CLI adapters should use.
type ValidationOrchestrator interface {
	// ValidateAndReport validates a commit and generates a report.
	// Returns exit code: 0 for success, 2 for validation failure, 1 for errors.
	ValidateAndReport(ctx context.Context, ref string, skipMergeCommits bool, options ports.ReportOptions) (int, error)

	// ValidateRangeAndReport validates a range and generates a report.
	ValidateRangeAndReport(ctx context.Context, fromHash, toHash string, skipMergeCommits bool, options ports.ReportOptions) (int, error)

	// ValidateMessageAndReport validates a message and generates a report.
	ValidateMessageAndReport(ctx context.Context, message string, options ports.ReportOptions) (int, error)
}
