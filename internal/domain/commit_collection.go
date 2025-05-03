// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package domain

import "github.com/itiquette/gommitlint/internal/contextx"

// CommitCollection represents a collection of commits with common operations.
// This is a pure domain entity with value semantics.
type CommitCollection struct {
	commits []CommitInfo
}

// NewCommitCollection creates a new CommitCollection from a slice of commits.
func NewCommitCollection(commits []CommitInfo) CommitCollection {
	return CommitCollection{
		commits: contextx.DeepCopy(commits),
	}
}

// FilterMergeCommits returns a new collection with merge commits filtered out.
func (c CommitCollection) FilterMergeCommits() CommitCollection {
	filtered := contextx.Filter(c.commits, func(commit CommitInfo) bool {
		return !commit.IsMergeCommit
	})

	return NewCommitCollection(filtered)
}

// FilterByAuthor returns a new collection with commits by the specified author or email.
// This uses the AuthorName and AuthorEmail fields of CommitInfo rather than accessing
// the raw Git commit directly, maintaining domain separation.
func (c CommitCollection) FilterByAuthor(authorNameOrEmail string) CommitCollection {
	filtered := contextx.Filter(c.commits, func(commit CommitInfo) bool {
		return commit.AuthorName == authorNameOrEmail || commit.AuthorEmail == authorNameOrEmail
	})

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

// All returns all commits in the collection as a new slice.
// This returns a copy to maintain value semantics.
func (c CommitCollection) All() []CommitInfo {
	return contextx.DeepCopy(c.commits)
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
	return contextx.Some(c.commits, func(commit CommitInfo) bool {
		return commit.Hash == hash
	})
}

// Add adds a commit to the collection and returns the updated collection.
// This creates a new collection rather than modifying the existing one,
// following value semantics.
func (c CommitCollection) Add(commit CommitInfo) CommitCollection {
	// We can't use the generic DeepCopy here directly since we need to append
	newCommits := contextx.DeepCopy(c.commits)
	newCommits = append(newCommits, commit)

	return CommitCollection{
		commits: newCommits,
	}
}

// AddAll adds all commits from another collection to this collection.
// This creates a new collection rather than modifying the existing one,
// following value semantics.
func (c CommitCollection) AddAll(other CommitCollection) CommitCollection {
	// We need to append all items from other collection
	newCommits := contextx.DeepCopy(c.commits)
	newCommits = append(newCommits, contextx.DeepCopy(other.commits)...)

	return CommitCollection{
		commits: newCommits,
	}
}
