// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/stretchr/testify/require"
)

func TestValidateTarget(t *testing.T) {
	tests := []struct {
		name         string
		target       ValidationTarget
		commitRules  []domain.CommitRule
		repoRules    []domain.RepositoryRule
		expectError  bool
		expectedType string
		description  string
	}{
		{
			name: "message validation",
			target: ValidationTarget{
				Type:   "message",
				Source: "/tmp/test-message.txt",
				Target: "",
			},
			commitRules: []domain.CommitRule{&mockCommitRule{name: "Subject"}},
			repoRules:   []domain.RepositoryRule{},
			expectError: true, // Will fail because file doesn't exist
			description: "should handle message validation",
		},
		{
			name: "commit validation",
			target: ValidationTarget{
				Type:   "commit",
				Source: "HEAD",
				Target: "",
			},
			commitRules: []domain.CommitRule{&mockCommitRule{name: "Subject"}},
			repoRules:   []domain.RepositoryRule{},
			expectError: false,
			description: "should handle commit validation",
		},
		{
			name: "range validation",
			target: ValidationTarget{
				Type:   "range",
				Source: "main",
				Target: "HEAD",
			},
			commitRules: []domain.CommitRule{&mockCommitRule{name: "Subject"}},
			repoRules:   []domain.RepositoryRule{},
			expectError: false,
			description: "should handle range validation",
		},
		{
			name: "count validation",
			target: ValidationTarget{
				Type:   "count",
				Source: "3",
				Target: "",
			},
			commitRules: []domain.CommitRule{&mockCommitRule{name: "Subject"}},
			repoRules:   []domain.RepositoryRule{},
			expectError: false,
			description: "should handle count validation",
		},
		{
			name: "unknown target type",
			target: ValidationTarget{
				Type:   "unknown",
				Source: "test",
				Target: "",
			},
			commitRules: []domain.CommitRule{},
			repoRules:   []domain.RepositoryRule{},
			expectError: true,
			description: "should reject unknown target types",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			repo := &mockRepository{
				commits: map[string]domain.Commit{
					"HEAD": {
						Hash:    "abc123",
						Subject: "Test commit",
						Message: "Test commit",
					},
				},
				commitRanges: map[string][]domain.Commit{
					"main..HEAD": {
						{Hash: "abc123", Subject: "Test commit", Message: "Test commit"},
						{Hash: "def456", Subject: "Another commit", Message: "Another commit"},
					},
					"HEAD~2..HEAD": { // For count=3 (creates HEAD~2..HEAD)
						{Hash: "abc123", Subject: "Test commit 1", Message: "Test commit 1"},
						{Hash: "def456", Subject: "Test commit 2", Message: "Test commit 2"},
						{Hash: "ghi789", Subject: "Test commit 3", Message: "Test commit 3"},
					},
				},
			}
			cfg := config.Config{}
			logger := &mockLogger{}

			report, err := ValidateTarget(ctx, testCase.target, testCase.commitRules, testCase.repoRules, repo, cfg, logger)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, report, testCase.description)
			}
		})
	}
}

func TestExecuteMessageValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func() (string, func())
		expectError bool
		description string
	}{
		{
			name: "valid message file",
			setupFile: func() (string, func()) {
				tmpFile, err := os.CreateTemp("", "test-message-*.txt")
				if err != nil {
					panic(err)
				}
				_, err = tmpFile.WriteString("Add new feature\n\nThis adds a new feature for testing")
				if err != nil {
					panic(err)
				}
				tmpFile.Close()

				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			expectError: false,
			description: "should validate message from file",
		},
		{
			name: "non-existent file",
			setupFile: func() (string, func()) {
				return "/nonexistent/file.txt", func() {}
			},
			expectError: true,
			description: "should fail for non-existent file",
		},
		{
			name: "stdin input",
			setupFile: func() (string, func()) {
				// Note: This test is limited because we can't easily mock stdin
				return "-", func() {}
			},
			expectError: true, // Will fail trying to read from stdin in test
			description: "should handle stdin input",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			filePath, cleanup := testCase.setupFile()
			defer cleanup()

			rules := []domain.CommitRule{&mockCommitRule{name: "Subject"}}
			cfg := config.Config{}
			logger := &mockLogger{}

			report, err := executeMessageValidation(filePath, rules, cfg, logger)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, report, testCase.description)
			}
		})
	}
}

