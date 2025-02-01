// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

//nolint:testpackage
package rules_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/rules"
	"github.com/itiquettemmitlint/internal/configuration"
)

//nolint:gocognit
func TestConventionalCommitRule(t *testing.T) {
	type testDesc struct {
		Name         string
		CreateCommit func(*git.Repository) error
		ExpectValid  bool
	}

	for _, test := range []testDesc{
		{
			Name:         "Valid",
			CreateCommit: createValidScopedCommit,
			ExpectValid:  true,
		},
		{
			Name:         "ValidBreaking",
			CreateCommit: createValidBreakingCommit,
			ExpectValid:  true,
		},
		{
			Name:         "InvalidBreakingSymbol",
			CreateCommit: createInvalidBreakingSymbolCommit,
			ExpectValid:  false,
		},
		{
			Name:         "ValidScopedBreaking",
			CreateCommit: createValidScopedBreakingCommit,
			ExpectValid:  true,
		},
		{
			Name:         "InvalidScopedBreaking",
			CreateCommit: createInvalidScopedBreakingCommit,
			ExpectValid:  false,
		},
		{
			Name:         "Invalid",
			CreateCommit: createInvalidCommit,
			ExpectValid:  false,
		},
		{
			Name:         "InvalidEmpty",
			CreateCommit: createInvalidEmptyCommit,
			ExpectValid:  false,
		},
	} {
		func(test testDesc) {
			t.Run(test.Name, func(testing *testing.T) {
				dir := t.TempDir()

				err := os.Chdir(dir)
				if err != nil {
					testing.Error(err)
				}

				repo, err := initRepo(dir)
				if err != nil {
					testing.Error(err)
				}

				err = test.CreateCommit(repo)
				if err != nil {
					testing.Error(err)
				}

				report, err := runCompliance()
				if err != nil {
					t.Error(err)
				}

				if test.ExpectValid {
					if !report.Valid() {
						testing.Error("Report is invalid with valid conventional commit")
					}
				} else {
					if report.Valid() {
						testing.Error("Report is valid with invalid conventional commit")
					}
				}
			})
		}(test)
	}
}

func TestValidateDCO(t *testing.T) {
	type testDesc struct {
		Name          string
		CommitMessage string
		ExpectValid   bool
	}

	for _, test := range []testDesc{
		{
			Name:          "Valid DCO",
			CommitMessage: "something nice\n\nSigned-off-by: Foo Bar <foobar@example.org>\n\n",
			ExpectValid:   true,
		},
		{
			Name:          "Valid DCO with CRLF",
			CommitMessage: "something nice\r\n\r\nSigned-off-by: Foo Bar <foobar@example.org>\r\n\r\n",
			ExpectValid:   true,
		},
		{
			Name:          "No DCO",
			CommitMessage: "something nice\n\nnot signed\n",
			ExpectValid:   false,
		},
	} {
		t.Run(test.Name, func(testing *testing.T) {
			var report model.Report

			commit := configuration.Gommit{Message: test.CommitMessage}
			report.AddCheck(rules.ValidateDCO(commit.Message))

			if test.ExpectValid {
				if !report.Valid() {
					testing.Error("Report is invalid with valid DCP")
				}
			} else {
				if report.Valid() {
					testing.Error("Report is valid with invalid DCO")
				}
			}
		})
	}
}

func TestValidRevisionRange(t *testing.T) {
	dir := t.TempDir()

	err := os.Chdir(dir)
	if err != nil {
		t.Error(err)
	}

	repo, err := initRepo(dir)
	if err != nil {
		t.Error(err)
	}

	revs, err := createValidCommitRange(repo)
	if err != nil {
		t.Fatal(err)
	}

	// Test with a valid revision range
	report, err := runComplianceRange(revs[0].String(), revs[len(revs)-1].String())
	if err != nil {
		t.Error(err)
	}

	if !report.Valid() {
		t.Error("Report is invalid with valid conventional commits")
	}

	// Test with HEAD as end of revision range
	report, err = runComplianceRange(revs[0].String(), "HEAD")
	if err != nil {
		t.Error(err)
	}

	if !report.Valid() {
		t.Error("Report is invalid with valid conventional commits")
	}

	// Test with empty end of revision range (should fail)
	_, err = runComplianceRange(revs[0].String(), "")
	if err == nil {
		t.Error("Invalid end of revision, got success, expecting failure")
	}

	// Test with empty start of revision (should fail)
	_, err = runComplianceRange("", "HEAD")
	if err == nil {
		t.Error("Invalid end of revision, got success, expecting failure")
	}

	// Test with start of revision not an ancestor of end of range (should fail)
	_, err = runComplianceRange(revs[1].String(), revs[0].String())
	if err == nil {
		t.Error("Invalid end of revision, got success, expecting failure")
	}
}

