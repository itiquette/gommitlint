// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/adapters/git"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
)

// ValidationResult represents the result of a validation operation.
type ValidationResult struct {
	Valid  bool
	Errors []domain.ValidationError
}

// TestValidation provides a simple interface for testing validation.
func TestValidation(t *testing.T, repoPath string, config config.Config) ValidationResult {
	t.Helper()

	// Create context
	ctx := context.Background()

	// Create git repository adapter directly
	gitRepo, err := git.NewRepository(repoPath)
	require.NoError(t, err, "Failed to create git repository")

	// Create validation rules
	commitRules := rules.CreateCommitRules(config)
	repoRules := rules.CreateRepositoryRules(config)

	// Get the latest commit (HEAD)
	commits, err := gitRepo.GetHeadCommits(ctx, 1)
	require.NoError(t, err, "Failed to get latest commit")
	require.Len(t, commits, 1, "Expected exactly one commit")

	commit := commits[0]

	// Validate commit
	result := domain.ValidateCommit(commit, commitRules, repoRules, gitRepo, config)

	// Use errors directly from ValidationResult
	allErrors := result.Errors

	// Convert to simple result
	return ValidationResult{
		Valid:  len(allErrors) == 0,
		Errors: allErrors,
	}
}

// GitRepo creates a test repository with specified commit messages.
// Returns the repository path and a cleanup function.
func GitRepo(t *testing.T, commits ...string) (string, func()) {
	t.Helper()

	// Create temp directory
	tempDir := t.TempDir()

	// Initialize repository
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err, "Failed to init repo")

	// Create initial commit if no commits provided
	if len(commits) == 0 {
		commits = []string{"Initial commit"}
	}

	// Add commits
	worktree, err := repo.Worktree()
	require.NoError(t, err, "Failed to get worktree")

	for i, message := range commits {
		// Create a file for each commit
		filename := fmt.Sprintf("file%d.txt", i)
		filePath := filepath.Join(tempDir, filename)
		content := fmt.Sprintf("Content for commit %d", i)

		err := os.WriteFile(filePath, []byte(content), 0600)
		require.NoError(t, err, "Failed to write file")

		// Stage file
		_, err = worktree.Add(filename)
		require.NoError(t, err, "Failed to stage file")

		// Create commit
		_, err = worktree.Commit(message, &gogit.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err, "Failed to create commit")
	}

	cleanup := func() {
		// Cleanup is handled by t.TempDir()
	}

	return tempDir, cleanup
}

// TestValidateMessage validates a commit message directly without requiring a git repository.
func TestValidateMessage(t *testing.T, message string, config config.Config) ValidationResult {
	t.Helper()

	// Create a temporary git repo with the message
	repoPath, cleanup := GitRepo(t, message)
	defer cleanup()

	return TestValidation(t, repoPath, config)
}

// DefaultConfig returns a sensible default configuration for testing.
func DefaultConfig() config.Config {
	return config.Config{
		Message: config.MessageConfig{
			Subject: config.SubjectConfig{
				Case:              "ignore",
				MaxLength:         72,
				RequireImperative: false,
				ForbidEndings:     []string{"."},
			},
			Body: config.BodyConfig{
				Required:         true,
				MinLength:        10,
				AllowSignoffOnly: false,
				MinSignoffCount:  0,
			},
		},
		Conventional: config.ConventionalConfig{
			RequireScope:         false,
			Types:                []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"},
			AllowBreaking:        true,
			MaxDescriptionLength: 72,
		},
		Rules: config.RulesConfig{
			Enabled: []string{
				"Subject",
				"ConventionalCommit",
			},
			Disabled: []string{
				"CommitBody",
				"Signature",
				"SignOff",
				"Identity",
				"JiraReference",
				"Spell",
			},
		},
		Signature: config.SignatureConfig{
			Required:       false,
			VerifyFormat:   false,
			KeyDirectory:   "",
			AllowedSigners: []string{},
		},
		Identity: config.IdentityConfig{
			AllowedAuthors: []string{},
		},
		Repo: config.RepoConfig{
			MaxCommitsAhead:   10,
			ReferenceBranch:   "main",
			AllowMergeCommits: true,
		},
		Output: "text",
		Spell: config.SpellConfig{
			Locale:      "en_US",
			IgnoreWords: []string{},
		},
		Jira: config.JiraConfig{
			ProjectPrefixes:      []string{},
			RequireInBody:        false,
			RequireInSubject:     false,
			IgnoreTicketPatterns: []string{},
		},
	}
}

// WithRules returns a config with only the specified rules enabled.
func WithRules(rules ...string) config.Config {
	config := DefaultConfig()
	config.Rules.Enabled = rules

	// Disable all other commonly enabled rules
	allRules := []string{
		"Subject", "ConventionalCommit", "CommitBody", "Signature",
		"SignOff", "Identity", "JiraReference", "Spell",
		"BranchAhead",
	}

	// Build disabled list from all rules except the ones we want enabled
	var disabled []string

	for _, rule := range allRules {
		found := false

		for _, enabledRule := range rules {
			if rule == enabledRule {
				found = true

				break
			}
		}

		if !found {
			disabled = append(disabled, rule)
		}
	}

	config.Rules.Disabled = disabled

	return config
}

// WithSubjectMaxLength returns a config with custom subject max length.
func WithSubjectMaxLength(length int) config.Config {
	config := DefaultConfig()
	config.Message.Subject.MaxLength = length

	return config
}
