// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package git provides Git repository adapters for the domain model.
package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/itiquette/gommitlint/internal/domain"
)

// RepositoryAdapter adapts a git repository to the domain model.
type RepositoryAdapter struct {
	repo *git.Repository
	path string
}

// NewRepositoryAdapter creates a new RepositoryAdapter for the given path.
func NewRepositoryAdapter(path string) (*RepositoryAdapter, error) {
	// If path is empty, use current directory
	if path == "" {
		var err error

		path, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Find the git repository
	gitDir, err := findGitDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to find git directory: %w", err)
	}

	// Open the git repository
	repo, err := git.PlainOpen(gitDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &RepositoryAdapter{
		repo: repo,
		path: gitDir,
	}, nil
}

// GetCommit returns a commit by its hash.
func (g RepositoryAdapter) GetCommit(ctx context.Context, hash string) (domain.CommitInfo, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return domain.CommitInfo{}, ctx.Err()
	}

	// Resolve the hash (handles empty hash for HEAD)
	resolvedHash, err := g.resolveRevision(hash)
	if err != nil {
		return domain.CommitInfo{}, err
	}

	// Get the commit
	commit, err := g.getCommitByHash(resolvedHash)
	if err != nil {
		return domain.CommitInfo{}, err
	}

	// Convert to domain commit
	return g.convertCommit(commit)
}

