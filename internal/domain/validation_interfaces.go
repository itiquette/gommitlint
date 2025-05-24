// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import "context"

// ValidationEngine defines the interface for commit validation engines.
// This interface is defined in the domain layer following hexagonal architecture
// principles where interfaces are defined at consumption sites.
type ValidationEngine interface {
	// ValidateCommit validates a single commit against active rules.
	ValidateCommit(ctx context.Context, commit CommitInfo) CommitResult

	// ValidateCommits validates multiple commits against active rules.
	ValidateCommits(ctx context.Context, commits []CommitInfo) ValidationResults
}