func TestExecuteCommitValidation(t *testing.T) {
	tests := []struct {
		name        string
		ref         string
		setupRepo   func(*mockRepository)
		expectError bool
		description string
	}{
		{
			name: "valid commit",
			ref:  "HEAD",
			setupRepo: func(repo *mockRepository) {
				repo.commits["HEAD"] = domain.Commit{
					Hash:    "abc123",
					Subject: "Add feature",
					Body:    "Detailed description",
				}
			},
			expectError: false,
			description: "should validate existing commit",
		},
		{
			name: "non-existent commit",
			ref:  "nonexistent",
			setupRepo: func(_ *mockRepository) {
				// Don't add the commit to mock repo
			},
			expectError: true,
			description: "should fail for non-existent commit",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			repo := &mockRepository{commits: make(map[string]domain.Commit)}
			testCase.setupRepo(repo)

			commitRules := []domain.CommitRule{&mockCommitRule{name: "Subject"}}
			repoRules := []domain.RepositoryRule{&mockRepoRule{name: "BranchAhead"}}
			cfg := config.Config{}
			logger := &mockLogger{}

			report, err := executeCommitValidation(ctx, testCase.ref, commitRules, repoRules, repo, cfg, logger)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, report, testCase.description)
			}
		})
	}
}

func TestExecuteRangeValidation(t *testing.T) {
	tests := []struct {
		name        string
		fromRef     string
		toRef       string
		setupRepo   func(*mockRepository)
		expectError bool
		description string
	}{
		{
			name:    "valid range",
			fromRef: "main",
			toRef:   "HEAD",
			setupRepo: func(repo *mockRepository) {
				repo.commitRanges = map[string][]domain.Commit{
					"main..HEAD": {
						{Hash: "abc123", Subject: "First commit"},
						{Hash: "def456", Subject: "Second commit"},
					},
				}
			},
			expectError: false,
			description: "should validate commit range",
		},
		{
			name:    "invalid range",
			fromRef: "nonexistent",
			toRef:   "HEAD",
			setupRepo: func(_ *mockRepository) {
				// Don't add the range to mock repo
			},
			expectError: true,
			description: "should fail for invalid range",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			repo := &mockRepository{
				commits:      make(map[string]domain.Commit),
				commitRanges: make(map[string][]domain.Commit),
			}
			testCase.setupRepo(repo)

			commitRules := []domain.CommitRule{&mockCommitRule{name: "Subject"}}
			repoRules := []domain.RepositoryRule{&mockRepoRule{name: "BranchAhead"}}
			cfg := config.Config{}
			logger := &mockLogger{}

			report, err := executeRangeValidation(ctx, testCase.fromRef, testCase.toRef, commitRules, repoRules, repo, cfg, logger)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, report, testCase.description)
			}
		})
	}
}

func TestExecuteCountValidation(t *testing.T) {
	tests := []struct {
		name        string
		countStr    string
		setupRepo   func(*mockRepository)
		expectError bool
		description string
	}{
		{
			name:     "single commit count",
			countStr: "1",
			setupRepo: func(repo *mockRepository) {
				repo.commits["HEAD"] = domain.Commit{
					Hash:    "abc123",
					Subject: "Single commit",
				}
			},
			expectError: false,
			description: "should handle single commit count",
		},
		{
			name:     "multiple commit count",
			countStr: "3",
			setupRepo: func(repo *mockRepository) {
				repo.commitRanges = map[string][]domain.Commit{
					"HEAD~2..HEAD": {
						{Hash: "abc123", Subject: "First"},
						{Hash: "def456", Subject: "Second"},
						{Hash: "ghi789", Subject: "Third"},
					},
				}
			},
			expectError: false,
			description: "should handle multiple commit count",
		},
		{
			name:        "invalid count string",
			countStr:    "invalid",
			setupRepo:   func(_ *mockRepository) {},
			expectError: true,
			description: "should fail for invalid count string",
		},
		{
			name:        "negative count",
			countStr:    "-1",
			setupRepo:   func(_ *mockRepository) {},
			expectError: true,
			description: "should fail for negative count",
		},
		{
			name:        "zero count",
			countStr:    "0",
			setupRepo:   func(_ *mockRepository) {},
			expectError: true,
			description: "should fail for zero count",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			repo := &mockRepository{
				commits:      make(map[string]domain.Commit),
				commitRanges: make(map[string][]domain.Commit),
			}
			testCase.setupRepo(repo)

			commitRules := []domain.CommitRule{&mockCommitRule{name: "Subject"}}
			repoRules := []domain.RepositoryRule{&mockRepoRule{name: "BranchAhead"}}
			cfg := config.Config{}
			logger := &mockLogger{}

			report, err := executeCountValidation(ctx, testCase.countStr, commitRules, repoRules, repo, cfg, logger)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, report, testCase.description)
			}
		})
	}
}

