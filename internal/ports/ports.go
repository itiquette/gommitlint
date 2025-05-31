// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package ports defines all adapter interfaces following hexagonal architecture.
// This centralizes port definitions at domain boundaries, eliminating scattered interface duplication.
package ports

import (
	"context"

	"github.com/itiquette/gommitlint/internal/domain"
)

// Logger provides structured logging capabilities.
// Used by all adapters - single definition eliminates duplication across cli, git, output packages.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// ValidationService provides commit validation operations.
// Used by CLI adapter for core validation functionality.
type ValidationService interface {
	// ValidateCommit validates a single commit by its reference
	ValidateCommit(ctx context.Context, ref string, skipMergeCommits bool) (domain.CommitResult, error)

	// ValidateCommits validates multiple commits by their hashes
	ValidateCommits(ctx context.Context, commitHashes []string, skipMergeCommits bool) (domain.ValidationResults, error)

	// ValidateCommitRange validates commits in a range
	ValidateCommitRange(ctx context.Context, fromHash, toHash string, skipMergeCommits bool) (domain.ValidationResults, error)

	// ValidateMessage validates a commit message directly
	ValidateMessage(ctx context.Context, message string) (domain.ValidationResults, error)
}

// Note: Domain interfaces (Repository, CryptoVerifier, etc.) remain in domain package
// to maintain proper hexagonal boundaries. Only adapter-facing interfaces are defined here.
