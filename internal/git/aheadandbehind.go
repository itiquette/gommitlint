// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/itiquette/gommitlint/internal/model"
)

// AheadAndBehind returns the number of commits that HEAD is ahead and behind
// relative to the specified ref.
func AheadAndBehind(gitPtr *model.Repository, ref string) (int, int, error) {
	// Get references
	ref1, err := gitPtr.Repo.Reference(plumbing.ReferenceName(ref), true)
	if err != nil {
		return 0, 0, fmt.Errorf("reference not found: %w", err)
	}

	ref2, err := gitPtr.Repo.Head()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Quick check: if references are identical, return 0, 0
	if ref1.Hash() == ref2.Hash() {
		return 0, 0, nil
	}

	// Find the merge base first to optimize counting
	commitObject1, err := gitPtr.Repo.CommitObject(ref1.Hash())
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get commit for %s: %w", ref, err)
	}

	commitObject2, err := gitPtr.Repo.CommitObject(ref2.Hash())
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	// Find merge base
	mergeBase, err := commitObject1.MergeBase(commitObject2)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to find merge base: %w", err)
	}

	// Check if we got a valid merge base
	if len(mergeBase) == 0 {
		// No common ancestor found - branches have completely diverged
		// Fall back to direct counting method
		ahead, behindErr := countCommitsBetween(gitPtr.Repo, ref1.Hash(), ref2.Hash(), "ahead")
		if behindErr != nil {
			return 0, 0, fmt.Errorf("failed to count ahead commits: %w", behindErr)
		}

		behind, aheadErr := countCommitsBetween(gitPtr.Repo, ref2.Hash(), ref1.Hash(), "behind")
		if aheadErr != nil {
			return 0, 0, fmt.Errorf("failed to count behind commits: %w", aheadErr)
		}

		return ahead, behind, nil
	}

	// Use merge base to count commits
	baseHash := mergeBase[0].Hash

	// Count ahead commits (from merge base to HEAD)
	ahead, err := countCommitsFrom(gitPtr.Repo, baseHash, ref2.Hash())
	if err != nil {
		return 0, 0, fmt.Errorf("error counting ahead commits: %w", err)
	}

	// Count behind commits (from merge base to ref)
	behind, err := countCommitsFrom(gitPtr.Repo, baseHash, ref1.Hash())
	if err != nil {
		return 0, 0, fmt.Errorf("error counting behind commits: %w", err)
	}

	return ahead, behind, nil
}

// countCommitsFrom counts commits from base to tip (excluding base).
func countCommitsFrom(repo *git.Repository, baseHash, tipHash plumbing.Hash) (int, error) {
	// Don't count if they're the same commit
	if baseHash == tipHash {
		return 0, nil
	}

	revList, err := repo.Log(&git.LogOptions{
		From:  tipHash,
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return 0, err
	}

	count := 0
	err = revList.ForEach(func(c *object.Commit) error {
		if c.Hash == baseHash {
			return storer.ErrStop
		}

		count++

		return nil
	})

	return count, err
}

// countCommitsBetween counts commits that are in 'to' but not in 'from'.
// This is a fallback method used when branches have completely diverged
// and no common ancestor exists.
// The direction parameter is only used for error messages.
func countCommitsBetween(repo *git.Repository, from, too plumbing.Hash, direction string) (int, error) {
	// Don't count if they're the same commit
	if from == too {
		return 0, nil
	}

	revList, err := repo.Log(&git.LogOptions{
		From:  too,
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get log for %s count: %w", direction, err)
	}

	count := 0

	err = revList.ForEach(func(c *object.Commit) error {
		if c.Hash == from {
			return storer.ErrStop
		}

		count++

		return nil
	})

	if err != nil {
		return 0, err
	}

	// If we didn't find the target commit in the history,
	// it means these branches have completely separate histories.
	// The count represents all commits in this branch.

	return count, nil
}
