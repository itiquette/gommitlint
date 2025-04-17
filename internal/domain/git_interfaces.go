// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

// CommitReader provides read access to individual commits.
type CommitReader interface {
	// GetCommit returns a commit by its hash.
	GetCommit(hash string) (*CommitInfo, error)
}

// CommitHistoryReader provides read access to commit history.
type CommitHistoryReader interface {
	// GetCommitRange returns all commits in the given range.
	GetCommitRange(fromHash, toHash string) ([]*CommitInfo, error)

	// GetHeadCommits returns the specified number of commits from HEAD.
	GetHeadCommits(count int) ([]*CommitInfo, error)
}

// RepositoryInfoProvider provides general information about the repository.
type RepositoryInfoProvider interface {
	// GetCurrentBranch returns the name of the current branch.
	GetCurrentBranch() (string, error)

	// GetRepositoryName returns the name of the repository.
	GetRepositoryName() string

	// IsValid checks if the repository is a valid Git repository.
	IsValid() bool
}

// CommitAnalyzer provides analysis functionality for commits.
type CommitAnalyzer interface {
	// GetCommitsAhead returns the number of commits ahead of the given reference.
	GetCommitsAhead(reference string) (int, error)
}

// GitRepositoryService combines all Git repository interfaces.
// It provides a complete interface for Git repository operations.
type GitRepositoryService interface {
	CommitReader
	CommitHistoryReader
	RepositoryInfoProvider
	CommitAnalyzer
}
