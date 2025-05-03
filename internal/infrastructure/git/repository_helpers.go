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
	"github.com/itiquette/gommitlint/internal/contextx"
)

// findGitDir finds the Git directory from a starting path.
// Implements security best practices for safe path handling.
func findGitDir(start string) (string, error) {
	// Normalize the path to prevent path traversal issues
	start = filepath.Clean(start)

	// Check if the directory exists
	info, err := os.Stat(start)
	if err != nil {
		return "", fmt.Errorf("failed to stat path %s: %w", start, err)
	}

	// If it's not a directory, use the parent directory
	if !info.IsDir() {
		start = filepath.Dir(start)
	}

	// Check for absolute path and convert if needed
	if !filepath.IsAbs(start) {
		absPath, err := filepath.Abs(start)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for %s: %w", start, err)
		}

		start = absPath
	}

	// Try to find .git directory by traversing up the directory tree
	current := start

	// Set a reasonable limit to prevent excessive traversal (e.g., 20 levels up)
	const maxLevels = 20

	level := 0

	for level < maxLevels {
		// Check if .git directory exists
		gitDir := filepath.Join(current, ".git")

		// Use Lstat instead of Stat to avoid following symlinks
		if fi, err := os.Lstat(gitDir); err == nil {
			// Verify it's either a directory or a file (Git submodules use a file)
			if fi.IsDir() || fi.Mode().IsRegular() {
				return current, nil // Found .git directory
			}
		}

		// Go up one level
		parent := filepath.Dir(current)
		if parent == current {
			// Reached the root directory, .git not found
			return "", fmt.Errorf("git repository not found in %s or any parent directory", start)
		}

		current = parent
		level++
	}

	return "", fmt.Errorf("exceeded maximum directory traversal levels (%d) without finding git repository", maxLevels)
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
// This function uses a functional approach with immutability principles.
func collectCommits(
	iter object.CommitIter,
	limit int,
	stopFn func(*object.Commit) bool,
) ([]*object.Commit, error) {
	// Pre-allocate capacity if we know the limit
	initialCapacity := 10 // Default reasonable capacity
	if limit > 0 {
		initialCapacity = limit
	}

	// Initialize the result slice
	result := make([]*object.Commit, 0, initialCapacity)

	// Define a reducer function that accumulates commits
	collectNext := func(state struct {
		commits []*object.Commit
		count   int
		done    bool
		err     error
	}) (struct {
		commits []*object.Commit
		count   int
		done    bool
		err     error
	}, error) {
		// If we're done or have an error, return current state
		if state.done || state.err != nil {
			return state, state.err
		}

		// Get the next commit
		commit, err := iter.Next()

		// Handle completion of iteration
		if err != nil {
			// Normal end of iteration
			if err.Error() == "EOF" {
				return struct {
					commits []*object.Commit
					count   int
					done    bool
					err     error
				}{
					commits: state.commits,
					count:   state.count,
					done:    true,
					err:     nil,
				}, nil
			}

			// Real error
			return struct {
				commits []*object.Commit
				count   int
				done    bool
				err     error
			}{
				commits: state.commits,
				count:   state.count,
				done:    true,
				err:     fmt.Errorf("error iterating commits: %w", err),
			}, err
		}

		// Check for nil commit
		if commit == nil {
			nilErr := errors.New("nil commit encountered")

			return struct {
				commits []*object.Commit
				count   int
				done    bool
				err     error
			}{
				commits: state.commits,
				count:   state.count,
				done:    true,
				err:     nilErr,
			}, nilErr
		}

		// Check if we should stop processing
		if stopFn != nil && stopFn(commit) {
			return struct {
				commits []*object.Commit
				count   int
				done    bool
				err     error
			}{
				commits: state.commits,
				count:   state.count,
				done:    true,
				err:     nil,
			}, nil
		}

		// Create a new slice to maintain immutability
		newCommits := append(contextx.DeepCopy(state.commits), commit)
		newCount := state.count + 1

		// Check if we've reached the limit
		done := limit > 0 && newCount >= limit

		return struct {
			commits []*object.Commit
			count   int
			done    bool
			err     error
		}{
			commits: newCommits,
			count:   newCount,
			done:    done,
			err:     nil,
		}, nil
	}

	// Start with initial state
	state := struct {
		commits []*object.Commit
		count   int
		done    bool
		err     error
	}{
		commits: result,
		count:   0,
		done:    false,
		err:     nil,
	}

	// Keep collecting until done
	for !state.done {
		var err error

		state, err = collectNext(state)
		if err != nil {
			return state.commits, err
		}
	}

	return state.commits, state.err
}

