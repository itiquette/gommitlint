// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/pkg/errors"
)

type Repository struct {
	Repo *git.Repository
}

// CommitInfo holds information about a commit.
type CommitInfo struct {
	Message   string         // Full commit message
	Subject   string         // First line of commit message
	Body      string         // Rest of commit message after first line
	Signature string         // Signature
	RawCommit *object.Commit // Gives access to the full commit object
}

func NewRepository(path string) (*Repository, error) {
	gitDir := ".git"
	if path != "" {
		gitDir = filepath.Join(path, ".git")
	}

	dotGitDirPath, err := readDotGit(gitDir)
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(filepath.Dir(dotGitDirPath))
	if err != nil {
		return nil, err
	}

	return &Repository{Repo: repo}, nil
}

func (gitPtr *Repository) Messages(commit1, commit2 string) ([]CommitInfo, error) {
	// Case: Neither revision specified - use HEAD
	if commit1 == "" && commit2 == "" {
		// Get HEAD reference
		ref, err := gitPtr.Repo.Head()
		if err != nil {
			return nil, fmt.Errorf("failed to get repository HEAD: %w", err)
		}

		// Get commit object for HEAD
		commit, err := gitPtr.Repo.CommitObject(ref.Hash())
		if err != nil {
			return nil, fmt.Errorf("failed to get commit object: %w", err)
		}

		// Create commit info for HEAD
		subject, body := messageToSubjectAndBody(commit.Message)

		return []CommitInfo{
			{
				Message:   commit.Message,
				Subject:   subject,
				Body:      body,
				Signature: commit.PGPSignature,
				RawCommit: commit,
			},
		}, nil
	}

	// Both revisions must be provided
	if commit1 == "" || commit2 == "" {
		return nil, errors.New("both commit1 and commit2 must be provided")
	}

	// Get hash for commit2
	hash2, err := gitPtr.Repo.ResolveRevision(plumbing.Revision(commit2))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", commit2, err)
	}

	// Get hash for commit1
	hash1, err := gitPtr.Repo.ResolveRevision(plumbing.Revision(commit1))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", commit1, err)
	}

	// Set up log iterator from commit2
	commitIter, err := gitPtr.Repo.Log(&git.LogOptions{
		From:  *hash2,
		Order: git.LogOrderDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}
	defer commitIter.Close()

	// Collect commits until we reach commit1
	commitInfos := make([]CommitInfo, 0, 16)
	err = commitIter.ForEach(func(commitObject *object.Commit) error {
		if commitObject.ID() == *hash1 {
			return storer.ErrStop
		}

		subject, body := messageToSubjectAndBody(commitObject.Message)
		commitInfos = append(commitInfos, CommitInfo{
			Message:   commitObject.Message,
			Subject:   subject,
			Body:      body,
			Signature: commitObject.PGPSignature,
			RawCommit: commitObject,
		})

		return nil
	})

	if err != nil && !errors.Is(err, storer.ErrStop) {
		return nil, fmt.Errorf("error iterating commits: %w", err)
	}

	return commitInfos, nil
}

// AheadBehind returns the number of commits that HEAD is ahead and behind
// relative to the specified ref.
func (gitPtr *Repository) AheadBehind(ref string) (int, int, error) {
	ref1, err := gitPtr.Repo.Reference(plumbing.ReferenceName(ref), true)
	if err != nil {
		return 0, 0, err
	}

	ref2, err := gitPtr.Repo.Head()
	if err != nil {
		return 0, 0, err
	}

	commit1, err := gitPtr.Repo.CommitObject(ref1.Hash())
	if err != nil {
		return 0, 0, err
	}

	commit2, err := gitPtr.Repo.CommitObject(ref2.Hash())
	if err != nil {
		return 0, 0, err
	}

	mergeBase, err := commit1.MergeBase(commit2)
	if err != nil {
		return 0, 0, err
	}

	ahead, err := countCommits(gitPtr.Repo, mergeBase[0], commit2)
	if err != nil {
		return 0, 0, err
	}

	behind, err := countCommits(gitPtr.Repo, mergeBase[0], commit1)
	if err != nil {
		return 0, 0, err
	}

	return ahead, behind, nil
}

func countCommits(_ *git.Repository, from, to *object.Commit) (int, error) {
	count := 0
	iter := object.NewCommitIterCTime(to, nil, nil)
	err := iter.ForEach(func(c *object.Commit) error {
		if c.Hash == from.Hash {
			return storer.ErrStop
		}

		count++

		return nil
	})

	return count, err
}

// messageToSubjectAndBody separates a commit message into subject and body.
func messageToSubjectAndBody(message string) (string, string) {
	// Trim any trailing newlines from the full message
	message = strings.TrimRight(message, "\n")

	// Split by newline
	parts := strings.SplitN(message, "\n", 2)

	subject := parts[0]
	body := ""

	if len(parts) > 1 {
		// Trim leading newlines from body for cleaner formatting
		body = strings.TrimLeft(parts[1], "\n")
	}

	return subject, body
}

func readDotGit(name string) (string, error) {
	dotGitDir, err := os.Stat(name)
	if err != nil {
		return "", fmt.Errorf(".git directory not found: %w", err)
	}

	if !dotGitDir.IsDir() {
		return "", fmt.Errorf("%s is not a directory", name)
	}

	return filepath.Abs(name)
}