func TestValidateMessageContent(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		rules       []domain.CommitRule
		expectError bool
		description string
	}{
		{
			name:        "valid message",
			message:     "Add new feature\n\nDetailed description",
			rules:       []domain.CommitRule{&mockCommitRule{name: "Subject"}},
			expectError: false,
			description: "should validate message content",
		},
		{
			name:        "empty message",
			message:     "",
			rules:       []domain.CommitRule{&mockCommitRule{name: "Subject"}},
			expectError: true,
			description: "should reject empty message",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{}

			report, err := ValidateMessageContent(testCase.message, testCase.rules, cfg)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, report, testCase.description)
			}
		})
	}
}

func TestValidateSingleCommit(t *testing.T) {
	tests := []struct {
		name        string
		commit      domain.Commit
		expectSkip  bool
		description string
	}{
		{
			name: "regular commit",
			commit: domain.Commit{
				Hash:          "abc123",
				Subject:       "Add feature",
				Body:          "Description",
				IsMergeCommit: false,
			},
			expectSkip:  false,
			description: "should validate regular commit",
		},
		{
			name: "merge commit",
			commit: domain.Commit{
				Hash:          "def456",
				Subject:       "Merge branch 'feature'",
				IsMergeCommit: true,
			},
			expectSkip:  true,
			description: "should skip merge commit validation",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			commitRules := []domain.CommitRule{&mockCommitRule{name: "Subject"}}
			repoRules := []domain.RepositoryRule{&mockRepoRule{name: "BranchAhead"}}
			repo := &mockRepository{}
			cfg := config.Config{}

			report, err := ValidateSingleCommit(testCase.commit, commitRules, repoRules, repo, cfg)

			require.NoError(t, err, testCase.description)
			require.NotEmpty(t, report, testCase.description)
			require.Len(t, report.Commits, 1, "should have one commit report")

			if testCase.expectSkip {
				// For merge commits, should have empty errors (skipped)
				require.True(t, report.Commits[0].Passed, "merge commit should be skipped")
			}
		})
	}
}

func TestValidateMultipleCommits(t *testing.T) {
	tests := []struct {
		name        string
		commits     []domain.Commit
		description string
	}{
		{
			name: "multiple regular commits",
			commits: []domain.Commit{
				{Hash: "abc123", Subject: "First commit", IsMergeCommit: false},
				{Hash: "def456", Subject: "Second commit", IsMergeCommit: false},
			},
			description: "should validate multiple commits",
		},
		{
			name: "mixed commits with merge",
			commits: []domain.Commit{
				{Hash: "abc123", Subject: "Regular commit", IsMergeCommit: false},
				{Hash: "def456", Subject: "Merge commit", IsMergeCommit: true},
				{Hash: "ghi789", Subject: "Another regular", IsMergeCommit: false},
			},
			description: "should filter out merge commits",
		},
		{
			name:        "empty commits list",
			commits:     []domain.Commit{},
			description: "should handle empty commits list",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			commitRules := []domain.CommitRule{&mockCommitRule{name: "Subject"}}
			repoRules := []domain.RepositoryRule{&mockRepoRule{name: "BranchAhead"}}
			repo := &mockRepository{}
			cfg := config.Config{}

			report, err := ValidateMultipleCommits(testCase.commits, commitRules, repoRules, repo, cfg)

			require.NoError(t, err, testCase.description)
			require.NotEmpty(t, report, testCase.description)
		})
	}
}

func TestReadMessageFile(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func() (string, func())
		expectedMsg string
		expectError bool
		description string
	}{
		{
			name: "read from regular file",
			setupFile: func() (string, func()) {
				tmpFile, err := os.CreateTemp("", "test-message-*.txt")
				if err != nil {
					panic(err)
				}
				content := "Test commit message\n\nDetailed description"
				_, err = tmpFile.WriteString(content)
				if err != nil {
					panic(err)
				}
				tmpFile.Close()

				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			expectedMsg: "Test commit message\n\nDetailed description",
			expectError: false,
			description: "should read content from file",
		},
		{
			name: "non-existent file",
			setupFile: func() (string, func()) {
				return "/nonexistent/file.txt", func() {}
			},
			expectError: true,
			description: "should fail for non-existent file",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			filePath, cleanup := testCase.setupFile()
			defer cleanup()

			message, err := readMessageFile(filePath)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.Equal(t, testCase.expectedMsg, message, testCase.description)
			}
		})
	}
}

