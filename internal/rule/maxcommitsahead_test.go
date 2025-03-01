// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"

	gitClient "github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/rule"
)

func TestValidateNumberOfCommits(t *testing.T) {
	tests := []struct {
		name          string
		setupRepo     func(t *testing.T, repo *git.Repository, w *git.Worktree)
		ref           string
		opts          []func(*rule.CommitsAheadConfig)
		expectedAhead int
		expectError   bool
		errorContains string
		currentBranch string
	}{
		{
			name: "single commit ahead of main",
			setupRepo: func(t *testing.T, repo *git.Repository, w *git.Worktree) {
				t.Helper()
				createCommit(t, repo, w, "feat: add new feature")
			},
			ref:           "main",
			expectedAhead: 1,
			expectError:   false,
			currentBranch: "feature",
		},
		{
			name: "two commits ahead exceeds limit",
			setupRepo: func(t *testing.T, repo *git.Repository, w *git.Worktree) {
				t.Helper()
				createCommit(t, repo, w, "feat: first feature")
				createCommit(t, repo, w, "feat: second feature")
			},
			ref:           "main",
			expectedAhead: 2,
			expectError:   true,
			errorContains: "HEAD is 2 commit(s) ahead of refs/heads/main (max: 1)",
			currentBranch: "feature",
		},
		{
			name: "ignored branch should pass",
			setupRepo: func(t *testing.T, repo *git.Repository, w *git.Worktree) {
				t.Helper()
				createCommit(t, repo, w, "feat: first feature")
				createCommit(t, repo, w, "feat: second feature")
			},
			ref:           "main",
			opts:          []func(*rule.CommitsAheadConfig){rule.WithIgnoreBranches("feature")},
			expectedAhead: 0,
			expectError:   false,
			currentBranch: "feature",
		},
		{
			name: "enforce on specific branch",
			setupRepo: func(t *testing.T, repo *git.Repository, w *git.Worktree) {
				t.Helper()
				createCommit(t, repo, w, "feat: first feature")
				createCommit(t, repo, w, "feat: second feature")
			},
			ref:           "main",
			opts:          []func(*rule.CommitsAheadConfig){rule.WithEnforceOnBranches("other")},
			expectedAhead: 0,
			expectError:   false,
			currentBranch: "feature",
		},
		{
			name: "custom max commits ahead",
			setupRepo: func(t *testing.T, repo *git.Repository, w *git.Worktree) {
				t.Helper()
				createCommit(t, repo, w, "feat: first feature")
				createCommit(t, repo, w, "feat: second feature")
			},
			ref:           "main",
			opts:          []func(*rule.CommitsAheadConfig){rule.WithMaxCommitsAhead(2)},
			expectedAhead: 2,
			expectError:   false,
			currentBranch: "feature",
		},
		{
			name: "non-existent reference",
			setupRepo: func(t *testing.T, repo *git.Repository, w *git.Worktree) {
				t.Helper()
				createCommit(t, repo, w, "feat: add feature")
			},
			ref:           "non-existent",
			expectedAhead: 0,
			expectError:   false,
			currentBranch: "feature",
		},
	}
	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Create temporary directory for the test repository
			tmpDir := t.TempDir()

			// Initialize repository
			repo, err := git.PlainInit(tmpDir, false)
			require.NoError(t, err)

			// Create worktree
			wtree, err := repo.Worktree()
			require.NoError(t, err)

			// Create initial commit
			createInitialCommit(t, repo, wtree)

			// Create and checkout main branch
			err = wtree.Checkout(&git.CheckoutOptions{
				Create: true,
				Branch: plumbing.NewBranchReferenceName("main"),
			})
			require.NoError(t, err)

			// If the test case is for a feature branch, create and checkout it
			if tabletest.currentBranch != "main" {
				err = wtree.Checkout(&git.CheckoutOptions{
					Create: true,
					Branch: plumbing.NewBranchReferenceName(tabletest.currentBranch),
				})
				require.NoError(t, err)
			}

			// Setup the test case (create additional commits if needed)
			tabletest.setupRepo(t, repo, wtree)

			// Create git client
			client := &gitClient.Repository{
				Repo: repo,
			}

			ruleInfo := rule.ValidateNumberOfCommits(client, tabletest.ref, tabletest.opts...)
			nr, _ := ruleInfo.(*rule.MaxCommitsAhead)
			require.Equal(t, tabletest.expectedAhead, nr.Ahead)

			if tabletest.expectError {
				require.NotEmpty(t, ruleInfo.Errors())

				if tabletest.errorContains != "" {
					require.Contains(t, ruleInfo.Errors()[0].Error(), tabletest.errorContains)
				}
			} else {
				require.Empty(t, ruleInfo.Errors())
			}
		})
	}
}

func createInitialCommit(t *testing.T, _ *git.Repository, wtree *git.Worktree) plumbing.Hash {
	t.Helper()
	// Create a dummy file
	filename := filepath.Join(wtree.Filesystem.Root(), "initial.txt")
	err := os.WriteFile(filename, []byte("initial content"), 0600)
	require.NoError(t, err)

	// Stage the file
	_, err = wtree.Add("initial.txt")
	require.NoError(t, err)

	// Commit the file
	hash, err := wtree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)
	fmt.Printf("DEBUG: Created initial commit with hash %s\n", hash.String())

	return hash
}

func createCommit(t *testing.T, _ *git.Repository, wtree *git.Worktree, message string) plumbing.Hash {
	t.Helper()
	// Create a new file for this commit
	filename := filepath.Join(wtree.Filesystem.Root(), message+".txt")
	err := os.WriteFile(filename, []byte("content for "+message), 0600)
	require.NoError(t, err)

	// Stage the file
	_, err = wtree.Add(filepath.Base(filename))
	require.NoError(t, err)

	// Commit the file
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
