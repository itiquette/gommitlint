// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

//nolint:testpackage
package rules_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/itiquette/gommitlint/internal"
	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/itiquette/gommitlint/internal/model"
)

// Common test structures and helpers.
type testDesc struct {
	name         string
	createCommit func(*git.Repository) error
	expectValid  bool
}

func runTestGroup(t *testing.T, tests []testDesc) {
	t.Helper()

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			dir := t.TempDir()

			err := os.Chdir(dir)
			if err != nil {
				t.Error(err)
			}

			repo, err := initRepo(dir)
			if err != nil {
				t.Error(err)
			}

			err = tabletest.createCommit(repo)
			if err != nil {
				t.Error(err)
			}

			report, err := runCompliance()
			if err != nil {
				t.Error(err)
			}

			if tabletest.expectValid != report.Valid() {
				t.Errorf("%s: Expected validity %v, got %v", tabletest.name, tabletest.expectValid, report.Valid())
			}
		})
	}
}

// Base test groups.
func TestTypeValidation(t *testing.T) {
	tests := []testDesc{
		{
			name:         "Valid Feat Type",
			createCommit: createCommitWithMsg("feat: add new feature"),
			expectValid:  true,
		},
		{
			name:         "Valid Fix Type",
			createCommit: createCommitWithMsg("fix: resolve bug"),
			expectValid:  true,
		},
		{
			name:         "Invalid Single Char Type",
			createCommit: createCommitWithMsg("f: too short"),
			expectValid:  false,
		},
		{
			name:         "Invalid Type With Symbols",
			createCommit: createCommitWithMsg("feat$: invalid symbol"),
			expectValid:  false,
		},
		{
			name:         "Invalid Type With Numbers",
			createCommit: createCommitWithMsg("feat1: no numbers allowed"),
			expectValid:  false,
		},
		{
			name:         "Invalid Mixed Case Type",
			createCommit: createCommitWithMsg("Feat: should be lowercase"),
			expectValid:  false,
		},
	}

	runTestGroup(t, tests)
}

func TestScopeValidation(t *testing.T) {
	tests := []testDesc{
		{
			name:         "Valid Single Scope",
			createCommit: createValidScopedCommit,
			expectValid:  true,
		},
		{
			name:         "Valid Multiple Scopes",
			createCommit: createCommitWithMsg("feat(scope1,scope2): multiple scopes"),
			expectValid:  true,
		},
		{
			name:         "Empty Scope",
			createCommit: createCommitWithMsg("feat(): empty scope"),
			expectValid:  false,
		},
		{
			name:         "Invalid Scope Characters",
			createCommit: createCommitWithMsg("feat(scope$): invalid character"),
			expectValid:  false,
		},
		{
			name:         "Scope With Spaces",
			createCommit: createCommitWithMsg("feat(scope 1): space not allowed"),
			expectValid:  false,
		},
		{
			name:         "Valid Scope With Underscore",
			createCommit: createValidCommitRegex,
			expectValid:  true,
		},
		{
			name:         "Invalid Scope Format",
			createCommit: createInvalidCommitRegex,
			expectValid:  false,
		},
	}

	runTestGroup(t, tests)
}

func TestBreakingChangeValidation(t *testing.T) {
	tests := []testDesc{
		{
			name:         "Valid Breaking Change Symbol",
			createCommit: createValidBreakingCommit,
			expectValid:  true,
		},
		{
			name:         "Valid Breaking Change With Scope",
			createCommit: createValidScopedBreakingCommit,
			expectValid:  true,
		},
		{
			name:         "Invalid Breaking Change Position",
			createCommit: createInvalidScopedBreakingCommit,
			expectValid:  false,
		},
		{
			name:         "Invalid Breaking Symbol",
			createCommit: createInvalidBreakingSymbolCommit,
			expectValid:  false,
		},
		{
			name:         "Multiple Breaking Symbols",
			createCommit: createCommitWithMsg("feat!!: multiple not allowed"),
			expectValid:  false,
		},
		{
			name:         "Breaking Change In Wrong Position",
			createCommit: createCommitWithMsg("feat:! wrong position"),
			expectValid:  false,
		},
	}

	runTestGroup(t, tests)
}

