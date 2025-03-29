// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/rule"
)

func TestValidateNumberOfCommits(t *testing.T) {
	tests := []struct {
		name          string
		setupRepo     func(t *testing.T, w *git.Worktree) int // returns # of commits created
		ref           string
		maxCommits    int
		expectError   bool
		errorContains string
	}{
		{
			name: "single commit ahead of main - within limit",
			setupRepo: func(t *testing.T, w *git.Worktree) int {
				t.Helper()
				createCommit(t, w, "feat: add new feature")

				return 1
			},
			ref:         "main",
			maxCommits:  20, // default
			expectError: false,
		},
		{
			name: "commits ahead exceed custom limit",
			setupRepo: func(t *testing.T, w *git.Worktree) int {
				t.Helper()
				createCommit(t, w, "feat: first feature")
				createCommit(t, w, "feat: second feature")

				return 2
			},
			ref:           "main",
			maxCommits:    1,
			expectError:   true,
			errorContains: "HEAD is 2 commit(s) ahead of main (max: 1)",
		},
		{
			name: "commits within custom limit",
			setupRepo: func(t *testing.T, w *git.Worktree) int {
				t.Helper()
				createCommit(t, w, "feat: first feature")
				createCommit(t, w, "feat: second feature")

				return 2
			},
			ref:         "main",
			maxCommits:  3,
			expectError: false,
		},
		{
			name: "non-existent reference",
			setupRepo: func(t *testing.T, w *git.Worktree) int {
				t.Helper()
				createCommit(t, w, "feat: add feature")

				return 0
			},
			ref:         "non-existent",
			maxCommits:  20,
			expectError: false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Setup test repository
			tmpDir := t.TempDir()
			repo, err := git.PlainInit(tmpDir, false)
			require.NoError(t, err)

			wtree, err := repo.Worktree()
			require.NoError(t, err)

			// Create initial commit on master
			createInitialCommit(t, wtree)

			// Create and checkout main branch (as reference)
			err = wtree.Checkout(&git.CheckoutOptions{
				Create: true,
				Branch: plumbing.NewBranchReferenceName("main"),
			})
			require.NoError(t, err)

			// Create and checkout feature branch for tests
			err = wtree.Checkout(&git.CheckoutOptions{
				Create: true,
				Branch: plumbing.NewBranchReferenceName("feature"),
			})
			require.NoError(t, err)

			// Setup the test case (create additional commits)
			commitsCreated := tabletest.setupRepo(t, wtree)

			// Create git client
			client := &model.Repository{Repo: repo}

			// Run the validation with options if provided
			var opts []rule.Option
			if tabletest.maxCommits != 20 { // Only add option if different from default
				opts = append(opts, rule.WithMaxCommitsAhead(tabletest.maxCommits))
			}

			// Validate commits ahead
			result := rule.ValidateNumberOfCommits(client, tabletest.ref, opts...)

			// Verify the result
			require.Equal(t, commitsCreated, result.Ahead, "Number of commits ahead should match")

			if tabletest.expectError {
				require.NotEmpty(t, result.Errors(), "Expected error but got none")

				if tabletest.errorContains != "" {
					require.Contains(t, result.Errors()[0].Error(), tabletest.errorContains,
						"Error message doesn't contain expected text")
				}
			} else {
				require.Empty(t, result.Errors(), "Expected no errors but got: %v", result.Errors())
			}
		})
	}
}

func createInitialCommit(t *testing.T, wtree *git.Worktree) plumbing.Hash {
	t.Helper()

	// Create a dummy file
	filename := filepath.Join(wtree.Filesystem.Root(), "initial.txt")
	err := os.WriteFile(filename, []byte("initial content"), 0600)
	require.NoError(t, err)

	// Stage and commit the file
	_, err = wtree.Add("initial.txt")
	require.NoError(t, err)

	hash, err := wtree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	return hash
}

func createCommit(t *testing.T, wtree *git.Worktree, message string) plumbing.Hash {
	t.Helper()

	// Create a new file for this commit
	filename := filepath.Join(wtree.Filesystem.Root(), message+".txt")
	err := os.WriteFile(filename, []byte("content for "+message), 0600)
	require.NoError(t, err)

	// Stage and commit the file
	_, err = wtree.Add(filepath.Base(filename))
	require.NoError(t, err)

	hash, err := wtree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	return hash
}