// getAncestors builds a map of all ancestors of a commit.
// This is a pure function that returns a new ancestors map rather than modifying a passed map.
func getAncestors(repo *git.Repository, commit *object.Commit) (map[plumbing.Hash]bool, error) {
	return getAncestorsWithAccumulator(repo, commit, make(map[plumbing.Hash]bool))
}

// getAncestorsWithAccumulator is a helper function that builds an ancestors map.
// This function allows accumulating results while maintaining functional purity at the public API.
// It uses an iterative breadth-first approach for better performance while maintaining functional principles.
func getAncestorsWithAccumulator(repo *git.Repository, commit *object.Commit, ancestors map[plumbing.Hash]bool) (map[plumbing.Hash]bool, error) {
	// Create a new map with the current ancestors using DeepCopyMap
	result := contextx.DeepCopyMap(ancestors)

	// Initialize a queue with the starting commit
	queue := []*object.Commit{commit}

	// Process commits breadth-first
	for len(queue) > 0 {
		// Dequeue the first commit
		current := queue[0]
		queue = queue[1:]

		// Mark this commit as an ancestor (immutably)
		result = contextx.DeepCopyMap(result)
		result[current.Hash] = true

		// Process all parents of the current commit
		for _, parentHash := range current.ParentHashes {
			// Skip if already processed
			if result[parentHash] {
				continue
			}

			// Get the parent commit
			parent, err := repo.CommitObject(parentHash)
			if err != nil {
				continue // Skip parents that can't be resolved
			}

			// Add to queue using functional approach (create new slice)
			queue = append(contextx.DeepCopy(queue), parent)
		}
	}

	return result, nil
}

// findMergeBase finds the common ancestor of two commits using a breadth-first search algorithm.
// This is now a pure function that doesn't mutate any state.
func findMergeBase(repo *git.Repository, hash1, hash2 plumbing.Hash) (plumbing.Hash, error) {
	// Get the first commit and its ancestors
	commit1, err := repo.CommitObject(hash1)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get commit %s: %w", hash1.String(), err)
	}

	// Get all ancestors of the first commit using our pure function
	ancestors1, err := getAncestors(repo, commit1)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get ancestors of %s: %w", hash1.String(), err)
	}

	// Get the second commit
	commit2, err := repo.CommitObject(hash2)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get commit %s: %w", hash2.String(), err)
	}

	// Use breadthFirstSearch to find the first common ancestor
	return findCommonAncestor(repo, commit2, ancestors1)
}

// findCommonAncestor implements a breadth-first search to find the first common ancestor.
// This has been extracted as a separate pure function for better separation of concerns.
func findCommonAncestor(repo *git.Repository, startCommit *object.Commit, ancestorsMap map[plumbing.Hash]bool) (plumbing.Hash, error) {
	// Initialize the queue with the start commit
	queue := []*object.Commit{startCommit}

	// Create a new visited map (immutable approach)
	visited := make(map[plumbing.Hash]bool)

	// Implement breadth-first search
	for len(queue) > 0 {
		// Dequeue the first commit
		current := queue[0]
		queue = queue[1:]

		// Skip if already visited (avoid cycles)
		if visited[current.Hash] {
			continue
		}

		// Mark as visited by creating a new map entry
		visited[current.Hash] = true

		// Check if this is a common ancestor
		if ancestorsMap[current.Hash] {
			return current.Hash, nil
		}

		// Add parents to the queue
		for _, parentHash := range current.ParentHashes {
			parent, err := repo.CommitObject(parentHash)
			if err != nil {
				continue
			}

			// Create a new queue instead of mutating the existing one
			queue = append(queue, parent)
		}
	}

	// No common ancestor found (should not happen in a normal Git repository)
	return plumbing.ZeroHash, fmt.Errorf("no common ancestor found with %s", startCommit.Hash.String())
}
