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
	"github.com/itiquette/gommitlint/internal/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// RepositoryAdapter adapts a git repository to the domain model.
// This version uses value semantics throughout.
type RepositoryAdapter struct {
	repo *git.Repository // Needs to remain a pointer as go-git requires this
	path string
}

// NewRepositoryAdapter creates a new RepositoryAdapter for the given path.
// This version returns a value rather than a pointer.
func NewRepositoryAdapter(ctx context.Context, path string) (RepositoryAdapter, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("path", path).Msg("Entering NewRepositoryAdapter")
	// If path is empty, use current directory
	if path == "" {
		var err error

		path, err = os.Getwd()
		if err != nil {
			return RepositoryAdapter{}, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Find the git repository
	gitDir, err := findGitDir(ctx, path)
	if err != nil {
		return RepositoryAdapter{}, fmt.Errorf("failed to find git directory: %w", err)
	}

	// Open the git repository
	repo, err := git.PlainOpen(gitDir)
	if err != nil {
		return RepositoryAdapter{}, fmt.Errorf("failed to open git repository: %w", err)
	}

	return RepositoryAdapter{
		repo: repo,
		path: gitDir,
	}, nil
}

// GetCommit returns a commit by its hash.
func (g RepositoryAdapter) GetCommit(ctx context.Context, hash string) (domain.CommitInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("hash", hash).Str("repository_path", g.path).Msg("Entering GetCommit")
	// Instead of repeating the logic for resolving and getting commits,
	// use our helper function to get exactly 1 commit
	commits, err := g.getCommitsFromHash(ctx, hash, 1)
	if err != nil {
		// Create a rich error context
		errCtx := appErrors.NewContext()

		// Create an enhanced error
		richErr := appErrors.CreateRichError(
			"GitRepository",
			appErrors.ErrGitOperationFailed,
			fmt.Sprintf("Failed to get commit with hash %s: %s", hash, err),
			"Check that the commit reference is valid and the repository is accessible.",
			errCtx,
		)

		// Add extra context
		richErr = richErr.WithContext("git_reference", hash)
		richErr = richErr.WithContext("repository_path", g.path)

		return domain.CommitInfo{}, richErr
	}

	// Ensure we got exactly one commit
	if len(commits) == 0 {
		// Create a rich error context
		errCtx := appErrors.NewContext()

		// Create an enhanced error
		richErr := appErrors.CreateRichError(
			"GitRepository",
			appErrors.ErrCommitNotFound,
			"Commit not found: "+hash,
			"Verify the commit reference is correct and exists in the repository.",
			errCtx,
		)

		// Add extra context
		richErr = richErr.WithContext("git_reference", hash)
		richErr = richErr.WithContext("repository_path", g.path)

		return domain.CommitInfo{}, richErr
	}

	// Return the first (and only) commit
	return commits[0], nil
}

// GetCommitRange returns all commits in the given range.
// This version uses functional patterns to avoid state mutation.
func (g RepositoryAdapter) GetCommitRange(ctx context.Context, fromHash, toHash string) ([]domain.CommitInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("from_hash", fromHash).Str("to_hash", toHash).Str("repository_path", g.path).Msg("Entering GetCommitRange")
	// Resolve the hashes
	hashes, err := g.resolveHashRange(ctx, fromHash, toHash)
	if err != nil {
		// Error is already enhanced from resolveHashRange
		return nil, err
	}

	// Create iterator - we don't use getCommitsFromHash here because we need a custom stop condition
	iter, err := g.createCommitIterator(ctx, hashes.toHash)
	if err != nil {
		// Create enhanced error
		errCtx := appErrors.NewContext()

		richErr := appErrors.CreateRichError(
			"GitRepository",
			appErrors.ErrGitOperationFailed,
			fmt.Sprintf("Failed to create commit iterator for hash %s: %s", toHash, err),
			"Check that the repository is accessible and the commit reference is valid.",
			errCtx,
		)

		// Add context
		richErr = richErr.WithContext("to_hash", toHash)
		richErr = richErr.WithContext("repository_path", g.path)

		return nil, richErr
	}

	// Collect and convert commits until we reach the "from" commit
	domainCommits, err := g.collectAndConvertCommits(ctx, iter, 0, func(commit *object.Commit) bool {
		return commit.Hash == hashes.fromHash
	})
	if err != nil {
		// Create enhanced error
		errCtx := appErrors.NewContext()

		richErr := appErrors.CreateRichError(
			"GitRepository",
			appErrors.ErrGitOperationFailed,
			fmt.Sprintf("Failed to collect commits from range %s..%s: %s", fromHash, toHash, err),
			"Check that the repository is accessible and the commit references are valid.",
			errCtx,
		)

		// Add context
		richErr = richErr.WithContext("from_hash", fromHash)
		richErr = richErr.WithContext("to_hash", toHash)
		richErr = richErr.WithContext("repository_path", g.path)

		return nil, richErr
	}

	// Create a new immutable collection (rather than mutating an existing one)
	result, err := g.ensureFromCommitIncluded(ctx, domainCommits, hashes.fromHash)
	if err != nil {
		// Error is already enhanced from ensureFromCommitIncluded
		return nil, err
	}

	return result, nil
}

// hashRange represents a range of commit hashes.
// This immutable value type cleanly encapsulates the range data.
type hashRange struct {
	fromHash plumbing.Hash
	toHash   plumbing.Hash
}

// resolveHashRange resolves the 'from' and 'to' revision strings to actual hash objects.
// This pure function handles the resolution without modifying state.
func (g RepositoryAdapter) resolveHashRange(
	ctx context.Context,
	fromHashStr,
	toHashStr string,
) (hashRange, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("from_hash_str", fromHashStr).Str("to_hash_str", toHashStr).Msg("Entering resolveHashRange")

	// Resolve the "to" hash
	toHash, err := g.resolveRevision(ctx, toHashStr)
	if err != nil {
		// Create enhanced error
		errCtx := appErrors.NewContext()

		richErr := appErrors.CreateRichError(
			"GitRepository",
			appErrors.ErrRangeNotFound,
			fmt.Sprintf("Failed to resolve 'to' hash %s: %s", toHashStr, err),
			"Check that the 'to' reference exists in the repository.",
			errCtx,
		)

		// Add context
		richErr = richErr.WithContext("to_hash", toHashStr)
		richErr = richErr.WithContext("repository_path", g.path)

		return hashRange{}, richErr
	}

	// Resolve the "from" hash
	fromHash, err := g.resolveRevision(ctx, fromHashStr)
	if err != nil {
		// Create enhanced error
		errCtx := appErrors.NewContext()

		richErr := appErrors.CreateRichError(
			"GitRepository",
			appErrors.ErrRangeNotFound,
			fmt.Sprintf("Failed to resolve 'from' hash %s: %s", fromHashStr, err),
			"Check that the 'from' reference exists in the repository.",
			errCtx,
		)

		// Add context
		richErr = richErr.WithContext("from_hash", fromHashStr)
		richErr = richErr.WithContext("repository_path", g.path)

		return hashRange{}, richErr
	}

	return hashRange{
		fromHash: fromHash,
		toHash:   toHash,
	}, nil
}

// ensureFromCommitIncluded creates a new commit collection that includes the 'from' commit.
// This pure function returns a new slice rather than modifying an existing one.
func (g RepositoryAdapter) ensureFromCommitIncluded(
	ctx context.Context,
	commits []domain.CommitInfo,
	fromHash plumbing.Hash,
) ([]domain.CommitInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("from_hash", fromHash.String()).Int("commit_count", len(commits)).Msg("Entering ensureFromCommitIncluded")

	// Create a new collection for immutable processing
	collection := domain.NewCommitCollection(commits)

	// If "from" commit is already included, return the original collection
	if collection.Contains(fromHash.String()) {
		return collection.All(), nil
	}

	// Get the "from" commit
	fromCommit, err := g.getCommitByHash(ctx, fromHash)
	if err != nil {
		return nil, err
	}

	// Convert to domain commit
	domainFromCommit := g.convertCommit(ctx, fromCommit)

	// Create a new collection with the additional commit
	newCollection := domain.NewCommitCollection(commits)
	newCollection.Add(domainFromCommit)

	return newCollection.All(), nil
}

// GetHeadCommits returns the specified number of commits from HEAD.
func (g RepositoryAdapter) GetHeadCommits(ctx context.Context, count int) ([]domain.CommitInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Int("count", count).Str("repository_path", g.path).Msg("Entering GetHeadCommits")
	// This is already a nicely structured functional approach, but we'll make it more explicit
	return g.getCommitsFromHash(ctx, "", count)
}

// getCommitsFromHash is a helper function that fetches commits from a specific hash.
// This pure function encapsulates the commit fetching logic to avoid duplication.
func (g RepositoryAdapter) getCommitsFromHash(
	ctx context.Context,
	hashStr string,
	count int,
) ([]domain.CommitInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("hash_str", hashStr).Int("count", count).Msg("Entering getCommitsFromHash")

	// Resolve hash
	hash, err := g.resolveRevision(ctx, hashStr) // Empty string means HEAD
	if err != nil {
		return nil, fmt.Errorf("failed to resolve revision %q: %w", hashStr, err)
	}

	// Create iterator
	iter, err := g.createCommitIterator(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to create commit iterator: %w", err)
	}

	// Collect and convert commits with limit
	return g.collectAndConvertCommits(ctx, iter, count, nil)
}

// GetCurrentBranch returns the name of the current branch.
// This uses functional patterns to maintain immutability and avoid state mutation.
func (g RepositoryAdapter) GetCurrentBranch(ctx context.Context) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("repository_path", g.path).Msg("Entering GetCurrentBranch")
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

	// Try to find a branch pointing to HEAD
	branchName, err := g.findBranchForCommit(ctx, headHash)
	if err != nil {
		// Only return error if it's not a "not found" type of error
		if err.Error() != "branch not found" {
			return "", fmt.Errorf("failed to find branch: %w", err)
		}
	}

	// If we found a branch, return it
	if branchName != "" {
		return branchName, nil
	}

	// We're in a detached HEAD state with no matching branch
	return "HEAD detached at " + headHash.String()[:7], nil
}

// findBranchForCommit finds a branch that points to the given commit.
// This is a pure function that implements the branch search functionally.
func (g RepositoryAdapter) findBranchForCommit(ctx context.Context, commitHash plumbing.Hash) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("commit_hash", commitHash.String()).Msg("Entering findBranchForCommit")
	// Get all branches
	branches, err := g.repo.Branches()
	if err != nil {
		return "", fmt.Errorf("failed to get branches: %w", err)
	}

	// Use a functional approach to find the matching branch
	// We process each branch in isolation without state mutation
	var result string

	err = branches.ForEach(func(branch *plumbing.Reference) error {
		if branch.Hash() == commitHash {
			// Store the result (immutable pattern would avoid this,
			// but we're constrained by the Branches() API)
			result = branch.Name().Short()

			return errors.New("stop") // Use error to stop iteration
		}

		return nil
	})

	// ForEach returns a "stop" error when we've found the branch
	if err != nil && err.Error() != "stop" {
		return "", fmt.Errorf("failed to iterate branches: %w", err)
	}

	if result != "" {
		return result, nil
	}

	// No matching branch found
	return "", errors.New("branch not found")
}

