// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"errors"
	"fmt"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/itiquette/gommitlint/internal/domain"
)

// Repository implements the CommitRepository port.
type Repository struct {
	repo *gogit.Repository
}

// NewRepository opens a git repository at the given path.
func NewRepository(path string) (*Repository, error) {
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("open repository: %w", err)
	}

	return &Repository{repo: repo}, nil
}

// GetCommit retrieves a single commit by hash or reference.
func (r *Repository) GetCommit(_ context.Context, ref string) (domain.Commit, error) {
	// Try to resolve as a reference first (handles HEAD, branch names, etc.)
	hash, err := r.resolveReference(ref)
	if err != nil {
		// If reference resolution fails, try as a direct hash
		hash = plumbing.NewHash(ref)
	}

	commit, err := r.repo.CommitObject(hash)
	if err != nil {
		return domain.Commit{}, fmt.Errorf("get commit: %w", err)
	}

	return r.convertCommit(commit), nil
}

// resolveReference resolves a reference (like HEAD, branch name) to a commit hash.
func (r *Repository) resolveReference(ref string) (plumbing.Hash, error) {
	// Handle HEAD specially
	if ref == "HEAD" {
		headRef, err := r.repo.Head()
		if err != nil {
			return plumbing.ZeroHash, err
		}

		return headRef.Hash(), nil
	}

	// Try common reference formats
	refFormats := []string{
		ref,                          // Direct reference name
		"refs/heads/" + ref,          // Branch reference
		"refs/remotes/origin/" + ref, // Remote branch reference
		"refs/tags/" + ref,           // Tag reference
	}

	for _, refName := range refFormats {
		resolvedRef, err := r.repo.Reference(plumbing.ReferenceName(refName), true)
		if err == nil {
			return resolvedRef.Hash(), nil
		}
	}

	return plumbing.ZeroHash, fmt.Errorf("reference not found: %s", ref)
}

// GetCommitRange retrieves commits in a range (from..to).
// Returns all commits reachable from 'to' but not reachable from 'from'.
func (r *Repository) GetCommitRange(_ context.Context, fromRef, toRef string) ([]domain.Commit, error) {
	// Resolve references to hashes
	fromHash, err := r.resolveReference(fromRef)
	if err != nil {
		// If reference resolution fails, try as a direct hash
		fromHash = plumbing.NewHash(fromRef)
	}

	toHash, err := r.resolveReference(toRef)
	if err != nil {
		// If reference resolution fails, try as a direct hash
		toHash = plumbing.NewHash(toRef)
	}

	// Validate that both commits exist
	_, err = r.repo.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve 'from' reference: %w", err)
	}

	_, err = r.repo.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve 'to' reference: %w", err)
	}

	// Get all commits reachable from 'to'
	reachableFromTo := make(map[plumbing.Hash]bool)

	err = r.collectReachableCommits(toHash, reachableFromTo)
	if err != nil {
		return nil, fmt.Errorf("collect commits reachable from 'to': %w", err)
	}

	// Get all commits reachable from 'from'
	reachableFromFrom := make(map[plumbing.Hash]bool)

	err = r.collectReachableCommits(fromHash, reachableFromFrom)
	if err != nil {
		return nil, fmt.Errorf("collect commits reachable from 'from': %w", err)
	}

	// Find commits in range: reachable from 'to' but not from 'from'
	var commits []domain.Commit

	for hash := range reachableFromTo {
		if !reachableFromFrom[hash] {
			commit, err := r.repo.CommitObject(hash)
			if err != nil {
				return nil, fmt.Errorf("get commit object: %w", err)
			}

			commits = append(commits, r.convertCommit(commit))
		}
	}

	return commits, nil
}

// collectReachableCommits recursively collects all commits reachable from the given hash.
func (r *Repository) collectReachableCommits(hash plumbing.Hash, reachable map[plumbing.Hash]bool) error {
	// Avoid cycles
	if reachable[hash] {
		return nil
	}

	reachable[hash] = true

	commit, err := r.repo.CommitObject(hash)
	if err != nil {
		return err
	}

	// Recursively collect from all parents
	for _, parentHash := range commit.ParentHashes {
		err = r.collectReachableCommits(parentHash, reachable)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetHeadCommits retrieves the latest N commits from HEAD.
func (r *Repository) GetHeadCommits(_ context.Context, count int) ([]domain.Commit, error) {
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

// GetCommitsAheadCount returns how many commits the current branch is ahead of the reference.
func (r *Repository) GetCommitsAheadCount(_ context.Context, referenceBranch string) (int, error) {
	head, err := r.repo.Head()
	if err != nil {
		return 0, fmt.Errorf("get HEAD: %w", err)
	}

	// Try different reference formats to find the target branch
	refFormats := []string{
		"refs/remotes/origin/" + referenceBranch, // Remote branch
		"refs/heads/" + referenceBranch,          // Local branch
		"refs/remotes/" + referenceBranch,        // Legacy format
	}

	var refHash plumbing.Hash

	found := false

	for _, refName := range refFormats {
		refCommit, err := r.repo.Reference(plumbing.ReferenceName(refName), true)
		if err == nil {
			refHash = refCommit.Hash()
			found = true

			break
		}
	}

	if !found {
		// Reference doesn't exist, return 0 (not ahead)
		return 0, nil
	}

	// Count commits between reference and HEAD
	iter, err := r.repo.Log(&gogit.LogOptions{From: head.Hash()})
	if err != nil {
		return 0, fmt.Errorf("get log: %w", err)
	}
	defer iter.Close()

	count := 0

	err = iter.ForEach(func(commit *object.Commit) error {
		if commit.Hash == refHash {
			return errors.New("found reference") // Stop iteration
		}

		count++

		return nil
	})

	if err != nil && err.Error() != "found reference" {
		return 0, fmt.Errorf("count commits: %w", err)
	}

	return count, nil
}

// convertCommit converts go-git commit to domain commit.
func (r *Repository) convertCommit(commit *object.Commit) domain.Commit {
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
