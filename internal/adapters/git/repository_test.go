// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package git_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/adapters/git"
	"github.com/itiquette/gommitlint/internal/domain"
)

// TestGetCommitRange tests the GetCommitRange functionality with various scenarios.
func TestGetCommitRange(t *testing.T) {
	tests := []struct {
		name          string
		setupRepo     func(t *testing.T, repo *gogit.Repository) (from, to string)
		expectedCount int
		expectedError string
		checkCommits  func(t *testing.T, commits []domain.Commit)
	}{
		{
			name: "Linear history - commits in range",
			setupRepo: func(t *testing.T, repo *gogit.Repository) (string, string) {
				t.Helper()
				// Create a linear history: A -> B -> C -> D
				hashA := createCommit(t, repo, "Initial commit", nil)
				hashB := createCommit(t, repo, "Second commit", []plumbing.Hash{hashA})
				hashC := createCommit(t, repo, "Third commit", []plumbing.Hash{hashB})
				hashD := createCommit(t, repo, "Fourth commit", []plumbing.Hash{hashC})

				// Return range from B to D (should include C and D, but not B)
				return hashB.String(), hashD.String()
			},
			expectedCount: 2, // Commits C and D (B is excluded as it's the base)
		},
		{
			name: "Diverged branches - feature branch ahead of main",
			setupRepo: func(t *testing.T, repo *gogit.Repository) (string, string) {
				t.Helper()
				// Create diverged history:
				//   A -> B -> C (main)
				//    \-> D -> E (feature)
				hashA := createCommit(t, repo, "Initial commit", nil)
				hashB := createCommit(t, repo, "Main commit 1", []plumbing.Hash{hashA})
				hashC := createCommit(t, repo, "Main commit 2", []plumbing.Hash{hashB})

				// Create feature branch from A
				hashD := createCommit(t, repo, "Feature commit 1", []plumbing.Hash{hashA})
				hashE := createCommit(t, repo, "Feature commit 2", []plumbing.Hash{hashD})

				// Create main branch ref
				mainRef := plumbing.NewHashReference("refs/heads/main", hashC)
				err := repo.Storer.SetReference(mainRef)
				require.NoError(t, err)

				// Return range from main to feature (should include D and E)
				return "main", hashE.String()
			},
			expectedCount: 2, // Only commits D and E (not in main)
			checkCommits: func(t *testing.T, commits []domain.Commit) {
				t.Helper()
				// Verify we got the feature commits
				subjects := make([]string, len(commits))
				for i, c := range commits {
					subjects[i] = c.Subject
				}
				require.Contains(t, subjects, "Feature commit 1")
				require.Contains(t, subjects, "Feature commit 2")
				require.NotContains(t, subjects, "Main commit 1")
				require.NotContains(t, subjects, "Main commit 2")
			},
		},
		{
			name: "Same commit for from and to",
			setupRepo: func(t *testing.T, repo *gogit.Repository) (string, string) {
				t.Helper()
				hashA := createCommit(t, repo, "Only commit", nil)

				return hashA.String(), hashA.String()
			},
			expectedCount: 0, // No commits between A and A
		},
		{
			name: "Non-existent from reference",
			setupRepo: func(t *testing.T, repo *gogit.Repository) (string, string) {
				t.Helper()
				hashA := createCommit(t, repo, "Only commit", nil)

				return "nonexistent", hashA.String()
			},
			expectedError: "failed to resolve 'from' reference",
		},
		{
			name: "Non-existent to reference",
			setupRepo: func(t *testing.T, repo *gogit.Repository) (string, string) {
				t.Helper()
				hashA := createCommit(t, repo, "Only commit", nil)

				return hashA.String(), "nonexistent"
			},
			expectedError: "failed to resolve 'to' reference",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create temporary directory for test repository
			tmpDir, err := os.MkdirTemp("", "gommitlint-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Initialize repository
			repo, err := gogit.PlainInit(tmpDir, false)
			require.NoError(t, err)

			// Setup test scenario
			fromRef, toRef := testCase.setupRepo(t, repo)

			// Create repository adapter
			adapter, err := git.NewRepository(tmpDir)
			require.NoError(t, err)

			repoAdapter := adapter

			// Execute GetCommitRange
			commits, err := repoAdapter.GetCommitRange(context.Background(), fromRef, toRef)

			// Check error expectations
			if testCase.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			// Check success expectations
			require.NoError(t, err)
			require.Len(t, commits, testCase.expectedCount)

			// Run custom checks if provided
			if testCase.checkCommits != nil {
				testCase.checkCommits(t, commits)
			}
		})
	}
}

// createCommit is a helper function to create a commit in the test repository.
func createCommit(t *testing.T, repo *gogit.Repository, message string, parents []plumbing.Hash) plumbing.Hash {
	t.Helper()

	// Get or create worktree
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	// Create a file to commit
	filename := filepath.Join(worktree.Filesystem.Root(), message+".txt")
	err = os.WriteFile(filename, []byte(message), 0600)
	require.NoError(t, err)

	// Add file to staging
	_, err = worktree.Add(message + ".txt")
	require.NoError(t, err)

	// Create commit
	commitOpts := &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
		Parents: parents,
	}

	hash, err := worktree.Commit(message, commitOpts)
	require.NoError(t, err)

	return hash
}

// TestGetCommitRangeWithMergeCommits tests handling of merge commits in ranges.
func TestGetCommitRangeWithMergeCommits(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "gommitlint-merge-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Initialize repository
	repo, err := gogit.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Create history with merge:
	//   A -> B -> D (merge) -> E
	//    \-> C ---/
	hashA := createCommit(t, repo, "Initial commit", nil)
	hashB := createCommit(t, repo, "Main branch commit", []plumbing.Hash{hashA})
	hashC := createCommit(t, repo, "Feature branch commit", []plumbing.Hash{hashA})
	hashD := createCommit(t, repo, "Merge commit", []plumbing.Hash{hashB, hashC})
	hashE := createCommit(t, repo, "Post-merge commit", []plumbing.Hash{hashD})

	// Create repository adapter
	adapter, err := git.NewRepository(tmpDir)
	require.NoError(t, err)

	repoAdapter := adapter

	// Test range from A to E (includes B, C, D, E - but not A)
	commits, err := repoAdapter.GetCommitRange(context.Background(), hashA.String(), hashE.String())
	require.NoError(t, err)
	require.Len(t, commits, 4) // B, C, D, E (A is excluded as it's the base)

	// Verify merge commit is included
	var foundMerge bool

	for _, commit := range commits {
		if commit.Subject == "Merge commit" {
			require.True(t, commit.IsMergeCommit)

			foundMerge = true
		}
	}

	require.True(t, foundMerge, "Merge commit should be included in range")
}
