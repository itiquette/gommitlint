// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

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
)

// TestBranchComparisonGetCommitRange tests that GetCommitRange correctly handles diverged branches.
func TestBranchComparisonGetCommitRange(t *testing.T) {
	tests := []struct {
		name            string
		setupRepo       func(t *testing.T, repo *gogit.Repository) (baseRef, targetRef string)
		expectedCommits []string // Expected commit subjects
	}{
		{
			name: "Feature branch ahead of main",
			setupRepo: func(t *testing.T, repo *gogit.Repository) (string, string) {
				t.Helper()
				// Create base commits on main
				hashA := createTestCommit(t, repo, "Initial commit", nil)
				hashB := createTestCommit(t, repo, "Main commit 1", []plumbing.Hash{hashA})

				// Create main branch ref
				mainRef := plumbing.NewHashReference("refs/heads/main", hashB)
				err := repo.Storer.SetReference(mainRef)
				require.NoError(t, err)

				// Create feature branch commits
				hashC := createTestCommit(t, repo, "Feature commit 1", []plumbing.Hash{hashB})
				hashD := createTestCommit(t, repo, "Feature commit 2", []plumbing.Hash{hashC})

				return "main", hashD.String()
			},
			expectedCommits: []string{"Feature commit 1", "Feature commit 2"},
		},
		{
			name: "Diverged branches - only new commits",
			setupRepo: func(t *testing.T, repo *gogit.Repository) (string, string) {
				t.Helper()
				// Common ancestor
				hashA := createTestCommit(t, repo, "Initial commit", nil)

				// Main branch continues
				hashB := createTestCommit(t, repo, "Main branch commit", []plumbing.Hash{hashA})
				hashC := createTestCommit(t, repo, "Main branch commit 2", []plumbing.Hash{hashB})

				// Create main branch ref
				mainRef := plumbing.NewHashReference("refs/heads/main", hashC)
				err := repo.Storer.SetReference(mainRef)
				require.NoError(t, err)

				// Feature branch from common ancestor
				hashD := createTestCommit(t, repo, "Feature branch commit", []plumbing.Hash{hashA})
				hashE := createTestCommit(t, repo, "Feature branch commit 2", []plumbing.Hash{hashD})

				return "main", hashE.String()
			},
			expectedCommits: []string{"Feature branch commit", "Feature branch commit 2"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "gommitlint-branch-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Initialize repository
			repo, err := gogit.PlainInit(tmpDir, false)
			require.NoError(t, err)

			// Setup test scenario
			baseRef, targetRef := testCase.setupRepo(t, repo)

			// Create repository adapter
			ctx := context.Background()
			adapter, err := git.NewRepository(tmpDir)
			require.NoError(t, err)

			commitRepo := adapter

			// Get commits in range
			commits, err := commitRepo.GetCommitRange(ctx, baseRef, targetRef)
			require.NoError(t, err)

			// Verify we got the expected commits
			subjects := make([]string, len(commits))
			for i, c := range commits {
				subjects[i] = c.Subject
			}

			require.Len(t, commits, len(testCase.expectedCommits), "Expected %d commits but got %d", len(testCase.expectedCommits), len(commits))

			for _, expectedSubject := range testCase.expectedCommits {
				require.Contains(t, subjects, expectedSubject, "Expected to find commit with subject '%s'", expectedSubject)
			}
		})
	}
}

// createTestCommit creates a commit with a specific message for testing.
func createTestCommit(t *testing.T, repo *gogit.Repository, message string, parents []plumbing.Hash) plumbing.Hash {
	t.Helper()

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	// Create unique file for each commit
	filename := filepath.Join(worktree.Filesystem.Root(), message+".txt")
	err = os.WriteFile(filename, []byte(message+"\n\nTest content"), 0600)
	require.NoError(t, err)

	_, err = worktree.Add(filepath.Base(filename))
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