func TestParseCommitCount(t *testing.T) {
	tests := []struct {
		name          string
		countStr      string
		expectedCount int
		expectError   bool
		description   string
	}{
		{
			name:          "valid positive count",
			countStr:      "5",
			expectedCount: 5,
			expectError:   false,
			description:   "should parse valid positive number",
		},
		{
			name:        "zero count",
			countStr:    "0",
			expectError: true,
			description: "should reject zero count",
		},
		{
			name:        "negative count",
			countStr:    "-1",
			expectError: true,
			description: "should reject negative count",
		},
		{
			name:        "invalid string",
			countStr:    "abc",
			expectError: true,
			description: "should reject non-numeric string",
		},
		{
			name:        "empty string",
			countStr:    "",
			expectError: true,
			description: "should reject empty string",
		},
		{
			name:          "large valid count",
			countStr:      "100",
			expectedCount: 100,
			expectError:   false,
			description:   "should parse large valid number",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			count, err := parseCommitCount(testCase.countStr)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.Equal(t, testCase.expectedCount, count, testCase.description)
			}
		})
	}
}

func TestValidateTarget_ContextCancellation(t *testing.T) {
	tests := []struct {
		name        string
		target      ValidationTarget
		setupCtx    func() (context.Context, context.CancelFunc)
		description string
	}{
		{
			name: "context cancelled during commit validation",
			target: ValidationTarget{
				Type:   "commit",
				Source: "HEAD",
			},
			setupCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				time.Sleep(2 * time.Nanosecond) // Ensure timeout

				return ctx, cancel
			},
			description: "should handle context cancellation",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := testCase.setupCtx()
			defer cancel()

			repo := &mockRepository{commits: make(map[string]domain.Commit)}
			commitRules := []domain.CommitRule{&mockCommitRule{name: "Subject"}}
			repoRules := []domain.RepositoryRule{&mockRepoRule{name: "BranchAhead"}}
			cfg := config.Config{}
			logger := &mockLogger{}

			_, err := ValidateTarget(ctx, testCase.target, commitRules, repoRules, repo, cfg, logger)

			require.Error(t, err, testCase.description)
			require.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled), "should be context error")
		})
	}
}

// Mock implementations for testing

type mockCommitRule struct {
	name string
}

func (m *mockCommitRule) Name() string {
	return m.name
}

func (m *mockCommitRule) Validate(_ domain.Commit, _ config.Config) []domain.ValidationError {
	return []domain.ValidationError{}
}

type mockRepoRule struct {
	name string
}

func (m *mockRepoRule) Name() string {
	return m.name
}

func (m *mockRepoRule) Validate(_ domain.Commit, _ domain.Repository, _ config.Config) []domain.ValidationError {
	return []domain.ValidationError{}
}

type mockRepository struct {
	commits      map[string]domain.Commit
	commitRanges map[string][]domain.Commit
}

func (m *mockRepository) GetCommit(_ context.Context, ref string) (domain.Commit, error) {
	if commit, exists := m.commits[ref]; exists {
		return commit, nil
	}

	return domain.Commit{}, domain.New("repository", "commit_not_found", "commit not found: "+ref)
}

func (m *mockRepository) GetCommitRange(_ context.Context, fromRef, toRef string) ([]domain.Commit, error) {
	rangeKey := fromRef + ".." + toRef
	if commits, exists := m.commitRanges[rangeKey]; exists {
		return commits, nil
	}

	return nil, domain.New("repository", "range_not_found", "range not found: "+rangeKey)
}

func (m *mockRepository) GetHeadCommits(_ context.Context, count int) ([]domain.Commit, error) {
	// For count=1, return HEAD
	if count == 1 {
		if commit, exists := m.commits["HEAD"]; exists {
			return []domain.Commit{commit}, nil
		}

		return nil, domain.New("repository", "commit_not_found", "HEAD not found")
	}

	// For count>1, use the appropriate range
	rangeKey := fmt.Sprintf("HEAD~%d..HEAD", count-1)
	if commits, exists := m.commitRanges[rangeKey]; exists {
		return commits, nil
	}

	// Fallback: return empty if range not found
	return []domain.Commit{}, nil
}

func (m *mockRepository) GetCommitsAheadCount(_ context.Context, _ string) (int, error) {
	// Simple mock implementation
	return 0, nil
}

type mockLogger struct{}

func (m *mockLogger) Log(_ string, _ string, _ ...interface{}) {}
func (m *mockLogger) Debug(_ string, _ ...interface{})         {}
func (m *mockLogger) Info(_ string, _ ...interface{})          {}
func (m *mockLogger) Error(_ string, _ ...interface{})         {}

// Ensure mocks implement the interfaces.
var _ domain.CommitRule = (*mockCommitRule)(nil)
var _ domain.RepositoryRule = (*mockRepoRule)(nil)
var _ domain.Repository = (*mockRepository)(nil)
var _ domain.Logger = (*mockLogger)(nil)
