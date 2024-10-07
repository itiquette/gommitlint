// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

//
//nolint:testpackage
package commit

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/janderssonse/gommitlint/internal/policy"
)

//nolint:gocognit
func TestConventionalCommitPolicy(t *testing.T) {
	type testDesc struct {
		Name         string
		CreateCommit func() error
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

				err = initRepo()
				if err != nil {
					testing.Error(err)
				}

				err = test.CreateCommit()
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
			var report policy.Report

			commit := Commit{msg: test.CommitMessage}
			report.AddCheck(commit.ValidateDCO())

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

func TestValidConventionalCommitPolicy(t *testing.T) {
	dir := t.TempDir()

	err := os.Chdir(dir)
	if err != nil {
		t.Error(err)
	}

	err = initRepo()
	if err != nil {
		t.Error(err)
	}

	err = createValidScopedCommit()
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

func TestInvalidConventionalCommitPolicy(t *testing.T) {
	dir := t.TempDir()

	err := os.Chdir(dir)
	if err != nil {
		t.Error(err)
	}

	err = initRepo()
	if err != nil {
		t.Error(err)
	}

	err = createInvalidCommit()
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

func TestEmptyConventionalCommitPolicy(t *testing.T) {
	dir := t.TempDir()

	err := os.Chdir(dir)
	if err != nil {
		t.Error(err)
	}

	err = initRepo()
	if err != nil {
		t.Error(err)
	}

	err = createInvalidEmptyCommit()
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

func TestValidConventionalCommitPolicyRegex(t *testing.T) {
	dir := t.TempDir()

	err := os.Chdir(dir)
	if err != nil {
		t.Error(err)
	}

	err = initRepo()
	if err != nil {
		t.Error(err)
	}

	err = createValidCommitRegex()
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

func TestInvalidConventionalCommitPolicyRegex(t *testing.T) {
	dir := t.TempDir()

	err := os.Chdir(dir)
	if err != nil {
		t.Error(err)
	}

	err = initRepo()
	if err != nil {
		t.Error(err)
	}

	err = createInvalidCommitRegex()
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

func TestValidRevisionRange(t *testing.T) {
	dir := t.TempDir()

	err := os.Chdir(dir)
	if err != nil {
		t.Error(err)
	}

	err = initRepo()
	if err != nil {
		t.Error(err)
	}

	revs, err := createValidCommitRange()
	if err != nil {
		t.Fatal(err)
	}

	// Test with a valid revision range
	report, err := runComplianceRange(revs[0], revs[len(revs)-1])
	if err != nil {
		t.Error(err)
	}

	if !report.Valid() {
		t.Error("Report is invalid with valid conventional commits")
	}

	// Test with HEAD as end of revision range
	report, err = runComplianceRange(revs[0], "HEAD")
	if err != nil {
		t.Error(err)
	}

	if !report.Valid() {
		t.Error("Report is invalid with valid conventional commits")
	}

	// Test with empty end of revision range (should fail)
	_, err = runComplianceRange(revs[0], "")
	if err == nil {
		t.Error("Invalid end of revision, got success, expecting failure")
	}

	// Test with empty start of revision (should fail)
	_, err = runComplianceRange("", "HEAD")
	if err == nil {
		t.Error("Invalid end of revision, got success, expecting failure")
	}

	// Test with start of revision not an ancestor of end of range (should fail)
	_, err = runComplianceRange(revs[1], revs[0])
	if err == nil {
		t.Error("Invalid end of revision, got success, expecting failure")
	}
}

func createValidCommitRange() ([]string, error) {
	var revs []string

	for index := 0; index < 4; index++ {
		err := os.WriteFile("test", []byte(strconv.Itoa(index)), 0o600)
		if err != nil {
			return nil, fmt.Errorf("writing test file failed: %w", err)
		}

		_, err = exec.Command("git", "add", "test").Output()
		if err != nil {
			return nil, fmt.Errorf("git add failed: %w", err)
		}

		//TO-DO replace me with a non real cli call, ie set up with go-git library
		//nolint:gosec
		_, err = exec.Command("git", "-c", "user.name='test'", "-c", "user.email='test@janderssonse.io'", "commit", "-m", fmt.Sprintf("type(scope): description %d", index)).Output()
		if err != nil {
			return nil, fmt.Errorf("git commit failed: %w", err)
		}

		id, err := exec.Command("git", "rev-parse", "HEAD").Output()
		if err != nil {
			return nil, fmt.Errorf("rev-parse failed: %w", err)
		}

		revs = append(revs, strings.TrimSpace(string(id)))
	}

	return revs, nil
}

func runComplianceRange(id1, id2 string) (*policy.Report, error) {
	commit := &Commit{
		Conventional: &Conventional{
			Types:  []string{"type"},
			Scopes: []string{"scope", "^valid"},
		},
	}

	return commit.Compliance(&policy.Options{
		RevisionRange: fmt.Sprintf("%s..%s", id1, id2),
	})
}

func runCompliance() (*policy.Report, error) {
	commit := &Commit{
		Conventional: &Conventional{
			Types:  []string{"type"},
			Scopes: []string{"scope", "^valid"},
		},
	}

	return commit.Compliance(&policy.Options{})
}

func initRepo() error {
	_, err := exec.Command("git", "init").Output()
	if err != nil {
		return err
	}

	_, err = exec.Command("touch", "test").Output()
	if err != nil {
		return err
	}

	_, err = exec.Command("git", "add", "test").Output()

	return err
}

func createValidScopedCommit() error {
	_, err := exec.Command("git", "-c", "user.name='test'", "-c", "user.email='test@janderssonse.io'", "commit", "-m", "type(scope): description").Output()

	return err
}

func createValidBreakingCommit() error {
	_, err := exec.Command("git", "-c", "user.name='test'", "-c", "user.email='test@janderssonse.io'", "commit", "-m", "feat!: description").Output()

	return err
}

func createInvalidBreakingSymbolCommit() error {
	_, err := exec.Command("git", "-c", "user.name='test'", "-c", "user.email='test@janderssonse.io'", "commit", "-m", "feat$: description").Output()

	return err
}

func createValidScopedBreakingCommit() error {
	_, err := exec.Command("git", "-c", "user.name='test'", "-c", "user.email='test@janderssonse.io'", "commit", "-m", "feat(scope)!: description").Output()

	return err
}

func createInvalidScopedBreakingCommit() error {
	_, err := exec.Command("git", "-c", "user.name='test'", "-c", "user.email='test@janderssonse.io'", "commit", "-m", "feat!(scope): description").Output()

	return err
}

func createInvalidCommit() error {
	_, err := exec.Command("git", "-c", "user.name='test'", "-c", "user.email='test@janderssonse.io'", "commit", "-m", "invalid commit").Output()

	return err
}

func createInvalidEmptyCommit() error {
	_, err := exec.Command("git", "-c", "user.name='test'", "-c", "user.email='test@janderssonse.io'", "commit", "--allow-empty-message", "-m", "").Output()

	return err
}

func createValidCommitRegex() error {
	_, err := exec.Command("git", "-c", "user.name='test'", "-c", "user.email='test@janderssonse.io'", "commit", "-m", "type(valid-1): description").Output()

	return err
}

func createInvalidCommitRegex() error {
	_, err := exec.Command("git", "-c", "user.name='test'", "-c", "user.email='test@janderssonse.io'", "commit", "-m", "type(invalid-1): description").Output()

	return err
}
