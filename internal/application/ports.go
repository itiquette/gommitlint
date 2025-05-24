// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package application defines the application layer's ports (interfaces).
// In hexagonal architecture, the application layer defines what it needs,
// and adapters implement these interfaces.
package application

import (
	"context"
	"io"

	"github.com/itiquette/gommitlint/internal/domain"
)

// ValidationService is the port that incoming adapters (like CLI) must implement.
// This interface is defined by the application layer to express what it needs.
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

// ReportService is the port for report generation.
// This interface defines what the application needs for reporting.
type ReportService interface {
	// GenerateReport creates a report from validation results.
	GenerateReport(ctx context.Context, results domain.ValidationResults, writer io.Writer) error
}

// Logger is the port for logging operations.
// This interface defines what the application needs for logging.
type Logger interface {
	Debug(msg string, kvs ...interface{})
	Info(msg string, kvs ...interface{})
	Warn(msg string, kvs ...interface{})
	Error(msg string, kvs ...interface{})
}

// ResultFormatter is the port for formatting validation results.
// This interface defines what the application needs for formatting.
type ResultFormatter interface {
	// Format converts validation results to a formatted string.
	Format(ctx context.Context, results interface{}) string
}
