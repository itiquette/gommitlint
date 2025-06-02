// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package git provides Git repository adapters for the domain model.
package git

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/itiquette/gommitlint/internal/domain"
)

// Repository provides Git operations for the domain.
// Simple, functional implementation with value semantics.
type Repository struct {
	repo *git.Repository // Only pointer because go-git requires it
	path string
}

// NewRepository opens a Git repository at the given path.
// Returns a value, not a pointer, following functional principles.
func NewRepository(_ context.Context, path string) (Repository, error) {
	if path == "" {
		var err error

		path, err = os.Getwd()
		if err != nil {
			return Repository{}, fmt.Errorf("get current directory: %w", err)
		}
	}

	// Find git directory by walking up the tree
	gitPath := findGitDir(path)
	if gitPath == "" {
		return Repository{}, fmt.Errorf("not a git repository: %s", path)
	}

	repo, err := git.PlainOpen(gitPath)
	if err != nil {
		return Repository{}, fmt.Errorf("open repository: %w", err)
	}

	return Repository{repo: repo, path: gitPath}, nil
}

// GetCommit retrieves a single commit by hash or reference.
func (r Repository) GetCommit(_ context.Context, ref string) (domain.Commit, error) {
	hash, err := r.repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return domain.Commit{}, fmt.Errorf("resolve reference %s: %w", ref, err)
	}

	commit, err := r.repo.CommitObject(*hash)
	if err != nil {
		return domain.Commit{}, fmt.Errorf("get commit: %w", err)
	}

	return convertCommit(commit), nil
}

// GetCommits retrieves the last n commits from HEAD.
func (r Repository) GetCommits(_ context.Context, count int) ([]domain.Commit, error) {
	head, err := r.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("get HEAD: %w", err)
	}

	return getCommitsFrom(r.repo, head.Hash(), count)
}

// GetCommitRange retrieves commits between two references (inclusive).
func (r Repository) GetCommitRange(_ context.Context, from, toRef string) ([]domain.Commit, error) {
	fromHash, err := r.repo.ResolveRevision(plumbing.Revision(from))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve 'from' reference '%s': %w", from, err)
	}

	toHash, err := r.repo.ResolveRevision(plumbing.Revision(toRef))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve 'to' reference '%s': %w", toRef, err)
	}

	return getCommitsBetween(r.repo, *fromHash, *toHash)
}

// GetHeadCommits retrieves n commits from HEAD.
func (r Repository) GetHeadCommits(ctx context.Context, count int) ([]domain.Commit, error) {
	return r.GetCommits(ctx, count)
}

// GetCurrentBranch returns the current branch name.
func (r Repository) GetCurrentBranch(_ context.Context) (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("get HEAD: %w", err)
	}

	if !head.Name().IsBranch() {
		return "", errors.New("HEAD is detached")
	}

	return head.Name().Short(), nil
}

// GetCommitsAhead counts commits ahead of a reference.
func (r Repository) GetCommitsAhead(_ context.Context, ref string) (int, error) {
	head, err := r.repo.Head()
	if err != nil {
		return 0, fmt.Errorf("get HEAD: %w", err)
	}

	refHash, err := r.repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return 0, fmt.Errorf("resolve reference: %w", err)
	}

	return countCommitsBetween(r.repo, head.Hash(), *refHash)
}

// IsValid checks if this is a valid repository.
func (r Repository) IsValid(_ context.Context) (bool, error) {
	_, err := r.repo.Head()

	return err == nil, nil
}

// GetRepositoryName returns the repository name.
func (r Repository) GetRepositoryName(_ context.Context) string {
	return filepath.Base(r.path)
}

// Pure helper functions - no receivers, functional style

func findGitDir(path string) string {
	for {
		gitPath := filepath.Join(path, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return path
		}

		parent := filepath.Dir(path)
		if parent == path {
			return ""
		}

		path = parent
	}
}

func getCommitsFrom(repo *git.Repository, start plumbing.Hash, limit int) ([]domain.Commit, error) {
	iter, err := repo.Log(&git.LogOptions{From: start})
	if err != nil {
		return nil, fmt.Errorf("create iterator: %w", err)
	}
	defer iter.Close()

	return collectCommits(iter, limit)
}

func getCommitsBetween(repo *git.Repository, from, toHash plumbing.Hash) ([]domain.Commit, error) {
	// If from and to are the same, return empty (no commits between)
	if from == toHash {
		return []domain.Commit{}, nil
	}

	// For all cases (linear and non-linear), use the approach of finding
	// commits reachable from 'to' but not from 'from'
	// This properly handles merge commits and complex histories
	return getCommitsInBranchNotInBase(repo, from, toHash)
}

func countCommitsBetween(repo *git.Repository, from, toHash plumbing.Hash) (int, error) {
	iter, err := repo.Log(&git.LogOptions{From: from})
	if err != nil {
		return 0, fmt.Errorf("create iterator: %w", err)
	}
	defer iter.Close()

	return countUntil(iter, toHash)
}

// Pure collection functions - functional style

func collectCommits(iter object.CommitIter, limit int) ([]domain.Commit, error) {
	commits := make([]domain.Commit, 0, limit)

	err := iter.ForEach(func(c *object.Commit) error {
		if limit > 0 && len(commits) >= limit {
			return io.EOF
		}

		commits = append(commits, convertCommit(c))

		return nil
	})

	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return commits, nil
}

func countUntil(iter object.CommitIter, stop plumbing.Hash) (int, error) {
	count := 0

	err := iter.ForEach(func(c *object.Commit) error {
		if c.Hash == stop {
			return io.EOF
		}

		count++

		return nil
	})

	if err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}

	return count, nil
}

// Pure conversion function.
func convertCommit(commit *object.Commit) domain.Commit {
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

// Helper functions for commit range logic

func getCommitsInBranchNotInBase(repo *git.Repository, base, branch plumbing.Hash) ([]domain.Commit, error) {
	// Get all commits from branch
	branchIter, err := repo.Log(&git.LogOptions{From: branch})
	if err != nil {
		return nil, fmt.Errorf("create branch iterator: %w", err)
	}
	defer branchIter.Close()

	// Get all commits from base
	baseCommits := make(map[plumbing.Hash]bool)

	baseIter, err := repo.Log(&git.LogOptions{From: base})
	if err != nil {
		return nil, fmt.Errorf("create base iterator: %w", err)
	}

	defer baseIter.Close()

	// Collect all base commits
	err = baseIter.ForEach(func(c *object.Commit) error {
		baseCommits[c.Hash] = true

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect base commits: %w", err)
	}

	// Collect branch commits not in base
	var commits []domain.Commit

	err = branchIter.ForEach(func(c *object.Commit) error {
		if !baseCommits[c.Hash] {
			commits = append(commits, convertCommit(c))
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect branch commits: %w", err)
	}

	return commits, nil
}