// GetRepositoryName returns the name of the repository.
func (g RepositoryAdapter) GetRepositoryName(_ context.Context) string {
	// No need to check for context cancellation for this simple operation
	// Extract the repository name from the path
	return filepath.Base(g.path)
}

// findGitDir is moved to repository_helpers.go

// convertCommit converts a go-git commit to a domain commit.
// This function is responsible for mapping all infrastructure-specific commit data
// to our domain model, ensuring that domain logic never has to access implementation details.
func (g RepositoryAdapter) convertCommit(ctx context.Context, commit *object.Commit) domain.CommitInfo {
	logger := log.Logger(ctx)
	logger.Trace().Str("commit_hash", commit.Hash.String()).Msg("Entering convertCommit")
	// Split the commit message into subject and body
	message := commit.Message
	subject, body := domain.SplitCommitMessage(message)

	// Check if the commit is a merge commit
	isMergeCommit := len(commit.ParentHashes) > 1

	// Format commit date as ISO string
	commitDate := commit.Committer.When.Format("2006-01-02T15:04:05Z07:00")

	// Create domain commit with all necessary information extracted from raw commit
	domainCommit := domain.CommitInfo{
		Hash:          commit.Hash.String(),
		Subject:       subject,
		Body:          body,
		Message:       message,
		AuthorName:    commit.Author.Name,
		AuthorEmail:   commit.Author.Email,
		CommitDate:    commitDate,
		IsMergeCommit: isMergeCommit,
	}

	// Get signature if available
	if commit.PGPSignature != "" {
		domainCommit.Signature = commit.PGPSignature
	}

	return domainCommit
}

