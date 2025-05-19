// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

// CommitCollection provides functional operations on commit slices.
type CommitCollection []CommitInfo

// NewCommitCollection creates a new CommitCollection from a slice of commits.
func NewCommitCollection(commits []CommitInfo) CommitCollection {
	result := make(CommitCollection, len(commits))
	copy(result, commits)

	return result
}

// Filter returns a new collection with commits matching the predicate.
func (c CommitCollection) Filter(predicate func(CommitInfo) bool) CommitCollection {
	result := make(CommitCollection, 0)

	for _, commit := range c {
		if predicate(commit) {
			result = append(result, commit)
		}
	}

	return result
}

// FilterMergeCommits returns a new collection with merge commits filtered out.
func (c CommitCollection) FilterMergeCommits() CommitCollection {
	return c.Filter(func(commit CommitInfo) bool {
		return !commit.IsMergeCommit
	})
}

// FilterByAuthor returns a new collection with commits by the specified author.
func (c CommitCollection) FilterByAuthor(authorNameOrEmail string) CommitCollection {
	return c.Filter(func(commit CommitInfo) bool {
		return commit.AuthorName == authorNameOrEmail || commit.AuthorEmail == authorNameOrEmail
	})
}

// Map transforms commits using the provided function.
func (c CommitCollection) Map(fn func(CommitInfo) CommitInfo) CommitCollection {
	result := make(CommitCollection, len(c))
	for i, commit := range c {
		result[i] = fn(commit)
	}

	return result
}

// Any returns true if any commit matches the predicate.
func (c CommitCollection) Any(predicate func(CommitInfo) bool) bool {
	for _, commit := range c {
		if predicate(commit) {
			return true
		}
	}

	return false
}

// All returns true if all commits match the predicate.
func (c CommitCollection) All(predicate func(CommitInfo) bool) bool {
	for _, commit := range c {
		if !predicate(commit) {
			return false
		}
	}

	return true
}

// First returns the first commit or empty CommitInfo if collection is empty.
func (c CommitCollection) First() CommitInfo {
	if len(c) == 0 {
		return CommitInfo{}
	}

	return c[0]
}

// Last returns the last commit or empty CommitInfo if collection is empty.
func (c CommitCollection) Last() CommitInfo {
	if len(c) == 0 {
		return CommitInfo{}
	}

	return c[len(c)-1]
}

// Count returns the number of commits in the collection.
func (c CommitCollection) Count() int {
	return len(c)
}

// IsEmpty returns true if the collection is empty.
func (c CommitCollection) IsEmpty() bool {
	return len(c) == 0
}

// Contains returns true if the collection contains a commit with the specified hash.
func (c CommitCollection) Contains(hash string) bool {
	return c.Any(func(commit CommitInfo) bool {
		return commit.Hash == hash
	})
}

// With returns a new collection with the commit added.
func (c CommitCollection) With(commit CommitInfo) CommitCollection {
	result := make(CommitCollection, len(c), len(c)+1)
	copy(result, c)

	return append(result, commit)
}

// WithAll returns a new collection with all commits from other added.
func (c CommitCollection) WithAll(other CommitCollection) CommitCollection {
	result := make(CommitCollection, len(c), len(c)+len(other))
	copy(result, c)

	return append(result, other...)
}
