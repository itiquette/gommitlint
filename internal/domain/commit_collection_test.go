// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain_test

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestCommitCollection_FilterMergeCommits(t *testing.T) {
	// Create test commits
	normalCommit := &domain.CommitInfo{
		Hash:          "normal123",
		Subject:       "Normal commit",
		IsMergeCommit: false,
	}

	mergeCommit := &domain.CommitInfo{
		Hash:          "merge456",
		Subject:       "Merge branch",
		IsMergeCommit: true,
	}

	// Create collection with both types
	commits := []*domain.CommitInfo{normalCommit, mergeCommit}
	collection := domain.NewCommitCollection(commits)

	// Filter merge commits
	filtered := collection.FilterMergeCommits()

	// Verify result
	assert.Equal(t, 1, filtered.Count(), "Should have filtered out merge commits")
	assert.Equal(t, normalCommit, filtered.First(), "Should only contain normal commit")
	assert.False(t, filtered.Contains(mergeCommit.Hash), "Should not contain merge commit")
}

func TestCommitCollection_Contains(t *testing.T) {
	// Create test commits
	commit1 := &domain.CommitInfo{Hash: "abc123"}
	commit2 := &domain.CommitInfo{Hash: "def456"}

	// Create collection
	collection := domain.NewCommitCollection([]*domain.CommitInfo{commit1})

	// Test contains
	assert.True(t, collection.Contains(commit1.Hash), "Should contain commit1")
	assert.False(t, collection.Contains(commit2.Hash), "Should not contain commit2")
	assert.False(t, collection.Contains("nonexistent"), "Should not contain nonexistent commit")
}

func TestCommitCollection_FirstLast(t *testing.T) {
	// Create test commits
	commit1 := &domain.CommitInfo{Hash: "abc123"}
	commit2 := &domain.CommitInfo{Hash: "def456"}
	commit3 := &domain.CommitInfo{Hash: "ghi789"}

	// Test with empty collection
	emptyCollection := domain.NewCommitCollection([]*domain.CommitInfo{})
	assert.Nil(t, emptyCollection.First(), "First() should return nil for empty collection")
	assert.Nil(t, emptyCollection.Last(), "Last() should return nil for empty collection")

	// Test with populated collection
	collection := domain.NewCommitCollection([]*domain.CommitInfo{commit1, commit2, commit3})
	assert.Equal(t, commit1, collection.First(), "First() should return first commit")
	assert.Equal(t, commit3, collection.Last(), "Last() should return last commit")
}

func TestCommitCollection_AddAndAddAll(t *testing.T) {
	// Create test commits
	commit1 := &domain.CommitInfo{Hash: "abc123"}
	commit2 := &domain.CommitInfo{Hash: "def456"}
	commit3 := &domain.CommitInfo{Hash: "ghi789"}

	// Test Add
	collection1 := domain.NewCommitCollection([]*domain.CommitInfo{commit1})
	collection1.Add(commit2)
	assert.Equal(t, 2, collection1.Count(), "Count should be 2 after adding a commit")
	assert.True(t, collection1.Contains(commit2.Hash), "Should contain added commit")

	// Test AddAll
	collection2 := domain.NewCommitCollection([]*domain.CommitInfo{commit3})
	collection1.AddAll(collection2)
	assert.Equal(t, 3, collection1.Count(), "Count should be 3 after adding all commits from another collection")
	assert.True(t, collection1.Contains(commit3.Hash), "Should contain commit from added collection")
}

func TestCommitCollection_FilterByAuthor(t *testing.T) {
	// Create author
	author := &object.Signature{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	otherAuthor := &object.Signature{
		Name:  "Jane Smith",
		Email: "jane@example.com",
	}

	// Create test commits with authors
	commit1 := &domain.CommitInfo{
		Hash: "abc123",
		RawCommit: &object.Commit{
			Author: *author,
		},
	}

	commit2 := &domain.CommitInfo{
		Hash: "def456",
		RawCommit: &object.Commit{
			Author: *otherAuthor,
		},
	}

	commit3 := &domain.CommitInfo{
		Hash: "ghi789",
		RawCommit: &object.Commit{
			Author: *author,
		},
	}

	// Create collection
	collection := domain.NewCommitCollection([]*domain.CommitInfo{commit1, commit2, commit3})

	// Filter by author email
	filtered := collection.FilterByAuthor(author.Email)
	assert.Equal(t, 2, filtered.Count(), "Should have filtered to author's commits only")
	assert.True(t, filtered.Contains(commit1.Hash), "Should contain commit1")
	assert.False(t, filtered.Contains(commit2.Hash), "Should not contain commit2")
	assert.True(t, filtered.Contains(commit3.Hash), "Should contain commit3")

	// Filter by author name
	filtered = collection.FilterByAuthor(author.Name)
	assert.Equal(t, 2, filtered.Count(), "Should have filtered to author's commits only")
}

func TestCommitCollection_IsEmpty(t *testing.T) {
	// Test empty collection
	emptyCollection := domain.NewCommitCollection([]*domain.CommitInfo{})
	assert.True(t, emptyCollection.IsEmpty(), "Empty collection should be empty")

	// Test non-empty collection
	commit := &domain.CommitInfo{Hash: "abc123"}
	nonEmptyCollection := domain.NewCommitCollection([]*domain.CommitInfo{commit})
	assert.False(t, nonEmptyCollection.IsEmpty(), "Non-empty collection should not be empty")
}
