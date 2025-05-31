// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import "context"

// Repository provides all Git repository operations in a single interface.
// This consolidates commit access, repository info, and analysis capabilities.
type Repository interface {
	// Commit operations
	GetCommit(ctx context.Context, hash string) (CommitInfo, error)
	GetCommits(ctx context.Context, limit int) ([]CommitInfo, error)
	GetCommitRange(ctx context.Context, fromHash, toHash string) ([]CommitInfo, error)
	GetHeadCommits(ctx context.Context, count int) ([]CommitInfo, error)

	// Repository information
	GetCurrentBranch(ctx context.Context) (string, error)
	GetRepositoryName(ctx context.Context) string
	IsValid(ctx context.Context) (bool, error)

	// Analysis operations
	GetCommitsAhead(ctx context.Context, ref string) (int, error)
}