// IsValid checks if the repository is a valid Git repository.
func (g RepositoryAdapter) IsValid(_ context.Context) bool {
	// No need to check for context cancellation for this simple operation
	// We were able to open the repository, so it's valid
	return g.repo != nil
}

// GetCommitsAhead returns the number of commits ahead of the given reference.
func (g RepositoryAdapter) GetCommitsAhead(ctx context.Context, reference string) (int, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("reference", reference).Str("repository_path", g.path).Msg("Entering GetCommitsAhead")
	// Get all the necessary information to compute commits ahead
	head, _, mergeBase, err := g.resolveCommitReferences(ctx, reference)
	if err != nil {
		return 0, err
	}

	// Create iterator from HEAD
	iter, err := g.createCommitIterator(ctx, head)
	if err != nil {
		return 0, err
	}

	// Collect commits between HEAD and merge base
	commits, err := g.collectCommits(ctx, iter, 0, func(commit *object.Commit) bool {
		return commit.Hash == mergeBase
	})

	if err != nil {
		return 0, fmt.Errorf("failed to count commits ahead: %w", err)
	}

	// Return the count of commits
	return len(commits), nil
}

// resolveCommitReferences resolves all necessary references for commit comparison.
// This pure function gathers all the information needed in a single place.
func (g RepositoryAdapter) resolveCommitReferences(
	ctx context.Context,
	reference string,
) (plumbing.Hash, plumbing.Hash, plumbing.Hash, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("reference", reference).Msg("Entering resolveCommitReferences")

	// Resolve HEAD
	head, err := g.resolveRevision(ctx, "")
	if err != nil {
		return plumbing.ZeroHash, plumbing.ZeroHash, plumbing.ZeroHash, err
	}

	// Resolve reference
	refHash, err := g.resolveRevision(ctx, reference)
	if err != nil {
		return plumbing.ZeroHash, plumbing.ZeroHash, plumbing.ZeroHash,
			fmt.Errorf("failed to resolve reference %s: %w", reference, err)
	}

	// Find merge base
	mergeBase, err := g.findMergeBase(ctx, head, refHash)
	if err != nil {
		return plumbing.ZeroHash, plumbing.ZeroHash, plumbing.ZeroHash,
			fmt.Errorf("failed to find merge base: %w", err)
	}

	return head, refHash, mergeBase, nil
}