// GetCommitRange returns all commits in the given range.
func (g RepositoryAdapter) GetCommitRange(ctx context.Context, fromHash, toHash string) ([]domain.CommitInfo, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Resolve the "to" hash
	toRevision, err := g.resolveRevision(toHash)
	if err != nil {
		return nil, err
	}

	// Resolve the "from" hash
	from, err := g.resolveRevision(fromHash)
	if err != nil {
		return nil, err
	}

	// Create iterator
	iter, err := g.createCommitIterator(toRevision)
	if err != nil {
		return nil, err
	}

	// Check for context cancellation before proceeding with potentially lengthy operation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Collect and convert commits until we reach the "from" commit
	domainCommits, err := g.collectAndConvertCommits(iter, 0, func(commit *object.Commit) bool {
		return commit.Hash == from
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Create a collection for easier manipulation
	collection := domain.NewCommitCollection(domainCommits)

	// If "from" commit is not already included, add it
	if !collection.Contains(from.String()) {
		// Check for context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Get the "from" commit
		fromCommit, err := g.getCommitByHash(from)
		if err != nil {
			return nil, err
		}

		// Convert and add to the collection
		domainFromCommit, err := g.convertCommit(fromCommit)
		if err != nil {
			return nil, err
		}

		collection.Add(domainFromCommit)
	}

	return collection.All(), nil
}

// GetHeadCommits returns the specified number of commits from HEAD.
func (g *RepositoryAdapter) GetHeadCommits(ctx context.Context, count int) ([]domain.CommitInfo, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Resolve HEAD hash
	headHash, err := g.resolveRevision("")
	if err != nil {
		return nil, err
	}

	// Create iterator
	iter, err := g.createCommitIterator(headHash)
	if err != nil {
		return nil, err
	}

	// Check for context cancellation before proceeding with potentially lengthy operation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Collect and convert commits with limit
	return g.collectAndConvertCommits(iter, count, nil)
}

// GetCurrentBranch returns the name of the current branch.
func (g RepositoryAdapter) GetCurrentBranch(ctx context.Context) (string, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// Get the HEAD reference
	ref, err := g.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	// Check if we're on a branch
	if ref.Name().IsBranch() {
		return ref.Name().Short(), nil
	}

	// We're in detached HEAD state, try to find the branch that contains HEAD
	headHash := ref.Hash()

	// Check for context cancellation before proceeding with potentially lengthy operation
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// Get all branches
	branches, err := g.repo.Branches()
	if err != nil {
		return "", fmt.Errorf("failed to get branches: %w", err)
	}

	// Check if any branch points to HEAD
	var branchName string

	err = branches.ForEach(func(branch *plumbing.Reference) error {
		// Check for context cancellation during iteration
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if branch.Hash() == headHash {
			branchName = branch.Name().Short()

			return errors.New("stop") // Use error to stop iteration
		}

		return nil
	})

	// ForEach returns a "stop" error when we've found the branch, or ctx.Err() when cancelled
	if ctx.Err() != nil {
		return "", ctx.Err()
	} else if err != nil && err.Error() != "stop" {
		return "", fmt.Errorf("failed to iterate branches: %w", err)
	}

	if branchName != "" {
		return branchName, nil
	}

	// We're in a detached HEAD state
	return "HEAD detached at " + headHash.String()[:7], nil
}

// GetRepositoryName returns the name of the repository.
func (g RepositoryAdapter) GetRepositoryName(_ context.Context) string {
	// No need to check for context cancellation for this simple operation
	// Extract the repository name from the path
	return filepath.Base(g.path)
}

// findGitDir is moved to repository_helpers.go

// convertCommit converts a go-git commit to a domain commit.
func (g RepositoryAdapter) convertCommit(commit *object.Commit) (domain.CommitInfo, error) {
	// Split the commit message into subject and body
	message := commit.Message
	subject, body := domain.SplitCommitMessage(message)

	// Check if the commit is a merge commit
	isMergeCommit := len(commit.ParentHashes) > 1

	// Create domain commit
	domainCommit := domain.CommitInfo{
		Hash:          commit.Hash.String(),
		Subject:       subject,
		Body:          body,
		Message:       message,
		RawCommit:     commit,
		IsMergeCommit: isMergeCommit,
	}

	// Get signature if available
	if commit.PGPSignature != "" {
		domainCommit.Signature = commit.PGPSignature
	}

	return domainCommit, nil
}

// IsValid checks if the repository is a valid Git repository.
func (g *RepositoryAdapter) IsValid(_ context.Context) bool {
	// No need to check for context cancellation for this simple operation
	// We were able to open the repository, so it's valid
	return g.repo != nil
}

// GetCommitsAhead returns the number of commits ahead of the given reference.
func (g *RepositoryAdapter) GetCommitsAhead(ctx context.Context, reference string) (int, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	// Resolve HEAD
	head, err := g.resolveRevision("")
	if err != nil {
		return 0, err
	}

	// Resolve reference
	ref, err := g.resolveRevision(reference)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve reference %s: %w", reference, err)
	}

	// Check for context cancellation
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	// Find merge base
	mergeBase, err := g.findMergeBase(head, ref)
	if err != nil {
		return 0, fmt.Errorf("failed to find merge base: %w", err)
	}

	// Create iterator from HEAD
	iter, err := g.createCommitIterator(head)
	if err != nil {
		return 0, err
	}

	// Check for context cancellation before proceeding with potentially lengthy operation
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	// Collect commits between HEAD and merge base
	commits, err := g.collectCommits(iter, 0, func(commit *object.Commit) bool {
		// Check for context cancellation during iteration (this check has minimal performance impact)
		if ctx.Err() != nil {
			return true // Break the iteration
		}

		return commit.Hash == mergeBase
	})

	// Handle context cancellation that might have happened during commit collection
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	if err != nil {
		return 0, fmt.Errorf("failed to count commits ahead: %w", err)
	}

	// Return the count of commits
	return len(commits), nil
}

// findMergeBase finds the common ancestor of two commits.
func (g *RepositoryAdapter) findMergeBase(hash1, hash2 plumbing.Hash) (plumbing.Hash, error) {
	return findMergeBase(g.repo, hash1, hash2)
}

// resolveRevision resolves a revision to a hash.
// If the revision is empty, HEAD is used.
func (g *RepositoryAdapter) resolveRevision(revision string) (plumbing.Hash, error) {
	return resolveRevision(g.repo, revision)
}

// getCommitByHash gets a commit by its hash.
func (g *RepositoryAdapter) getCommitByHash(hash plumbing.Hash) (*object.Commit, error) {
	return getCommitByHash(g.repo, hash)
}

// createCommitIterator creates a commit iterator starting from the given hash.
func (g *RepositoryAdapter) createCommitIterator(hash plumbing.Hash) (object.CommitIter, error) {
	return createCommitIterator(g.repo, hash)
}

// collectCommits collects commits from an iterator, with optional limit and stop condition.
func (g *RepositoryAdapter) collectCommits(
	iter object.CommitIter,
	limit int,
	stopFn func(*object.Commit) bool,
) ([]*object.Commit, error) {
	return collectCommits(iter, limit, stopFn)
}

// collectAndConvertCommits collects commits from an iterator, converts them to domain commits.
func (g *RepositoryAdapter) collectAndConvertCommits(
	iter object.CommitIter,
	limit int,
	stopFn func(*object.Commit) bool,
) ([]domain.CommitInfo, error) {
	// Collect git commits
	commits, err := g.collectCommits(iter, limit, stopFn)
	if err != nil {
		return nil, err
	}

	// Convert to domain commits
	domainCommits := make([]domain.CommitInfo, 0, len(commits))

	for _, commit := range commits {
		domainCommit, err := g.convertCommit(commit)
		if err != nil {
			return nil, err
		}

		domainCommits = append(domainCommits, domainCommit)
	}

	return domainCommits, nil
}