func createValidCommitRange(repo *git.Repository) ([]plumbing.Hash, error) {
	var revs []plumbing.Hash

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("getting worktree failed: %w", err)
	}

	for index := 0; index < 4; index++ {
		err := os.WriteFile("test", []byte(strconv.Itoa(index)), 0o600)
		if err != nil {
			return nil, fmt.Errorf("writing test file failed: %w", err)
		}

		_, err = worktree.Add("test")
		if err != nil {
			return nil, fmt.Errorf("git add failed: %w", err)
		}

		commit, err := worktree.Commit(fmt.Sprintf("type(scope): description %d", index), &git.CommitOptions{
			Author: &object.Signature{
				Name:  "test",
				Email: "test@commiter.io",
				When:  time.Now(),
			},
		})
		if err != nil {
			return nil, fmt.Errorf("git commit failed: %w", err)
		}

		revs = append(revs, commit)
	}

	return revs, nil
}

func runComplianceRange(id1, id2 string) (*model.Report, error) {
	gommit := &configuration.Gommit{
		Conventional: &configuration.Conventional{
			Types:  []string{"type"},
			Scopes: []string{"scope", "^valid"},
		},
	}

	return configuration.Compliance(&configuration.Options{
		RevisionRange: fmt.Sprintf("%s..%s", id1, id2),
	}, gommit)
}

func runCompliance() (*model.Report, error) {
	gommit := &configuration.Gommit{
		Conventional: &configuration.Conventional{
			Types:  []string{"type"},
			Scopes: []string{"scope", "^valid"},
		},
	}

	return configuration.Compliance(&configuration.Options{}, gommit)
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

func TestValidConventionalCommitRule(t *testing.T) {
	dir := t.TempDir()

	err := os.Chdir(dir)
	if err != nil {
		t.Error(err)
	}

	repo, err := initRepo(dir)
	if err != nil {
		t.Error(err)
	}

	err = createValidScopedCommit(repo)
	if err != nil {
		t.Error(err)
	}

	report, err := runCompliance()
	if err != nil {
		t.Error(err)
	}

	if !report.Valid() {
		t.Errorf("Report is invalid with valid conventional commit")
	}
}

func TestInvalidConventionalCommitRule(t *testing.T) {
	dir := t.TempDir()

	err := os.Chdir(dir)
	if err != nil {
		t.Error(err)
	}

	repo, err := initRepo(dir)
	if err != nil {
		t.Error(err)
	}

	err = createInvalidCommit(repo)
	if err != nil {
		t.Error(err)
	}

	report, err := runCompliance()
	if err != nil {
		t.Error(err)
	}

	if report.Valid() {
		t.Errorf("Report is valid with invalid conventional commit")
	}
}

func TestEmptyConventionalCommitRule(t *testing.T) {
	dir := t.TempDir()

	err := os.Chdir(dir)
	if err != nil {
		t.Error(err)
	}

	repo, err := initRepo(dir)
	if err != nil {
		t.Error(err)
	}

	err = createInvalidEmptyCommit(repo)
	if err != nil {
		t.Error(err)
	}

	report, err := runCompliance()
	if err != nil {
		t.Error(err)
	}

	if report.Valid() {
		t.Error("Report is valid with invalid conventional commit")
	}
}

func TestValidConventionalCommitRuleRegex(t *testing.T) {
	dir := t.TempDir()

	err := os.Chdir(dir)
	if err != nil {
		t.Error(err)
	}

	repo, err := initRepo(dir)
	if err != nil {
		t.Error(err)
	}

	err = createValidCommitRegex(repo)
	if err != nil {
		t.Error(err)
	}

	report, err := runCompliance()
	if err != nil {
		t.Error(err)
	}

	if !report.Valid() {
		t.Error("Report is invalid with valid conventional commit")
	}
}

func TestInvalidConventionalCommitRuleRegex(t *testing.T) {
	dir := t.TempDir()

	err := os.Chdir(dir)
	if err != nil {
		t.Error(err)
	}

	repo, err := initRepo(dir)
	if err != nil {
		t.Error(err)
	}

	err = createInvalidCommitRegex(repo)
	if err != nil {
		t.Error(err)
	}

	report, err := runCompliance()
	if err != nil {
		t.Error(err)
	}

	if report.Valid() {
		t.Error("Report is valid with invalid conventional commit")
	}
}
