// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validate_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCommitReader implements the domain.CommitReader interface for testing.
type mockCommitReader struct {
	commits map[string]*domain.CommitInfo
}

func (m *mockCommitReader) GetCommit(hash string) (*domain.CommitInfo, error) {
	if commit, ok := m.commits[hash]; ok {
		return commit, nil
	}

	return nil, &domain.ValidationError{Message: "commit not found"}
}

// mockHistoryReader implements the domain.CommitHistoryReader interface for testing.
type mockHistoryReader struct {
	headCommits []*domain.CommitInfo
	ranges      map[string]map[string][]*domain.CommitInfo
}

func (m *mockHistoryReader) GetHeadCommits(count int) ([]*domain.CommitInfo, error) {
	if count >= len(m.headCommits) {
		return m.headCommits, nil
	}

	return m.headCommits[:count], nil
}

func (m *mockHistoryReader) GetCommitRange(fromHash, toHash string) ([]*domain.CommitInfo, error) {
	if toRanges, ok := m.ranges[toHash]; ok {
		if commits, ok := toRanges[fromHash]; ok {
			return commits, nil
		}
	}

	return nil, &domain.ValidationError{Message: "range not found"}
}

// mockInfoProvider implements the domain.RepositoryInfoProvider interface for testing.
type mockInfoProvider struct {
	currentBranch string
	repoName      string
	isValid       bool
}

func (m *mockInfoProvider) GetCurrentBranch() (string, error) {
	return m.currentBranch, nil
}

func (m *mockInfoProvider) GetRepositoryName() string {
	return m.repoName
}

func (m *mockInfoProvider) IsValid() bool {
	return m.isValid
}

// mockValidationEngine is a simple implementation of a validation engine for testing.
type mockValidationEngine struct{}

func (m *mockValidationEngine) ValidateCommit(_ context.Context, commit *domain.CommitInfo) domain.CommitResult {
	return domain.CommitResult{
		CommitInfo: commit,
		Passed:     true,
	}
}

func (m *mockValidationEngine) ValidateCommits(ctx context.Context, commits []*domain.CommitInfo) *domain.ValidationResults {
	results := domain.NewValidationResults()

	for _, commit := range commits {
		result := m.ValidateCommit(ctx, commit)
		results.AddCommitResult(result)
	}

	return results
}

