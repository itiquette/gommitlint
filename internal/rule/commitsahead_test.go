// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/rule"
)

func TestValidateNumberOfCommits(t *testing.T) {
	tests := []struct {
		name                    string
		setupRepo               func(t *testing.T, w *git.Worktree) int // returns # of commits created
		ref                     string
		maxCommits              int
		expectError             bool
		errorCode               string
		errorContains           string
		expectedResult          string
		expectedVerboseContains []string // New field for testing verbose output
	}{
		{
			name: "single commit ahead of main - within limit",
			setupRepo: func(t *testing.T, w *git.Worktree) int {
				t.Helper()
				createCommit(t, w, "feat: add new feature")

				return 1
			},
			ref:                     "main",
			maxCommits:              20, // default
			expectError:             false,
			expectedResult:          "HEAD is 1 commit(s) ahead of main",
			expectedVerboseContains: []string{"HEAD is 1 commit ahead of main", "within limit of 20"},
		},
		{
			name: "commits ahead exceed custom limit",
			setupRepo: func(t *testing.T, w *git.Worktree) int {
				t.Helper()
				createCommit(t, w, "feat: first feature")
				createCommit(t, w, "feat: second feature")

				return 2
			},
			ref:                     "main",
			maxCommits:              1,
			expectError:             true,
			errorCode:               "too_many_commits",
			errorContains:           "HEAD is 2 commit(s) ahead of main (max: 1)",
			expectedResult:          "Too many commits ahead of main",
			expectedVerboseContains: []string{"HEAD is 2 commit(s) ahead of main", "maximum allowed: 1", "Consider merging or rebasing"},
		},
		{
			name: "commits within custom limit",
			setupRepo: func(t *testing.T, w *git.Worktree) int {
				t.Helper()
				createCommit(t, w, "feat: first feature")
				createCommit(t, w, "feat: second feature")

				return 2
			},
			ref:                     "main",
			maxCommits:              3,
			expectError:             false,
			expectedResult:          "HEAD is 2 commit(s) ahead of main",
			expectedVerboseContains: []string{"HEAD is 2 commits ahead of main", "within limit of 3"},
		},
		{
			name: "no commits ahead",
			setupRepo: func(t *testing.T, _ *git.Worktree) int {
				t.Helper()
				// Don't create any commits
				return 0
			},
			ref:                     "main",
			maxCommits:              20,
			expectError:             false,
			expectedResult:          "HEAD is 0 commit(s) ahead of main",
			expectedVerboseContains: []string{"HEAD is up-to-date with main", "0 commits ahead"},
		},
		{
			name: "empty reference",
			setupRepo: func(t *testing.T, _ *git.Worktree) int {
				t.Helper()

				return 0
			},
			ref:                     "",
			maxCommits:              20,
			expectError:             true,
			errorCode:               "empty_ref",
			errorContains:           "reference cannot be empty",
			expectedResult:          "Too many commits ahead of ", // Empty ref
			expectedVerboseContains: []string{"Reference branch name is empty"},
		},
		{
			name: "many commits exceed limit",
			setupRepo: func(t *testing.T, w *git.Worktree) int {
				t.Helper()
				// Create 30 commits
				for i := 1; i <= 30; i++ {
					createCommit(t, w, fmt.Sprintf("feat: feature %d", i))
				}

				return 30
			},
			ref:                     "main",
			maxCommits:              20,
			expectError:             true,
			errorCode:               "too_many_commits",
			errorContains:           "HEAD is 30 commit(s) ahead of main (max: 20)",
			expectedResult:          "Too many commits ahead of main",
			expectedVerboseContains: []string{"HEAD is 30 commit(s) ahead of main", "maximum allowed: 20"},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Setup test repository
			repo, wtree := setupTestRepo(t)

			// Create initial commit on master
			createInitialCommit(t, wtree)

			// Create and checkout main branch (as reference)
			err := wtree.Checkout(&git.CheckoutOptions{
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
			assert.Equal(t, commitsCreated, result.Ahead, "Number of commits ahead should match")

			// Check rule name
			assert.Equal(t, "CommitsAhead", result.Name(), "Rule name should be correct")

			// Check Result() output
			if tabletest.expectedResult != "" {
				assert.Contains(t, result.Result(), tabletest.expectedResult, "Result() should match expected value")
			}

			// Check VerboseResult() output
			verboseResult := result.VerboseResult()
			assert.NotEmpty(t, verboseResult, "VerboseResult should not be empty")

			// The verbose output should be different from the regular result
			if !tabletest.expectError || commitsCreated > 0 { // Skip this check for errors with 0 commits
				assert.NotEqual(t, result.Result(), verboseResult,
					"VerboseResult should differ from Result to provide more details")
			}

			// Verify expected phrases in verbose output
			for _, phrase := range tabletest.expectedVerboseContains {
				assert.Contains(t, verboseResult, phrase,
					"VerboseResult should contain expected phrase: %s", phrase)
			}

			// Check errors
			if tabletest.expectError {
				assert.NotEmpty(t, result.Errors(), "Expected error but got none")

				valErr := result.Errors()[0]
				assert.Equal(t, "CommitsAhead", valErr.Rule, "Rule name should be set in ValidationError")

				if tabletest.errorCode != "" {
					assert.Equal(t, tabletest.errorCode, valErr.Code, "Error code should match expected")
				}

				if tabletest.errorContains != "" {
					assert.Contains(t, valErr.Message, tabletest.errorContains,
						"Error message doesn't contain expected text")
				}

				// Verify help method returns non-empty string
				assert.NotEmpty(t, result.Help(), "Help should provide guidance")

				// Verify context for too_many_commits error
				if tabletest.errorCode == "too_many_commits" {
					assert.Contains(t, valErr.Context, "actual")
					assert.Contains(t, valErr.Context, "maximum")
					assert.Contains(t, valErr.Context, "reference")
				}
			} else {
				assert.Empty(t, result.Errors(), "Expected no errors but got: %v", result.Errors())
			}
		})
	}
}

func TestValidateNumberOfCommitsWithNilRepo(t *testing.T) {
	// Test with nil repository
	result := rule.ValidateNumberOfCommits(nil, "main")

	assert.NotNil(t, result, "Result should not be nil even with nil repo")
	assert.NotEmpty(t, result.Errors(), "Should have errors with nil repo")

	valErr := result.Errors()[0]
	assert.Equal(t, "CommitsAhead", valErr.Rule, "Rule name should be set")
	assert.Equal(t, "nil_repo", valErr.Code, "Error code should indicate nil repository")
	assert.Equal(t, "repository cannot be nil", valErr.Message, "Error message should indicate nil repository")

	// Also test verbose result for nil repo
	verboseResult := result.VerboseResult()
	assert.Contains(t, verboseResult, "Repository object is nil",
		"VerboseResult should explain the nil repository issue")
}

func TestCommitsAheadHelpMethod(t *testing.T) {
	// Add a mock error using reflection or create a rule with a known error
	repo, wtree := setupTestRepo(t)
	createInitialCommit(t, wtree)

	err := wtree.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: plumbing.NewBranchReferenceName("main"),
	})
	require.NoError(t, err)

	err = wtree.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: plumbing.NewBranchReferenceName("feature"),
	})
	require.NoError(t, err)

	// Create 25 commits
	for i := 1; i <= 25; i++ {
		createCommit(t, wtree, fmt.Sprintf("feat: feature %d", i))
	}

	client := &model.Repository{Repo: repo}
	commitsAhead := rule.ValidateNumberOfCommits(client, "main")

	// Verify help content
	helpText := commitsAhead.Help()
	assert.Contains(t, helpText, "too many commits ahead", "Help should explain the issue")
	assert.Contains(t, helpText, "Merge or rebase", "Help should suggest merging or rebasing")
	assert.Contains(t, helpText, "splitting your changes", "Help should suggest splitting changes")

	// Verify error code
	assert.Equal(t, "too_many_commits", commitsAhead.Errors()[0].Code, "Error code should be 'too_many_commits'")

	// Test and verify verbose output
	verboseOutput := commitsAhead.VerboseResult()
	assert.Contains(t, verboseOutput, "HEAD is 25 commit(s) ahead of main")
	assert.Contains(t, verboseOutput, "maximum allowed: 20")
	assert.Contains(t, verboseOutput, "Consider merging or rebasing")

	// Test different error codes' help messages
	t.Run("help for nil repo", func(t *testing.T) {
		result := rule.ValidateNumberOfCommits(nil, "main")
		helpText := result.Help()
		assert.Contains(t, helpText, "valid git repository", "Help should mention repository requirement")
	})

	t.Run("help for empty reference", func(t *testing.T) {
		result := rule.ValidateNumberOfCommits(client, "")
		helpText := result.Help()
		assert.Contains(t, helpText, "valid reference branch", "Help should mention reference requirement")
	})
}

// setupTestRepo creates a new Git repository for testing.
func setupTestRepo(t *testing.T) (*git.Repository, *git.Worktree) {
	t.Helper()

	tmpDir := t.TempDir()
	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	wtree, err := repo.Worktree()
	require.NoError(t, err)

	return repo, wtree
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
	filename := filepath.Join(wtree.Filesystem.Root(), fmt.Sprintf("%s-%d.txt", message, time.Now().UnixNano()))
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
