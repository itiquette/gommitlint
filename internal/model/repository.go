// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package model

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// ErrRevisionRange is returned when both revisions aren't provided for a range.
var ErrRevisionRange = errors.New("both rev1 and rev2 must be provided for a range")

// Repository wraps a git.Repository with additional functionality.
type Repository struct {
	Repo *git.Repository
}

// NewRepository opens a git repository at the specified path.
// If path is empty, it looks for a repository in the current directory.
func NewRepository(path string) (*Repository, error) {
	repoPath, err := findGitDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to find git directory: %w", err)
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &Repository{Repo: repo}, nil
}

// findGitDir locates the .git directory for a repository.
func findGitDir(path string) (string, error) {
	if path == "" {
		path = "."
	}

	// Check if path is a .git directory
	gitPath := filepath.Join(path, ".git")

	fi, err := os.Stat(gitPath)
	if err != nil {
		return "", fmt.Errorf(".git directory not found at %s: %w", path, err)
	}

	if !fi.IsDir() {
		return "", fmt.Errorf("%s is not a directory", gitPath)
	}

	// Return the parent directory of .git
	return filepath.Abs(path)
}

// GetHeadCommit returns information about the HEAD commit.
func (r *Repository) GetHeadCommit() (*CommitInfo, error) {
	ref, err := r.Repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository HEAD: %w", err)
	}

	commit, err := r.Repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	subject, body := SplitCommitMessage(commit.Message)

	return &CommitInfo{
		Message:   commit.Message,
		Subject:   subject,
		Body:      body,
		Signature: commit.PGPSignature,
		RawCommit: commit,
	}, nil
}

// CommitInfos retrieves commit information between two revisions.
// If both revisions are empty, it returns only the HEAD commit.
// If one revision is provided but the other is empty, it returns ErrRevisionRange.
func (r *Repository) CommitInfos(rev1, rev2 string) ([]CommitInfo, error) {
	// Case: Neither revision specified - use HEAD only
	if rev1 == "" && rev2 == "" {
		headCommit, err := r.GetHeadCommit()
		if err != nil {
			return nil, err
		}

		return []CommitInfo{*headCommit}, nil
	}

	// Both revisions must be provided for a range
	if rev1 == "" || rev2 == "" {
		return nil, ErrRevisionRange
	}

	return r.getCommitRange(rev1, rev2)
}

// getCommitRange returns commits between rev1 and rev2 (exclusive of rev1).
// Format is similar to git log rev1..rev2.
func (r *Repository) getCommitRange(rev1, rev2 string) ([]CommitInfo, error) {
	// Resolve revisions to hashes
	hash1, err := r.Repo.ResolveRevision(plumbing.Revision(rev1))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", rev1, err)
	}

	hash2, err := r.Repo.ResolveRevision(plumbing.Revision(rev2))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", rev2, err)
	}

	// Create log iterator starting from hash2
	commitIter, err := r.Repo.Log(&git.LogOptions{
		From:  *hash2,
		Order: git.LogOrderDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}
	defer commitIter.Close()

	// Collect commits until we reach hash1
	commits := make([]CommitInfo, 0, 16)
	err = commitIter.ForEach(func(commitObject *object.Commit) error {
		// Stop when we reach the starting commit
		if commitObject.Hash == *hash1 {
			return storer.ErrStop
		}

		subject, body := SplitCommitMessage(commitObject.Message)
		commits = append(commits, CommitInfo{
			Message:   commitObject.Message,
			Subject:   subject,
			Body:      body,
			Signature: commitObject.PGPSignature,
			RawCommit: commitObject,
		})

		return nil
	})

	// ErrStop is expected when reaching the starting commit
	if err != nil && !errors.Is(err, storer.ErrStop) {
		return nil, fmt.Errorf("error iterating commits: %w", err)
	}

	return commits, nil
}

// SplitCommitMessage separates a commit message into subject and body.
// It returns the first line as subject and the rest (if any) as body.
func SplitCommitMessage(message string) (string, string) {
	// Trim trailing newlines
	message = strings.TrimRight(message, "\n")

	// Split by newline
	parts := strings.SplitN(message, "\n", 2)
	subject := parts[0]
	body := ""

	// Extract body if it exists
	if len(parts) > 1 {
		body = strings.TrimLeft(parts[1], "\n")
	}

	return subject, body
}
