// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"errors"
	"fmt"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/itiquette/gommitlint/internal/domain"
)

// SimpleRepository implements the CommitRepository port.
// Much simpler than current implementation - just what we need.
type SimpleRepository struct {
	repo *gogit.Repository
}

// NewSimpleRepository opens a git repository at the given path.
func NewSimpleRepository(path string) (*SimpleRepository, error) {
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("open repository: %w", err)
	}

	return &SimpleRepository{repo: repo}, nil
}

// GetCommit retrieves a single commit by hash.
func (r *SimpleRepository) GetCommit(hash string) (domain.Commit, error) {
	h := plumbing.NewHash(hash)

	commit, err := r.repo.CommitObject(h)
	if err != nil {
		return domain.Commit{}, fmt.Errorf("get commit: %w", err)
	}

	return r.convertCommit(commit), nil
}

// GetCommits retrieves commits in a range.
func (r *SimpleRepository) GetCommits(from, to string) ([]domain.Commit, error) {
	fromHash := plumbing.NewHash(from)
	toHash := plumbing.NewHash(to)

	// Get commit objects
	fromCommit, err := r.repo.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("get from commit: %w", err)
	}

	toCommit, err := r.repo.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("get to commit: %w", err)
	}

	// Find commits between
	var commits []domain.Commit

	iter, err := r.repo.Log(&gogit.LogOptions{From: toHash})
	if err != nil {
		return nil, fmt.Errorf("create iterator: %w", err)
	}

	// Collect commits until we reach 'from'
	err = iter.ForEach(func(c *object.Commit) error {
		if c.Hash == fromHash {
			return object.ErrCanceled // Stop iteration
		}

		commits = append(commits, r.convertCommit(c))

		return nil
	})

	if err != nil && !errors.Is(err, object.ErrCanceled) {
		return nil, fmt.Errorf("iterate commits: %w", err)
	}

	// Don't include the 'from' commit unless it equals 'to'
	if fromCommit.Hash != toCommit.Hash {
		commits = append(commits, r.convertCommit(fromCommit))
	}

	return commits, nil
}

// GetHeadCommits retrieves the latest N commits from HEAD.
func (r *SimpleRepository) GetHeadCommits(count int) ([]domain.Commit, error) {
	ref, err := r.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("get HEAD: %w", err)
	}

	iter, err := r.repo.Log(&gogit.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("create iterator: %w", err)
	}

	commits := make([]domain.Commit, 0, count)
	collected := 0

	err = iter.ForEach(func(c *object.Commit) error {
		if collected >= count {
			return object.ErrCanceled
		}

		commits = append(commits, r.convertCommit(c))
		collected++

		return nil
	})

	if err != nil && !errors.Is(err, object.ErrCanceled) {
		return nil, fmt.Errorf("iterate commits: %w", err)
	}

	return commits, nil
}

// convertCommit converts go-git commit to domain commit.
func (r *SimpleRepository) convertCommit(commit *object.Commit) domain.Commit {
	return domain.NewCommit(
		commit.Hash.String(),
		commit.Message,
		commit.Author.Name,
		commit.Author.Email,
		commit.Author.When.Format("2006-01-02T15:04:05Z"),
		commit.PGPSignature,
		len(commit.ParentHashes) > 1,
	)
}

// Compare with current implementation:
// - No complex helper functions
// - No context passing everywhere
// - Direct conversion to domain types
// - ~70% less code
// - Easier to test with mock git.Repository
