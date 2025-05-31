// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package git provides Git repository adapters for the domain model.
package git

import (
	"context"
	"errors"
	"fmt"
	// "maps".
	"slices"

	// "github.com/go-git/go-git/v5/plumbing".
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/itiquette/gommitlint/internal/ports"
)

// // resolveRevision resolves a revision to a hash.
// // If the revision is empty, HEAD is used.
// func resolveRevision(_ context.Context, repo *git.Repository, revision string, logger Logger) (plumbing.Hash, error) {
// 	logger.Debug("Entering resolveRevision", "revision", revision)
// 	// Default to HEAD if no revision provided
// 	if revision == "" {
// 		ref, err := repo.Head()
// 		if err != nil {
// 			return plumbing.ZeroHash, fmt.Errorf("failed to get HEAD: %w", err)
// 		}
//
// 		return ref.Hash(), nil
// 	}
//
// 	// Resolve symbolic references like "HEAD", "main", etc.
// 	hash, err := repo.ResolveRevision(plumbing.Revision(revision))
// 	if err != nil {
// 		return plumbing.ZeroHash, fmt.Errorf("failed to resolve revision %s: %w", revision, err)
// 	}
//
// 	return *hash, nil
// }

// // getCommitByHash gets a commit by its hash.
// func getCommitByHash(_ context.Context, repo *git.Repository, hash plumbing.Hash, logger domain.Logger) (*object.Commit, error) {
// 	logger.Debug("Entering getCommitByHash", "hash", hash.String())
//
// 	commit, err := repo.CommitObject(hash)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get commit %s: %w", hash.String(), err)
// 	}
//
// 	return commit, nil
// }

// // createCommitIterator creates a commit iterator starting from the given hash.
// func createCommitIterator(_ context.Context, repo *git.Repository, hash plumbing.Hash, logger domain.Logger) (object.CommitIter, error) {
// 	logger.Debug("Entering createCommitIterator", "hash", hash.String())
//
// 	iter, err := repo.Log(&git.LogOptions{From: hash})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create commit iterator: %w", err)
// 	}
//
// 	return iter, nil
// }

