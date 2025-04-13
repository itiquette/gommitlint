// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/itiquette/gommitlint/internal/model"
)

// ErrBranchBehind indicates the current branch is behind the reference branch
// and needs to be updated before proceeding.
var ErrBranchBehind = errors.New("current branch is behind reference branch; please update your branch first")

// IsAhead reports whether HEAD is ahead of the specified reference branch.
// It returns the number of commits ahead and an error if behind.
func IsAhead(gitPtr *model.Repository, ref string) (int, error) {
	// Get references
	refObj, err := gitPtr.Repo.Reference(plumbing.ReferenceName(ref), true)
	if err != nil {
		return 0, fmt.Errorf("reference not found: %w", err)
	}

	headRef, err := gitPtr.Repo.Head()
	if err != nil {
		return 0, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Quick check: if references are identical, we're neither ahead nor behind
	if refObj.Hash() == headRef.Hash() {
		return 0, nil
	}

	// Get commit objects
	refCommit, err := gitPtr.Repo.CommitObject(refObj.Hash())
	if err != nil {
		return 0, fmt.Errorf("failed to get commit for %s: %w", ref, err)
	}

	headCommit, err := gitPtr.Repo.CommitObject(headRef.Hash())
	if err != nil {
		return 0, fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	// Find merge base (common ancestor)
	mergeBase, err := refCommit.MergeBase(headCommit)
	if err != nil {
		return 0, fmt.Errorf("failed to find merge base: %w", err)
	}

	// If no common ancestor found, branches have completely diverged
	if len(mergeBase) == 0 {
		return 0, errors.New("branches have completely diverged and have no common history")
	}

	baseHash := mergeBase[0].Hash

	// Count how many commits HEAD is ahead of the merge base
	ahead, err := countCommits(gitPtr.Repo, baseHash, headRef.Hash())
	if err != nil {
		return 0, fmt.Errorf("error counting ahead commits: %w", err)
	}

	// Count how many commits reference is ahead of the merge base
	behind, err := countCommits(gitPtr.Repo, baseHash, refObj.Hash())
	if err != nil {
		return 0, fmt.Errorf("error counting behind commits: %w", err)
	}

	// If we're behind at all, return an error
	if behind > 0 {
		return 0, ErrBranchBehind
	}

	return ahead, nil
}

// countCommits returns the number of commits from base to tip (excluding base).
func countCommits(repo *git.Repository, baseHash, tipHash plumbing.Hash) (int, error) {
	// If base and tip are the same, there are 0 commits between them
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
