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
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
)

// Compile-time interface satisfaction checks.
var (
	_ domain.CommitRepository       = (*RepositoryAdapter)(nil)
	_ domain.RepositoryInfoProvider = (*RepositoryAdapter)(nil)
	_ domain.CommitAnalyzer         = (*RepositoryAdapter)(nil)
)

// RepositoryAdapter adapts a git repository to the domain model.
// This version uses value semantics throughout.
type RepositoryAdapter struct {
	repo   *git.Repository // Needs to remain a pointer as go-git requires this
	path   string
	logger outgoing.Logger // Injected logger
}

// NewRepositoryAdapter creates a new RepositoryAdapter for the given path.
// This version returns a value rather than a pointer.
func NewRepositoryAdapter(ctx context.Context, path string, logger outgoing.Logger) (RepositoryAdapter, error) {
	logger.Debug("Entering NewRepositoryAdapter", "path", path)
	// If path is empty, use current directory
	if path == "" {
		var err error

		path, err = os.Getwd()
		if err != nil {
			return RepositoryAdapter{}, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Find the git repository
	gitDir, err := findGitDir(ctx, path, logger)
	if err != nil {
		return RepositoryAdapter{}, fmt.Errorf("failed to find git directory: %w", err)
	}

	// Open the git repository
	repo, err := git.PlainOpen(gitDir)
	if err != nil {
		return RepositoryAdapter{}, fmt.Errorf("failed to open git repository: %w", err)
	}

	return RepositoryAdapter{
		repo:   repo,
		path:   gitDir,
		logger: logger,
	}, nil
}

// GetCommit returns a commit by its hash.
func (g RepositoryAdapter) GetCommit(ctx context.Context, hash string) (domain.CommitInfo, error) {
	g.logger.Debug("Entering GetCommit", "hash", hash, "repository_path", g.path)
	// Instead of repeating the logic for resolving and getting commits,
	// use our helper function to get exactly 1 commit
	commits, err := g.getCommitsFromHash(ctx, hash, 1)
	if err != nil {
		// Create an enhanced error
		richErr := appErrors.New(
			"GitRepository",
			appErrors.ErrGitOperationFailed,
			fmt.Sprintf("Failed to get commit with hash %s: %s", hash, err),
		).WithContextMap(map[string]string{"help": "Check that the commit reference is valid and the repository is accessible."})

		// Add extra context
		richErr = richErr.WithContextMap(map[string]string{"git_reference": hash})
		richErr = richErr.WithContextMap(map[string]string{"repository_path": g.path})

		return domain.CommitInfo{}, richErr
	}

	// Ensure we got exactly one commit
	if len(commits) == 0 {
		// Create an enhanced error
		richErr := appErrors.New(
			"GitRepository",
			appErrors.ErrCommitNotFound,
			"Commit not found: "+hash,
		).WithContextMap(map[string]string{"help": "Verify the commit reference is correct and exists in the repository."})

		// Add extra context
		richErr = richErr.WithContextMap(map[string]string{"git_reference": hash})
		richErr = richErr.WithContextMap(map[string]string{"repository_path": g.path})

		return domain.CommitInfo{}, richErr
	}

	return commits[0], nil
}

// GetCommits returns the last n commits from HEAD.
func (g RepositoryAdapter) GetCommits(ctx context.Context, count int) ([]domain.CommitInfo, error) {
	g.logger.Debug("Entering GetCommits", "count", count, "repository_path", g.path)
	// Use HEAD reference
	return g.getCommitsFromHash(ctx, "HEAD", count)
}

// getCommitsFromHash gets a specific number of commits starting from a given hash or reference.
// This helper function consolidates the logic for getting commits.
func (g RepositoryAdapter) getCommitsFromHash(_ context.Context, hashOrRef string, count int) ([]domain.CommitInfo, error) {
	// Resolve the reference to ensure we have a valid hash
	hash, err := g.repo.ResolveRevision(plumbing.Revision(hashOrRef))
	if err != nil {
		richErr := appErrors.New(
			"GitRepository",
			appErrors.ErrInvalidReference,
			fmt.Sprintf("Invalid git reference '%s': %s", hashOrRef, err),
		).WithContextMap(map[string]string{"help": "Ensure the reference (commit hash, branch name, or tag) exists in the repository."})

		// Add extra context
		richErr = richErr.WithContextMap(map[string]string{"git_reference": hashOrRef})
		richErr = richErr.WithContextMap(map[string]string{"repository_path": g.path})

		return nil, richErr
	}

	// Get the commit object
	commit, err := g.repo.CommitObject(*hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	// Create commit iterator starting from this commit
	iter := object.NewCommitIterCTime(commit, nil, nil)
	defer iter.Close()

	// Collect commits
	var commits []domain.CommitInfo

	collected := 0

	err = iter.ForEach(func(commit *object.Commit) error {
		if collected >= count && count > 0 {
			// Stop iteration once we have enough commits
			return errors.New("stop iteration")
		}

		commitInfo := createCommitInfo(commit)
		commits = append(commits, commitInfo)
		collected++

		return nil
	})

	// Ignore the "stop iteration" error as it's our way of breaking the loop
	if err != nil && err.Error() != "stop iteration" {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

// GetCommitRange returns commits between two references (inclusive).
func (g RepositoryAdapter) GetCommitRange(_ context.Context, fromHash, toHash string) ([]domain.CommitInfo, error) {
	g.logger.Debug("Entering GetCommitRange", "from", fromHash, "to", toHash, "repository_path", g.path)
	// Resolve both references
	fromOid, err := g.repo.ResolveRevision(plumbing.Revision(fromHash))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve 'from' reference '%s': %w", fromHash, err)
	}

	toOid, err := g.repo.ResolveRevision(plumbing.Revision(toHash))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve 'to' reference '%s': %w", toHash, err)
	}

	// Get both commit objects
	fromCommit, err := g.repo.CommitObject(*fromOid)
	if err != nil {
		return nil, fmt.Errorf("failed to get 'from' commit object: %w", err)
	}

	toCommit, err := g.repo.CommitObject(*toOid)
	if err != nil {
		return nil, fmt.Errorf("failed to get 'to' commit object: %w", err)
	}

	// Create a commit iterator from 'to' commit
	iter := object.NewCommitIterCTime(toCommit, nil, nil)
	defer iter.Close()

	// Collect commits until we reach 'from' commit or hit a reasonable limit
	var commits []domain.CommitInfo

	const maxCommits = 1000 // Reasonable limit to prevent excessive iteration

	foundFrom := false
	commitCount := 0

	err = iter.ForEach(func(commit *object.Commit) error {
		// Add safety limit
		if commitCount >= maxCommits {
			g.logger.Warn("Reached maximum commit limit in range iteration", "limit", maxCommits)

			return errors.New("stop iteration")
		}

		// Check if we've reached the 'from' commit - stop before adding it
		if commit.Hash == fromCommit.Hash {
			foundFrom = true

			g.logger.Debug("Found 'from' commit in iteration", "hash", commit.Hash.String())

			return errors.New("stop iteration")
		}

		// Add the commit (only if it's not the 'from' commit)
		commitInfo := createCommitInfo(commit)
		commits = append(commits, commitInfo)
		commitCount++

		return nil
	})

	// Handle iteration errors
	if err != nil && err.Error() != "stop iteration" {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	if !foundFrom && len(commits) > 0 {
		// The 'from' commit wasn't found in the range
		// This is expected when comparing diverged branches (e.g., main..feature-branch)
		// In this case, we need to find commits that are in 'to' but not in 'from'
		// We'll collect commits until we find a common ancestor
		// Get ancestors of the 'from' commit
		fromAncestors := make(map[plumbing.Hash]bool)

		fromIter := object.NewCommitIterCTime(fromCommit, nil, nil)
		defer fromIter.Close()

		err = fromIter.ForEach(func(commit *object.Commit) error {
			fromAncestors[commit.Hash] = true

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to get ancestors of %s: %w", fromHash, err)
		}

		// Filter commits to only include those not in 'from' branch
		var filteredCommits []domain.CommitInfo

		for _, commitInfo := range commits {
			hash := plumbing.NewHash(commitInfo.Hash)
			if !fromAncestors[hash] {
				filteredCommits = append(filteredCommits, commitInfo)
			}
		}

		g.logger.Debug("Filtered commits for range",
			"from", fromHash,
			"to", toHash,
			"total_commits", len(commits),
			"filtered_commits", len(filteredCommits),
			"from_ancestors", len(fromAncestors))

		return filteredCommits, nil
	}

	return commits, nil
}

// CountCommitsSince counts commits from HEAD to the given reference.
func (g RepositoryAdapter) CountCommitsSince(_ context.Context, reference string) (int, error) {
	g.logger.Debug("Entering CountCommitsSince", "reference", reference, "repository_path", g.path)
	// Get HEAD
	head, err := g.repo.Head()
	if err != nil {
		return 0, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get HEAD commit
	headCommit, err := g.repo.CommitObject(head.Hash())
	if err != nil {
		return 0, fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	// Resolve the reference
	refHash, err := g.repo.ResolveRevision(plumbing.Revision(reference))
	if err != nil {
		// Reference doesn't exist - count all commits
		iter := object.NewCommitIterCTime(headCommit, nil, nil)
		defer iter.Close()

		count := 0
		err = iter.ForEach(func(*object.Commit) error {
			count++

			return nil
		})

		if err != nil {
			return 0, fmt.Errorf("failed to count commits: %w", err)
		}

		return count, nil
	}

	// Get the reference commit
	refCommit, err := g.repo.CommitObject(*refHash)
	if err != nil {
		return 0, fmt.Errorf("failed to get reference commit: %w", err)
	}

	// If HEAD and reference are the same, there are 0 commits difference
	if headCommit.Hash == refCommit.Hash {
		return 0, nil
	}

	// Check if reference is an ancestor of HEAD
	isAncestor, err := isAncestor(g.repo, refCommit.Hash, headCommit.Hash)
	if err != nil {
		return 0, fmt.Errorf("failed to check ancestry: %w", err)
	}

	if !isAncestor {
		// Reference is not an ancestor of HEAD
		return 0, fmt.Errorf("reference %s is not an ancestor of HEAD", reference)
	}

	// Count commits from HEAD until we reach the reference
	iter := object.NewCommitIterCTime(headCommit, nil, nil)
	defer iter.Close()

	count := 0
	found := false

	err = iter.ForEach(func(commit *object.Commit) error {
		// Don't count the reference commit itself
		if commit.Hash == refCommit.Hash {
			found = true

			return errors.New("stop iteration")
		}

		count++

		return nil
	})

	if err != nil && err.Error() != "stop iteration" {
		return 0, fmt.Errorf("failed to iterate commits: %w", err)
	}

	if !found {
		return 0, errors.New("reference commit not found in history")
	}

	return count, nil
}

// GetRepositoryName returns the name of the repository.
func (g RepositoryAdapter) GetRepositoryName(_ context.Context) string {
	return filepath.Base(g.path)
}

// IsValid checks if the repository is valid.
func (g RepositoryAdapter) IsValid(_ context.Context) (bool, error) {
	// Try to get HEAD as a basic validity check
	_, err := g.repo.Head()
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			// Repository exists but has no commits
			return true, nil
		}

		return false, err
	}

	return true, nil
}

// GetHeadCommits returns the specified number of commits from HEAD.
func (g RepositoryAdapter) GetHeadCommits(ctx context.Context, count int) ([]domain.CommitInfo, error) {
	g.logger.Debug("Getting HEAD commits", "count", count)

	// Get HEAD reference
	ref, err := g.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get commit object for HEAD
	commit, err := g.repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	// Create iterator
	iter := object.NewCommitIterCTime(commit, nil, nil)
	defer iter.Close()

	// Collect commits
	commits, err := collectCommits(ctx, iter, count, nil, g.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to collect commits: %w", err)
	}

	// Convert to domain.CommitInfo
	result := make([]domain.CommitInfo, len(commits))
	for i, commit := range commits {
		result[i] = createCommitInfo(commit)
	}

	return result, nil
}

// GetCurrentBranch returns the name of the current branch.
func (g RepositoryAdapter) GetCurrentBranch(_ context.Context) (string, error) {
	g.logger.Debug("Getting current branch")

	// Get HEAD reference
	ref, err := g.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Check if HEAD is detached
	if !ref.Name().IsBranch() {
		return "", errors.New("HEAD is detached")
	}

	// Extract branch name
	branchName := ref.Name().Short()
	g.logger.Debug("Current branch found", "branch", branchName)

	return branchName, nil
}

// GetCommitsAhead returns the number of commits ahead of a reference.
func (g RepositoryAdapter) GetCommitsAhead(ctx context.Context, ref string) (int, error) {
	// This is the same as CountCommitsSince
	return g.CountCommitsSince(ctx, ref)
}

// createCommitInfo creates a domain.CommitInfo from a git commit object.
func createCommitInfo(commit *object.Commit) domain.CommitInfo {
	// Split message into subject and body
	subject, body := domain.SplitCommitMessage(commit.Message)

	return domain.CommitInfo{
		Hash:          commit.Hash.String(),
		Subject:       subject,
		Body:          body,
		Message:       commit.Message,
		Signature:     commit.PGPSignature,
		AuthorName:    commit.Author.Name,
		AuthorEmail:   commit.Author.Email,
		CommitDate:    commit.Author.When.Format("2006-01-02T15:04:05Z"),
		IsMergeCommit: len(commit.ParentHashes) > 1,
	}
}

// isAncestor checks if 'ancestor' is an ancestor of 'descendant'.
func isAncestor(repo *git.Repository, ancestor, descendant plumbing.Hash) (bool, error) {
	// Get the descendant commit
	descCommit, err := repo.CommitObject(descendant)
	if err != nil {
		return false, err
	}

	// Walk the commit history from descendant
	iter := object.NewCommitIterCTime(descCommit, nil, nil)
	defer iter.Close()

	found := false
	err = iter.ForEach(func(commit *object.Commit) error {
		if commit.Hash == ancestor {
			found = true

			return errors.New("stop iteration")
		}

		return nil
	})

	if err != nil && err.Error() != "stop iteration" {
		return false, err
	}

	return found, nil
}

// findGitDir finds the .git directory starting from the given path.
func findGitDir(_ context.Context, startPath string, logger outgoing.Logger) (string, error) {
	logger.Debug("Looking for git directory", "start_path", startPath)
	// Clean and make the path absolute
	path, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if the path itself is a git directory
	if isGitDir(path) {
		return path, nil
	}

	// If the path ends with .git, try its parent
	if filepath.Base(path) == ".git" {
		parent := filepath.Dir(path)
		if isGitDir(parent) {
			return parent, nil
		}
	}

	// Walk up the directory tree looking for a .git directory
	current := path

	for {
		gitPath := filepath.Join(current, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			if info.IsDir() {
				// Found a .git directory
				return current, nil
			}
			// .git file (submodule or worktree)
			// For now, we'll treat the parent as the git directory
			return current, nil
		}

		// Move to parent directory
		parent := filepath.Dir(current)
		if parent == current {
			// Reached the root directory
			break
		}

		current = parent
	}

	// If we're in a subdirectory, the original path might be inside a git repo
	// Try to open it as a git repository
	if _, err := git.PlainOpen(path); err == nil {
		return path, nil
	}

	// Try parent directories
	current = path

	for {
		if _, err := git.PlainOpen(current); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}

		current = parent
	}

	return "", fmt.Errorf("not inside a git repository: could not find .git directory in %s or any parent directory", startPath)
}

// isGitDir checks if a path is a git directory.
func isGitDir(path string) bool {
	// Check for .git subdirectory
	gitPath := filepath.Join(path, ".git")
	if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
		return true
	}

	// Check if the path itself can be opened as a git repository
	if _, err := git.PlainOpen(path); err == nil {
		return true
	}

	return false
}
