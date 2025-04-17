// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package git provides Git repository adapters for the domain model.
package git

import (
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
func (g *RepositoryAdapter) GetCommit(hash string) (*domain.CommitInfo, error) {
	// Resolve the hash (handles empty hash for HEAD)
	resolvedHash, err := g.resolveRevision(hash)
	if err != nil {
		return nil, err
	}

	// Get the commit
	commit, err := g.getCommitByHash(resolvedHash)
	if err != nil {
		return nil, err
	}

	// Convert to domain commit
	return g.convertCommit(commit)
}

// GetCommitRange returns all commits in the given range.
func (g *RepositoryAdapter) GetCommitRange(fromHash, toHash string) ([]*domain.CommitInfo, error) {
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
func (g *RepositoryAdapter) GetHeadCommits(count int) ([]*domain.CommitInfo, error) {
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

	// Collect and convert commits with limit
	return g.collectAndConvertCommits(iter, count, nil)
}

// GetCurrentBranch returns the name of the current branch.
func (g *RepositoryAdapter) GetCurrentBranch() (string, error) {
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

	// Get all branches
	branches, err := g.repo.Branches()
	if err != nil {
		return "", fmt.Errorf("failed to get branches: %w", err)
	}

	// Check if any branch points to HEAD
	var branchName string

	err = branches.ForEach(func(branch *plumbing.Reference) error {
		if branch.Hash() == headHash {
			branchName = branch.Name().Short()

			return errors.New("stop") // Use error to stop iteration
		}

		return nil
	})

	// ForEach returns a "stop" error when we've found the branch
	if err != nil && err.Error() != "stop" {
		return "", fmt.Errorf("failed to iterate branches: %w", err)
	}

	if branchName != "" {
		return branchName, nil
	}

	// We're in a detached HEAD state
	return "HEAD detached at " + headHash.String()[:7], nil
}

// GetRepositoryName returns the name of the repository.
func (g *RepositoryAdapter) GetRepositoryName() string {
	// Extract the repository name from the path
	return filepath.Base(g.path)
}

// findGitDir finds the Git directory from a starting path.
func findGitDir(start string) (string, error) {
	// Check if the directory exists
	info, err := os.Stat(start)
	if err != nil {
		return "", fmt.Errorf("failed to stat path %s: %w", start, err)
	}

	// If it's not a directory, use the parent directory
	if !info.IsDir() {
		start = filepath.Dir(start)
	}

	// Try to find .git directory by traversing up the directory tree
	current := start

	for {
		// Check if .git directory exists
		gitDir := filepath.Join(current, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return current, nil // Found .git directory
		}

		// Go up one level
		parent := filepath.Dir(current)
		if parent == current {
			// Reached the root directory, .git not found
			return "", fmt.Errorf("git repository not found in %s or any parent directory", start)
		}

		current = parent
	}
}

// convertCommit converts a go-git commit to a domain commit.
func (g *RepositoryAdapter) convertCommit(commit *object.Commit) (*domain.CommitInfo, error) {
	// Split the commit message into subject and body
	message := commit.Message
	subject, body := domain.SplitCommitMessage(message)

	// Check if the commit is a merge commit
	isMergeCommit := len(commit.ParentHashes) > 1

	// Create domain commit
	domainCommit := &domain.CommitInfo{
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
func (g *RepositoryAdapter) IsValid() bool {
	// We were able to open the repository, so it's valid
	return g.repo != nil
}

// GetCommitsAhead returns the number of commits ahead of the given reference.
func (g *RepositoryAdapter) GetCommitsAhead(reference string) (int, error) {
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

	// Collect commits between HEAD and merge base
	commits, err := g.collectCommits(iter, 0, func(commit *object.Commit) bool {
		return commit.Hash == mergeBase
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count commits ahead: %w", err)
	}

	// Return the count of commits
	return len(commits), nil
}

// findMergeBase finds the common ancestor of two commits.
func (g *RepositoryAdapter) findMergeBase(hash1, hash2 plumbing.Hash) (plumbing.Hash, error) {
	// Get the first commit and its ancestors
	commit1, err := g.repo.CommitObject(hash1)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get commit %s: %w", hash1.String(), err)
	}

	ancestors1 := make(map[plumbing.Hash]bool)

	err = g.getAncestors(commit1, ancestors1)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get ancestors of %s: %w", hash1.String(), err)
	}

	// Get the second commit
	commit2, err := g.repo.CommitObject(hash2)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get commit %s: %w", hash2.String(), err)
	}

	// Walk up the ancestry chain to find the first common ancestor
	queue := []*object.Commit{commit2}
	visited := make(map[plumbing.Hash]bool)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Skip if already visited
		if visited[current.Hash] {
			continue
		}

		visited[current.Hash] = true

		// Check if this is a common ancestor
		if ancestors1[current.Hash] {
			return current.Hash, nil
		}

		// Add parents to the queue
		for _, parentHash := range current.ParentHashes {
			parent, err := g.repo.CommitObject(parentHash)
			if err != nil {
				continue
			}

			queue = append(queue, parent)
		}
	}

	// No common ancestor found (should not happen in a normal Git repository)
	return plumbing.ZeroHash, fmt.Errorf("no common ancestor found between %s and %s", hash1.String(), hash2.String())
}

// getAncestors builds a map of all ancestors of a commit.
func (g *RepositoryAdapter) getAncestors(commit *object.Commit, ancestors map[plumbing.Hash]bool) error {
	// Mark this commit as an ancestor
	ancestors[commit.Hash] = true

	// Process parents
	for _, parentHash := range commit.ParentHashes {
		// Skip if already processed
		if ancestors[parentHash] {
			continue
		}

		parent, err := g.repo.CommitObject(parentHash)
		if err != nil {
			continue
		}

		err = g.getAncestors(parent, ancestors)
		if err != nil {
			return err
		}
	}

	return nil
}

// resolveRevision resolves a revision to a hash.
// If the revision is empty, HEAD is used.
func (g *RepositoryAdapter) resolveRevision(revision string) (plumbing.Hash, error) {
	// Default to HEAD if no revision provided
	if revision == "" {
		ref, err := g.repo.Head()
		if err != nil {
			return plumbing.ZeroHash, fmt.Errorf("failed to get HEAD: %w", err)
		}

		return ref.Hash(), nil
	}

	// Resolve symbolic references like "HEAD", "main", etc.
	hash, err := g.repo.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to resolve revision %s: %w", revision, err)
	}

	return *hash, nil
}

