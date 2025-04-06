// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

func TestGitService(t *testing.T) {
	// Create a temporary directory for our test repos
	tmpDir, err := os.MkdirTemp("", "gommitlint-git-service-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("DetectMainBranch", func(t *testing.T) {
		tests := []struct {
			name           string
			setup          func(t *testing.T) string
			expectedBranch string
			expectError    bool
		}{
			{
				name: "detect_main_branch",
				setup: func(t *testing.T) string {
					t.Helper()
					repoPath := filepath.Join(tmpDir, "main-branch")
					setupRepo(t, repoPath, "main")

					return repoPath
				},
				expectedBranch: "main",
				expectError:    false,
			},
			{
				name: "detect_master_branch",
				setup: func(t *testing.T) string {
					t.Helper()
					repoPath := filepath.Join(tmpDir, "master-branch")
					setupRepo(t, repoPath, "master")

					return repoPath
				},
				expectedBranch: "master",
				expectError:    false,
			},
			{
				name: "fallback_with_warning",
				setup: func(t *testing.T) string {
					t.Helper()
					repoPath := filepath.Join(tmpDir, "custom-branch")
					// Only create a development branch, no main or master
					setupRepoWithCustomBranch(t, repoPath, "development")

					return repoPath
				},
				expectedBranch: "main", // Fall back to main with warning
				expectError:    true,   // Now we expect an error message
			},
			{
				name: "not_a_git_repo",
				setup: func(t *testing.T) string {
					t.Helper()
					nonRepoPath := filepath.Join(tmpDir, "not-a-repo")
					err := os.MkdirAll(nonRepoPath, 0755)
					require.NoError(t, err)

					return nonRepoPath
				},
				expectedBranch: "",    // Empty string for non-git repo
				expectError:    false, // Not an error, just returns empty string
			},
		}

		for _, tabletest := range tests {
			t.Run(tabletest.name, func(t *testing.T) {
				repoPath := tabletest.setup(t)

				// Create service for the specific repo path
				service, err := NewServiceForPath(repoPath)
				require.NoError(t, err)

				// Test the detection
				branch, err := service.DetectMainBranch()

				if tabletest.expectError {
					require.Error(t, err)
					require.Contains(t, err.Error(), "neither 'main' nor 'master' branch found")
				} else {
					require.NoError(t, err)
				}

				require.Equal(t, tabletest.expectedBranch, branch)
			})
		}
	})

	t.Run("RefExists", func(t *testing.T) {
		// Setup a repo with main branch and a commit
		repoPath := filepath.Join(tmpDir, "ref-exists-repo")
		repo := setupRepo(t, repoPath, "main")

		// Add another branch
		headRef, err := repo.Head()
		require.NoError(t, err)

		featureBranchRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName("feature"), headRef.Hash())
		err = repo.Storer.SetReference(featureBranchRef)
		require.NoError(t, err)

		// Create a tag reference
		tagRef := plumbing.NewHashReference(plumbing.NewTagReferenceName("v1.0.0"), headRef.Hash())
		err = repo.Storer.SetReference(tagRef)
		require.NoError(t, err)

		// Create service for testing
		service, err := NewServiceForPath(repoPath)
		require.NoError(t, err)

		tests := []struct {
			name      string
			reference string
			expected  bool
		}{
			{
				name:      "main_branch_exists",
				reference: "refs/heads/main",
				expected:  true,
			},
			{
				name:      "feature_branch_exists",
				reference: "refs/heads/feature",
				expected:  true,
			},
			{
				name:      "tag_exists",
				reference: "refs/tags/v1.0.0",
				expected:  true,
			},
			{
				name:      "non_existent_branch",
				reference: "refs/heads/nonexistent",
				expected:  false,
			},
			{
				name:      "non_existent_tag",
				reference: "refs/tags/v2.0.0",
				expected:  false,
			},
			{
				name:      "commit_hash_exists",
				reference: headRef.Hash().String(),
				expected:  true,
			},
			{
				name:      "invalid_commit_hash",
				reference: "1234567890abcdef",
				expected:  false,
			},
		}

		for _, tabletest := range tests {
			t.Run(tabletest.name, func(t *testing.T) {
				exists := service.RefExists(tabletest.reference)
				require.Equal(t, tabletest.expected, exists)
			})
		}
	})

	t.Run("IsMergeCommit", func(t *testing.T) {
		// Create a repo with a merge commit
		repoPath := filepath.Join(tmpDir, "merge-commit-repo")
		repo := setupRepo(t, repoPath, "main")

		// Get the initial commit hash (non-merge commit)
		headRef, err := repo.Head()
		require.NoError(t, err)

		nonMergeHash := headRef.Hash()

		// Create a feature branch
		featureBranchRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName("feature"), nonMergeHash)
		err = repo.Storer.SetReference(featureBranchRef)
		require.NoError(t, err)

		// Create a develop branch
		developBranchRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName("develop"), nonMergeHash)
		err = repo.Storer.SetReference(developBranchRef)
		require.NoError(t, err)

		// Add a commit to develop branch
		worktree, err := repo.Worktree()
		require.NoError(t, err)

		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName("develop"),
		})
		require.NoError(t, err)

		developFile := filepath.Join(repoPath, "develop.txt")
		err = os.WriteFile(developFile, []byte("Develop branch"), 0600)
		require.NoError(t, err)

		_, err = worktree.Add("develop.txt")
		require.NoError(t, err)

		developHash, err := worktree.Commit("Develop commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test",
				Email: "test@example.com",
			},
		})
		require.NoError(t, err)

		// Create a merge commit with multiple parents
		// We have to construct it manually as go-git doesn't support merge operations directly

		// First, checkout the feature branch
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName("feature"),
		})
		require.NoError(t, err)

		// Add a change
		featureFile := filepath.Join(repoPath, "feature.txt")
		err = os.WriteFile(featureFile, []byte("Feature branch"), 0600)
		require.NoError(t, err)

		_, err = worktree.Add("feature.txt")
		require.NoError(t, err)

		// Make a commit on feature
		featureHash, err := worktree.Commit("Feature commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test",
				Email: "test@example.com",
			},
		})
		require.NoError(t, err)

		// Return to main branch
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName("main"),
		})
		require.NoError(t, err)

		// Create a merge commit by simulating a merge (go-git doesn't have direct merge support)
		// We do this by creating a commit with multiple parents

		// Copy both files to the main branch to simulate merge result
		err = os.WriteFile(filepath.Join(repoPath, "develop.txt"), []byte("Develop branch"), 0600)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(repoPath, "feature.txt"), []byte("Feature branch"), 0600)
		require.NoError(t, err)

		_, err = worktree.Add("develop.txt")
		require.NoError(t, err)
		_, err = worktree.Add("feature.txt")
		require.NoError(t, err)

		// Create merge commit with multiple parents (simulating the merge)
		mergeHash, err := worktree.Commit("Merge branches", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test",
				Email: "test@example.com",
			},
			Parents: []plumbing.Hash{nonMergeHash, featureHash, developHash},
		})
		require.NoError(t, err)

		// Test each commit
		isMerge, err := IsMergeCommit(repo, mergeHash)
		require.NoError(t, err)
		require.True(t, isMerge, "Merge commit should be identified as such")

		isMerge, err = IsMergeCommit(repo, nonMergeHash)
		require.NoError(t, err)
		require.False(t, isMerge, "Regular commit should not be identified as merge commit")
	})
}