// findMergeBase finds the common ancestor of two commits.
// This delegates to the pure function implementation in repository_helpers.go.
func (g RepositoryAdapter) findMergeBase(ctx context.Context, hash1, hash2 plumbing.Hash) (plumbing.Hash, error) {
	return findMergeBase(ctx, g.repo, hash1, hash2)
}

// resolveRevision resolves a revision to a hash.
// If the revision is empty, HEAD is used.
// This delegates to the pure function implementation in repository_helpers.go.
func (g RepositoryAdapter) resolveRevision(ctx context.Context, revision string) (plumbing.Hash, error) {
	return resolveRevision(ctx, g.repo, revision)
}

// getCommitByHash gets a commit by its hash.
func (g RepositoryAdapter) getCommitByHash(ctx context.Context, hash plumbing.Hash) (*object.Commit, error) {
	return getCommitByHash(ctx, g.repo, hash)
}

// createCommitIterator creates a commit iterator starting from the given hash.
func (g RepositoryAdapter) createCommitIterator(ctx context.Context, hash plumbing.Hash) (object.CommitIter, error) {
	return createCommitIterator(ctx, g.repo, hash)
}

// collectCommits collects commits from an iterator, with optional limit and stop condition.
func (g RepositoryAdapter) collectCommits(
	ctx context.Context,
	iter object.CommitIter,
	limit int,
	stopFn func(*object.Commit) bool,
) ([]*object.Commit, error) {
	return collectCommits(ctx, iter, limit, stopFn)
}

// collectAndConvertCommits collects commits from an iterator, converts them to domain commits.
// This is now implemented using a functional approach with value semantics.
func (g RepositoryAdapter) collectAndConvertCommits(
	ctx context.Context,
	iter object.CommitIter,
	limit int,
	stopFn func(*object.Commit) bool,
) ([]domain.CommitInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Int("limit", limit).Bool("has_stop_fn", stopFn != nil).Msg("Entering collectAndConvertCommits")

	// Collect git commits
	commits, err := g.collectCommits(ctx, iter, limit, stopFn)
	if err != nil {
		return nil, err
	}

	// Convert to domain commits using mapCommits function for cleaner, functional transformation
	return g.mapCommits(ctx, commits)
}

// mapCommits transforms a slice of git commits to domain commits.
// This pure function handles the transformation without modifying state.
func (g RepositoryAdapter) mapCommits(ctx context.Context, commits []*object.Commit) ([]domain.CommitInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Int("commit_count", len(commits)).Msg("Entering mapCommits")
	// Use contextx.Map for a more concise functional transformation
	domainCommits := contextx.Map(commits, func(commit *object.Commit) domain.CommitInfo {
		return g.convertCommit(ctx, commit)
	})

	return domainCommits, nil
}