func TestDescriptionValidation(t *testing.T) {
	tests := []testDesc{
		{
			name:         "Valid Description",
			createCommit: createValidScopedCommit,
			expectValid:  true,
		},
		{
			name:         "Description At Max Length",
			createCommit: createCommitWithMsg("feat: " + strings.Repeat("a", 72)),
			expectValid:  true,
		},
		{
			name:         "Description Over Max Length",
			createCommit: createCommitWithMsg("feat: " + strings.Repeat("a", 73)),
			expectValid:  false,
		},
		{
			name:         "Empty Description",
			createCommit: createInvalidEmptyCommit,
			expectValid:  false,
		},
		{
			name:         "Whitespace Only Description",
			createCommit: createCommitWithMsg("feat:  "),
			expectValid:  false,
		},
		{
			name:         "Multiline Description",
			createCommit: createCommitWithMsg("feat: first line\nsecond line"),
			expectValid:  true,
		},
		{
			name:         "Description With Unicode",
			createCommit: createCommitWithMsg("feat: 你好世界"),
			expectValid:  true,
		},
	}

	runTestGroup(t, tests)
}
func TestMessageFormatValidation(t *testing.T) {
	tests := []testDesc{
		{
			name:         "Valid Format",
			createCommit: createValidScopedCommit,
			expectValid:  true,
		},
		{
			name:         "Multiple Colons",
			createCommit: createCommitWithMsg("feat: description: with colon"),
			expectValid:  true,
		},
		{
			name:         "No Space After Type",
			createCommit: createCommitWithMsg("feat:no space"),
			expectValid:  false,
		},
		{
			name:         "Extra Spaces After Colon",
			createCommit: createCommitWithMsg("feat:   extra spaces"),
			expectValid:  false,
		},
		{
			name:         "Invalid Format",
			createCommit: createInvalidCommit,
			expectValid:  false,
		},
		{
			name:         "Tab Instead of Space",
			createCommit: createCommitWithMsg("feat:\tdescription"),
			expectValid:  false,
		},
	}

	runTestGroup(t, tests)
}

func TestGitHubCompatibilityValidation(t *testing.T) {
	tests := []testDesc{
		{
			name:         "CRLF Line Endings",
			createCommit: createCommitWithMsg("feat: description\r\n"),
			expectValid:  true,
		},
		{
			name:         "Multiple Leading Newlines",
			createCommit: createCommitWithMsg("\n\nfeat: description"),
			expectValid:  false,
		},
		{
			name:         "Trailing Whitespace",
			createCommit: createCommitWithMsg("feat: description  "),
			expectValid:  true,
		},
		{
			name:         "GitHub Squash Merge",
			createCommit: createCommitWithMsg("\nfeat: description"),
			expectValid:  true,
		},
	}

	runTestGroup(t, tests)
}

// Commit creation helpers.
func createCommitWithMsg(msg string) func(*git.Repository) error {
	return func(repo *git.Repository) error {
		w, err := repo.Worktree()
		if err != nil {
			return err
		}

		_, err = w.Commit(msg, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "test",
				Email: "test@commiter.io",
				When:  time.Now(),
			},
		})

		return err
	}
}

func initRepo(path string) (*git.Repository, error) {
	repo, err := git.PlainInit(path, false)
	if err != nil {
		return nil, fmt.Errorf("initializing repository failed: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("getting worktree failed: %w", err)
	}

	err = os.WriteFile("test", []byte(""), 0o600)
	if err != nil {
		return nil, fmt.Errorf("creating test file failed: %w", err)
	}

	_, err = worktree.Add("test")
	if err != nil {
		return nil, fmt.Errorf("adding test file failed: %w", err)
	}

	return repo, nil
}

func createValidScopedCommit(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Commit("type(scope): description", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@commiter.io",
			When:  time.Now(),
		},
	})

	return err
}

func createValidBreakingCommit(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Commit("feat!: description", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@commiter.io",
			When:  time.Now(),
		},
	})

	return err
}

func createInvalidBreakingSymbolCommit(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Commit("feat$: description", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@commiter.io",
			When:  time.Now(),
		},
	})

	return err
}

func createValidScopedBreakingCommit(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Commit("feat(scope)!: description", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@commiter.io",
			When:  time.Now(),
		},
	})

	return err
}

func createInvalidScopedBreakingCommit(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Commit("feat!(scope): description", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@commiter.io",
			When:  time.Now(),
		},
	})

	return err
}

func createInvalidCommit(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Commit("invalid commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@commiter.io",
			When:  time.Now(),
		},
	})

	return err
}

func createInvalidEmptyCommit(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Commit("", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@commiter.io",
			When:  time.Now(),
		},
	})

	return err
}

func createValidCommitRegex(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Commit("type(valid-1): description", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@commiter.io",
			When:  time.Now(),
		},
	})

	return err
}

func createInvalidCommitRegex(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Commit("type(invalid-1): description", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@commiter.io",
			When:  time.Now(),
		},
	})

	return err
}

// Configuration helpers.
func runCompliance() (*model.Report, error) {
	gommit := &configuration.Gommit{
		Conventional: &configuration.Conventional{
			Types:  []string{"type", "feat", "fix"},
			Scopes: []string{"scope", "^valid"},
		},
	}

	return internal.Compliance(&model.Options{}, gommit)
}