// getCommitByHash gets a commit by its hash.
func (g *RepositoryAdapter) getCommitByHash(hash plumbing.Hash) (*object.Commit, error) {
	commit, err := g.repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit %s: %w", hash.String(), err)
	}

	return commit, nil
}

// createCommitIterator creates a commit iterator starting from the given hash.
func (g *RepositoryAdapter) createCommitIterator(hash plumbing.Hash) (object.CommitIter, error) {
	iter, err := g.repo.Log(&git.LogOptions{From: hash})
	if err != nil {
		return nil, fmt.Errorf("failed to create commit iterator: %w", err)
	}

	return iter, nil
}

// collectCommits collects commits from an iterator, with optional limit and stop condition.
func (g *RepositoryAdapter) collectCommits(
	iter object.CommitIter,
	limit int,
	stopFn func(*object.Commit) bool,
) ([]*object.Commit, error) {
	var commits []*object.Commit

	count := 0

	// Iterate through the commits
	err := iter.ForEach(func(commit *object.Commit) error {
		// Check for nil commit
		if commit == nil {
			return errors.New("nil commit encountered")
		}

		// Check if we should stop processing
		if stopFn != nil && stopFn(commit) {
			return errors.New("stop")
		}

		// Add the commit to the list
		commits = append(commits, commit)
		count++

		// Check if we've reached the limit
		if limit > 0 && count >= limit {
			return errors.New("limit")
		}

		return nil
	})

	// The iterator will return an error if we stopped early
	if err != nil {
		if err.Error() == "stop" || err.Error() == "limit" {
			// These are expected control-flow errors, not actual errors
			return commits, nil
		}

		return nil, fmt.Errorf("error iterating commits: %w", err)
	}

	return commits, nil
}

// collectAndConvertCommits collects commits from an iterator, converts them to domain commits.
func (g *RepositoryAdapter) collectAndConvertCommits(
	iter object.CommitIter,
	limit int,
	stopFn func(*object.Commit) bool,
) ([]*domain.CommitInfo, error) {
	// Collect git commits
	commits, err := g.collectCommits(iter, limit, stopFn)
	if err != nil {
		return nil, err
	}

	// Convert to domain commits
	domainCommits := make([]*domain.CommitInfo, 0, len(commits))

	for _, commit := range commits {
		domainCommit, err := g.convertCommit(commit)
		if err != nil {
			return nil, err
		}

		domainCommits = append(domainCommits, domainCommit)
	}

	return domainCommits, nil
}
