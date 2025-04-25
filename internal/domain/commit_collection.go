// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package domain

import (
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CommitCollection represents a collection of commits with common operations.
type CommitCollection struct {
	commits []CommitInfo
}

// NewCommitCollection creates a new CommitCollection from a slice of commits.
func NewCommitCollection(commits []CommitInfo) CommitCollection {
	return CommitCollection{
		commits: commits,
	}
}

// FilterMergeCommits returns a new collection with merge commits filtered out.
func (c CommitCollection) FilterMergeCommits() CommitCollection {
	filtered := make([]CommitInfo, 0, len(c.commits))

	for _, commit := range c.commits {
		if !commit.IsMergeCommit {
			filtered = append(filtered, commit)
		}
	}

	return NewCommitCollection(filtered)
}

// FilterByAuthor returns a new collection with commits by the specified author.
func (c CommitCollection) FilterByAuthor(author string) CommitCollection {
	filtered := make([]CommitInfo, 0, len(c.commits))

	for _, commit := range c.commits {
		if commit.RawCommit != nil {
			// Use type assertion to check if we have a go-git commit
			if gitCommit, ok := commit.RawCommit.(*object.Commit); ok {
				if gitCommit.Author.Email == author || gitCommit.Author.Name == author {
					filtered = append(filtered, commit)
				}
			}
		}
	}

	return NewCommitCollection(filtered)
}

// First returns the first commit in the collection or an empty CommitInfo if empty.
func (c CommitCollection) First() CommitInfo {
	if len(c.commits) == 0 {
		return CommitInfo{}
	}

	return c.commits[0]
}

// Last returns the last commit in the collection or an empty CommitInfo if empty.
func (c CommitCollection) Last() CommitInfo {
	if len(c.commits) == 0 {
		return CommitInfo{}
	}

	return c.commits[len(c.commits)-1]
}

// All returns all commits in the collection.
func (c CommitCollection) All() []CommitInfo {
	return c.commits
}

// Count returns the number of commits in the collection.
func (c CommitCollection) Count() int {
	return len(c.commits)
}

// IsEmpty returns true if the collection is empty.
func (c CommitCollection) IsEmpty() bool {
	return len(c.commits) == 0
}

// Contains returns true if the collection contains a commit with the specified hash.
func (c CommitCollection) Contains(hash string) bool {
	for _, commit := range c.commits {
		if commit.Hash == hash {
			return true
		}
	}

	return false
}

// Add adds a commit to the collection and returns the updated collection.
func (c CommitCollection) Add(commit CommitInfo) CommitCollection {
	c.commits = append(c.commits, commit)

	return c
}

// AddAll adds all commits from another collection to this collection.
func (c CommitCollection) AddAll(other CommitCollection) CommitCollection {
	c.commits = append(c.commits, other.commits...)

	return c
}