func TestValidationService_ValidateCommit(t *testing.T) {
	// Create mocks
	mockReader := &mockCommitReader{
		commits: map[string]*domain.CommitInfo{
			"abc123": {
				Hash:    "abc123",
				Subject: "Test commit",
				Body:    "Test body",
			},
			"HEAD": {
				Hash:    "def456",
				Subject: "Head commit",
			},
		},
	}

	mockHistory := &mockHistoryReader{}
	mockInfo := &mockInfoProvider{
		currentBranch: "main",
		repoName:      "gommitlint",
		isValid:       true,
	}

	// Create mock engine
	mockEngine := &mockValidationEngine{}

	// Create service with mock engine
	service := validate.NewValidationService(
		mockEngine,
		mockReader,
		mockHistory,
		mockInfo,
	)

	// Test ValidateCommit
	t.Run("Validate existing commit", func(t *testing.T) {
		result, err := service.ValidateCommit(context.Background(), "abc123")
		require.NoError(t, err)
		assert.Equal(t, "abc123", result.CommitInfo.Hash)
	})

	t.Run("Validate non-existent commit", func(t *testing.T) {
		_, err := service.ValidateCommit(context.Background(), "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "commit not found")
	})
}

func TestValidationService_ValidateHeadCommits(t *testing.T) {
	// Create test commits
	commits := []*domain.CommitInfo{
		{
			Hash:          "abc123",
			Subject:       "Test commit 1",
			IsMergeCommit: false,
		},
		{
			Hash:          "def456",
			Subject:       "Merge branch",
			IsMergeCommit: true,
		},
		{
			Hash:          "ghi789",
			Subject:       "Test commit 3",
			IsMergeCommit: false,
		},
	}

	// Create mocks
	mockReader := &mockCommitReader{}
	mockHistory := &mockHistoryReader{
		headCommits: commits,
	}
	mockInfo := &mockInfoProvider{}

	// Create mock engine
	mockEngine := &mockValidationEngine{}

	// Create service with mock engine
	service := validate.NewValidationService(
		mockEngine,
		mockReader,
		mockHistory,
		mockInfo,
	)

	// Test ValidateHeadCommits
	t.Run("Validate head commits without filtering", func(t *testing.T) {
		results, err := service.ValidateHeadCommits(context.Background(), 3, false)
		require.NoError(t, err)
		assert.Equal(t, 3, results.Count())
	})

	t.Run("Validate head commits with merge filtering", func(t *testing.T) {
		results, err := service.ValidateHeadCommits(context.Background(), 3, true)
		require.NoError(t, err)
		assert.Equal(t, 2, results.Count())
	})
}

func TestValidationService_ValidateCommitRange(t *testing.T) {
	// Create test commits
	commits := []*domain.CommitInfo{
		{
			Hash:          "abc123",
			Subject:       "Test commit 1",
			IsMergeCommit: false,
		},
		{
			Hash:          "def456",
			Subject:       "Merge branch",
			IsMergeCommit: true,
		},
		{
			Hash:          "ghi789",
			Subject:       "Test commit 3",
			IsMergeCommit: false,
		},
	}

	// Create mocks
	mockReader := &mockCommitReader{}
	mockHistory := &mockHistoryReader{
		ranges: map[string]map[string][]*domain.CommitInfo{
			"master": {
				"feature": commits,
			},
		},
	}
	mockInfo := &mockInfoProvider{}

	// Create mock engine
	mockEngine := &mockValidationEngine{}

	// Create service with mock engine
	service := validate.NewValidationService(
		mockEngine,
		mockReader,
		mockHistory,
		mockInfo,
	)

	// Test ValidateCommitRange
	t.Run("Validate commit range without filtering", func(t *testing.T) {
		results, err := service.ValidateCommitRange(context.Background(), "feature", "master", false)
		require.NoError(t, err)
		assert.Equal(t, 3, results.Count())
	})

	t.Run("Validate commit range with merge filtering", func(t *testing.T) {
		results, err := service.ValidateCommitRange(context.Background(), "feature", "master", true)
		require.NoError(t, err)
		assert.Equal(t, 2, results.Count())
	})

	t.Run("Validate non-existent range", func(t *testing.T) {
		_, err := service.ValidateCommitRange(context.Background(), "nonexistent", "master", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "range not found")
	})
}

func TestValidationService_ValidateWithOptions(t *testing.T) {
	// Create mocks with test data
	testCommit := &domain.CommitInfo{
		Hash:          "abc123",
		Subject:       "Test commit",
		Body:          "Test body",
		IsMergeCommit: false,
	}

	// Create mocks
	mockReader := &mockCommitReader{
		commits: map[string]*domain.CommitInfo{
			"abc123": testCommit,
			"HEAD":   testCommit,
		},
	}

	mockHistory := &mockHistoryReader{
		headCommits: []*domain.CommitInfo{testCommit},
		ranges: map[string]map[string][]*domain.CommitInfo{
			"to": {
				"from": []*domain.CommitInfo{testCommit},
			},
		},
	}

	mockInfo := &mockInfoProvider{}

	// Create mock engine
	mockEngine := &mockValidationEngine{}

	// Create service with mock engine
	service := validate.NewValidationService(
		mockEngine,
		mockReader,
		mockHistory,
		mockInfo,
	)

	// Test different options
	t.Run("Validate specific commit", func(t *testing.T) {
		opts := validate.ValidationOptions{
			CommitHash: "abc123",
		}
		results, err := service.ValidateWithOptions(context.Background(), opts)
		require.NoError(t, err)
		assert.Equal(t, 1, results.Count())
	})

	t.Run("Validate head commits", func(t *testing.T) {
		opts := validate.ValidationOptions{
			CommitCount:      1,
			SkipMergeCommits: false,
		}
		results, err := service.ValidateWithOptions(context.Background(), opts)
		require.NoError(t, err)
		assert.Equal(t, 1, results.Count())
	})

	t.Run("Validate commit range", func(t *testing.T) {
		opts := validate.ValidationOptions{
			FromHash:         "from",
			ToHash:           "to",
			SkipMergeCommits: false,
		}
		results, err := service.ValidateWithOptions(context.Background(), opts)
		require.NoError(t, err)
		assert.Equal(t, 1, results.Count())
	})

	t.Run("Default to HEAD", func(t *testing.T) {
		opts := validate.ValidationOptions{}
		results, err := service.ValidateWithOptions(context.Background(), opts)
		require.NoError(t, err)
		assert.Equal(t, 1, results.Count())
	})
}
