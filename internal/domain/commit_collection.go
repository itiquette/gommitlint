// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package domain

// CommitCollection represents a collection of commits with common operations.
// This is a pure domain entity with value semantics.
type CommitCollection struct {
	commits []CommitInfo
}

// NewCommitCollection creates a new CommitCollection from a slice of commits.
func NewCommitCollection(commits []CommitInfo) CommitCollection {
	// Create a defensive copy to ensure value semantics
	copiedCommits := make([]CommitInfo, len(commits))
	copy(copiedCommits, commits)

	return CommitCollection{
		commits: copiedCommits,
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

// FilterByAuthor returns a new collection with commits by the specified author or email.
// This uses the AuthorName and AuthorEmail fields of CommitInfo rather than accessing
// the raw Git commit directly, maintaining domain separation.
func (c CommitCollection) FilterByAuthor(authorNameOrEmail string) CommitCollection {
	filtered := make([]CommitInfo, 0, len(c.commits))

	for _, commit := range c.commits {
		if commit.AuthorName == authorNameOrEmail || commit.AuthorEmail == authorNameOrEmail {
			filtered = append(filtered, commit)
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

// All returns all commits in the collection as a new slice.
// This returns a copy to maintain value semantics.
func (c CommitCollection) All() []CommitInfo {
	result := make([]CommitInfo, len(c.commits))
	copy(result, c.commits)

	return result
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
// This creates a new collection rather than modifying the existing one,
// following value semantics.
func (c CommitCollection) Add(commit CommitInfo) CommitCollection {
	newCommits := make([]CommitInfo, len(c.commits)+1)
	copy(newCommits, c.commits)
	newCommits[len(c.commits)] = commit

	return CommitCollection{
		commits: newCommits,
	}
}

// AddAll adds all commits from another collection to this collection.
// This creates a new collection rather than modifying the existing one,
// following value semantics.
func (c CommitCollection) AddAll(other CommitCollection) CommitCollection {
	newCommits := make([]CommitInfo, len(c.commits)+len(other.commits))
	copy(newCommits, c.commits)
	copy(newCommits[len(c.commits):], other.commits)

	return CommitCollection{
		commits: newCommits,
	}
}
