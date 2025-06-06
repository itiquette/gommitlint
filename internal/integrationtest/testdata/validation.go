// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package testdata provides test fixtures and helpers for integration tests.
package testdata

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/adapters/git"
	gitTestdata "github.com/itiquette/gommitlint/internal/adapters/git/testdata"
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
// It creates all necessary components internally without exposing complexity.
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

// TestValidateMessage validates a commit message directly without requiring a git repository.
func TestValidateMessage(t *testing.T, message string, config config.Config) ValidationResult {
	t.Helper()

	// Create a temporary git repo with the message
	repoPath, cleanup := gitTestdata.GitRepo(t, message)
	defer cleanup()

	return TestValidation(t, repoPath, config)
}

// DefaultConfig returns a sensible default configuration for testing.
func DefaultConfig() config.Config {
	return config.Config{
		Message: config.MessageConfig{
			Subject: config.SubjectConfig{
				Case:              "ignore", // Disable case checking by default for simplicity
				MaxLength:         72,
				RequireImperative: false,
				ForbidEndings:     []string{"."},
			},
			Body: config.BodyConfig{
				MinLength:        10,
				MinLines:         3,
				AllowSignoffOnly: false,
				RequireSignoff:   false,
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
				"SignedIdentity",
				"JiraReference",
				"Spell",
			},
		},
		Signing: config.SigningConfig{
			RequireSignature:    false,
			RequireVerification: false,
			RequireMultiSignoff: false,
			KeyDirectory:        "",
			AllowedSigners:      []string{},
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
		"SignOff", "SignedIdentity", "JiraReference", "Spell",
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

// WithConventionalTypes returns a config with custom conventional commit config.
func WithConventionalTypes(types ...string) config.Config {
	config := DefaultConfig()
	config.Conventional.Types = types

	return config
}