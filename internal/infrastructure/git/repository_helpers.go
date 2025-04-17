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
)

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

// resolveRevision resolves a revision to a hash.
// If the revision is empty, HEAD is used.
func resolveRevision(repo *git.Repository, revision string) (plumbing.Hash, error) {
	// Default to HEAD if no revision provided
	if revision == "" {
		ref, err := repo.Head()
		if err != nil {
			return plumbing.ZeroHash, fmt.Errorf("failed to get HEAD: %w", err)
		}

		return ref.Hash(), nil
	}

	// Resolve symbolic references like "HEAD", "main", etc.
	hash, err := repo.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to resolve revision %s: %w", revision, err)
	}

	return *hash, nil
}

// getCommitByHash gets a commit by its hash.
func getCommitByHash(repo *git.Repository, hash plumbing.Hash) (*object.Commit, error) {
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit %s: %w", hash.String(), err)
	}

	return commit, nil
}

// createCommitIterator creates a commit iterator starting from the given hash.
func createCommitIterator(repo *git.Repository, hash plumbing.Hash) (object.CommitIter, error) {
	iter, err := repo.Log(&git.LogOptions{From: hash})
	if err != nil {
		return nil, fmt.Errorf("failed to create commit iterator: %w", err)
	}

	return iter, nil
}

// collectCommits collects commits from an iterator, with optional limit and stop condition.
func collectCommits(
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

// getAncestors builds a map of all ancestors of a commit.
func getAncestors(repo *git.Repository, commit *object.Commit, ancestors map[plumbing.Hash]bool) error {
	// Mark this commit as an ancestor
	ancestors[commit.Hash] = true

	// Process parents
	for _, parentHash := range commit.ParentHashes {
		// Skip if already processed
		if ancestors[parentHash] {
			continue
		}

		parent, err := repo.CommitObject(parentHash)
		if err != nil {
			continue
		}

		err = getAncestors(repo, parent, ancestors)
		if err != nil {
			return err
		}
	}

	return nil
}

// findMergeBase finds the common ancestor of two commits using a breadth-first search algorithm.
func findMergeBase(repo *git.Repository, hash1, hash2 plumbing.Hash) (plumbing.Hash, error) {
	// Get the first commit and its ancestors
	commit1, err := repo.CommitObject(hash1)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get commit %s: %w", hash1.String(), err)
	}

	ancestors1 := make(map[plumbing.Hash]bool)

	err = getAncestors(repo, commit1, ancestors1)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get ancestors of %s: %w", hash1.String(), err)
	}

	// Get the second commit
	commit2, err := repo.CommitObject(hash2)
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
			parent, err := repo.CommitObject(parentHash)
			if err != nil {
				continue
			}

			queue = append(queue, parent)
		}
	}

	// No common ancestor found (should not happen in a normal Git repository)
	return plumbing.ZeroHash, fmt.Errorf("no common ancestor found between %s and %s", hash1.String(), hash2.String())
}