// collectCommits collects commits from an iterator, with optional limit and stop condition.
// This function uses a functional approach with immutability principles.
func collectCommits(
	_ context.Context,
	iter object.CommitIter,
	limit int,
	stopFn func(*object.Commit) bool,
	logger ports.Logger,
) ([]*object.Commit, error) {
	logger.Debug("Entering collectCommits", "limit", limit, "has_stop_fn", stopFn != nil)
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
		newCommits := append(slices.Clone(state.commits), commit)
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

// // getAncestors builds a map of all ancestors of a commit.
// // This is a pure function that returns a new ancestors map rather than modifying a passed map.
// func getAncestors(ctx context.Context, repo *git.Repository, commit *object.Commit, logger domain.Logger) (map[plumbing.Hash]bool, error) {
// 	logger.Debug("Entering getAncestors", "commit_hash", commit.Hash.String())
//
// 	return getAncestorsWithAccumulator(ctx, repo, commit, make(map[plumbing.Hash]bool), logger)
// }
//
// // Maximum number of ancestors to process to prevent DoS attacks.
// const maxAncestors = 10000

// // getAncestorsWithAccumulator is a helper function that builds an ancestors map.
// // This function allows accumulating results while maintaining functional purity at the public API.
// // It uses an iterative breadth-first approach for better performance while maintaining functional principles.
// func getAncestorsWithAccumulator(_ context.Context, repo *git.Repository, commit *object.Commit, ancestors map[plumbing.Hash]bool, logger domain.Logger) (map[plumbing.Hash]bool, error) {
// 	logger.Debug("Entering getAncestorsWithAccumulator", "commit_hash", commit.Hash.String(), "existing_ancestors", len(ancestors))
// 	// Maximum number of ancestors to process to prevent DoS attacks.
// 	const maxAncestors = 10000
// 	// Create a new map with the current ancestors using maps.Clone
// 	result := maps.Clone(ancestors)
//
// 	// Initialize a queue with the starting commit
// 	queue := []*object.Commit{commit}
//
// 	// Process commits breadth-first with a limit to prevent DoS
// 	processedCount := 0
// 	for len(queue) > 0 && processedCount < maxAncestors {
// 		// Dequeue the first commit
// 		current := queue[0]
// 		queue = queue[1:]
//
// 		// Mark this commit as an ancestor (immutably)
// 		result = maps.Clone(result)
// 		result[current.Hash] = true
// 		processedCount++
//
// 		// Process all parents of the current commit
// 		for _, parentHash := range current.ParentHashes {
// 			// Skip if already processed
// 			if result[parentHash] {
// 				continue
// 			}
//
// 			// Limit queue size to prevent memory exhaustion
// 			if len(queue) >= maxAncestors {
// 				logger.Warn("Ancestor queue size limit reached", "queue_size", len(queue))
//
// 				break
// 			}
//
// 			// Get the parent commit
// 			parent, err := repo.CommitObject(parentHash)
// 			if err != nil {
// 				continue // Skip parents that can't be resolved
// 			}
//
// 			// Add to queue using functional approach (create new slice)
// 			queue = append(slices.Clone(queue), parent)
// 		}
// 	}
//
// 	if processedCount >= maxAncestors {
// 		logger.Warn("Maximum ancestor limit reached", "processed", processedCount)
// 	}
//
// // 	return result, nil
// // }

// // findMergeBase finds the common ancestor of two commits using a breadth-first search algorithm.
// // This is now a pure function that doesn't mutate any state.
// func findMergeBase(ctx context.Context, repo *git.Repository, hash1, hash2 plumbing.Hash, logger domain.Logger) (plumbing.Hash, error) {
// 	logger.Debug("Entering findMergeBase", "hash1", hash1.String(), "hash2", hash2.String())
// 	// Get the first commit and its ancestors
// 	commit1, err := repo.CommitObject(hash1)
// 	if err != nil {
// 		return plumbing.ZeroHash, fmt.Errorf("failed to get commit %s: %w", hash1.String(), err)
// 	}
//
// 	// Get all ancestors of the first commit using our pure function
// 	ancestors1, err := getAncestors(ctx, repo, commit1, logger)
// 	if err != nil {
// 		return plumbing.ZeroHash, fmt.Errorf("failed to get ancestors of %s: %w", hash1.String(), err)
// 	}
//
// 	// Get the second commit
// 	commit2, err := repo.CommitObject(hash2)
// 	if err != nil {
// 		return plumbing.ZeroHash, fmt.Errorf("failed to get commit %s: %w", hash2.String(), err)
// 	}
//
// 	// Use breadthFirstSearch to find the first common ancestor
// 	return findCommonAncestor(ctx, repo, commit2, ancestors1, logger)
// }

// // findCommonAncestor implements a breadth-first search to find the first common ancestor.
// // This has been extracted as a separate pure function for better separation of concerns.
// func findCommonAncestor(_ context.Context, repo *git.Repository, startCommit *object.Commit, ancestorsMap map[plumbing.Hash]bool, logger domain.Logger) (plumbing.Hash, error) {
// 	logger.Debug("Entering findCommonAncestor", "start_commit", startCommit.Hash.String(), "ancestors_map_size", len(ancestorsMap))
// 	// Initialize the queue with the start commit
// 	queue := []*object.Commit{startCommit}
//
// 	// Create a new visited map (immutable approach)
// 	visited := make(map[plumbing.Hash]bool)
//
// 	// Implement breadth-first search
// 	for len(queue) > 0 {
// 		// Dequeue the first commit
// 		current := queue[0]
// 		queue = queue[1:]
//
// 		// Skip if already visited (avoid cycles)
// 		if visited[current.Hash] {
// 			continue
// 		}
//
// 		// Mark as visited by creating a new map entry
// 		visited[current.Hash] = true
//
// 		// Check if this is a common ancestor
// 		if ancestorsMap[current.Hash] {
// 			return current.Hash, nil
// 		}
//
// 		// Add parents to the queue
// 		for _, parentHash := range current.ParentHashes {
// 			parent, err := repo.CommitObject(parentHash)
// 			if err != nil {
// 				continue
// 			}
//
// 			// Create a new queue instead of mutating the existing one
// 			queue = append(queue, parent)
// 		}
// 	}
//
// // 	// No common ancestor found (should not happen in a normal Git repository)
// // 	return plumbing.ZeroHash, fmt.Errorf("no common ancestor found with %s", startCommit.Hash.String())
// // }