// Helper function to set up a git repository with an initial commit and a specified branch.
func setupRepo(t *testing.T, path string, branchName string) *git.Repository {
	t.Helper()

	err := os.MkdirAll(path, 0755)
	require.NoError(t, err)

	// Initialize a new repo
	repo, err := git.PlainInit(path, false)
	require.NoError(t, err)

	// Create an initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	// Add a file
	filePath := filepath.Join(path, "file.txt")
	err = os.WriteFile(filePath, []byte("Initial content"), 0600)
	require.NoError(t, err)

	_, err = worktree.Add("file.txt")
	require.NoError(t, err)

	// Commit
	hash, err := worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Create the branch
	branchRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branchName), hash)
	err = repo.Storer.SetReference(branchRef)
	require.NoError(t, err)

	// Make HEAD point to this branch
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName(branchName))
	err = repo.Storer.SetReference(headRef)
	require.NoError(t, err)

	return repo
}

// New helper function to create a repo with only a custom branch (no main or master).
func setupRepoWithCustomBranch(t *testing.T, path string, branchName string) *git.Repository {
	t.Helper()

	err := os.MkdirAll(path, 0755)
	require.NoError(t, err)

	// Initialize a new repo
	repo, err := git.PlainInit(path, false)
	require.NoError(t, err)

	// Create an initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	// Add a file
	filePath := filepath.Join(path, "file.txt")
	err = os.WriteFile(filePath, []byte("Initial content"), 0600)
	require.NoError(t, err)

	_, err = worktree.Add("file.txt")
	require.NoError(t, err)

	// Commit
	hash, err := worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Create only the requested branch and remove defaults
	customBranchRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branchName), hash)
	err = repo.Storer.SetReference(customBranchRef)
	require.NoError(t, err)

	// Make HEAD point to this branch
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName(branchName))
	err = repo.Storer.SetReference(headRef)
	require.NoError(t, err)

	// The issue is that git.PlainInit might create a "master" or "main" branch by default
	// Let's check and remove any default branches that might have been created

	// Remove "main" if it exists (and is not our custom branch)
	if branchName != "main" {
		mainRef := plumbing.NewBranchReferenceName("main")
		_, err = repo.Reference(mainRef, true)

		if err == nil {
			// If no error, the reference exists, so remove it
			err = repo.Storer.RemoveReference(mainRef)
			require.NoError(t, err)
		}
	}

	// Remove "master" if it exists (and is not our custom branch)
	if branchName != "master" {
		masterRef := plumbing.NewBranchReferenceName("master")
		_, err = repo.Reference(masterRef, true)

		if err == nil {
			// If no error, the reference exists, so remove it
			err = repo.Storer.RemoveReference(masterRef)
			require.NoError(t, err)
		}
	}

	// Verify only our custom branch and HEAD exist
	refs, err := repo.References()
	require.NoError(t, err)

	count := 0
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		t.Logf("Reference found: %s", ref.Name())

		count++

		return nil
	})
	require.NoError(t, err)

	// Should only have our custom branch and HEAD = 2 references
	require.Equal(t, 2, count, "Should only have custom branch and HEAD references")

	return repo
}
