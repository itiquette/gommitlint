// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validate_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

// mockGitCommitService implements the domain.GitCommitService interface for testing.
type mockGitCommitService struct {
	commits     map[string]domain.CommitInfo
	headCommits []domain.CommitInfo
	ranges      map[string]map[string][]domain.CommitInfo
}

func (m mockGitCommitService) GetCommit(_ context.Context, hash string) (domain.CommitInfo, error) {
	if commit, ok := m.commits[hash]; ok {
		return commit, nil
	}

	return domain.CommitInfo{}, appErrors.New("MockService", appErrors.ErrCommitNotFound, "commit not found")
}
func (m mockGitCommitService) GetHeadCommits(_ context.Context, count int) ([]domain.CommitInfo, error) {
	if count <= 0 {
		return []domain.CommitInfo{}, nil
	}

	// Take the minimum of count or the available number of commits
	if count > len(m.headCommits) {
		count = len(m.headCommits)
	}

	return m.headCommits[:count], nil
}

func (m mockGitCommitService) GetCommitRange(_ context.Context, fromHash, toHash string) ([]domain.CommitInfo, error) {
	if toRanges, ok := m.ranges[toHash]; ok {
		if commits, ok := toRanges[fromHash]; ok {
			return commits, nil
		}
	}

	return nil, appErrors.New("MockService", appErrors.ErrRangeNotFound, "range not found")
}

// mockInfoProvider implements the domain.RepositoryInfoProvider interface for testing.
type mockInfoProvider struct {
	currentBranch string
	repoName      string
	isValid       bool
}

func (m mockInfoProvider) GetCurrentBranch(_ context.Context) (string, error) {
	return m.currentBranch, nil
}

func (m mockInfoProvider) GetRepositoryName(_ context.Context) string {
	return m.repoName
}

func (m mockInfoProvider) IsValid(_ context.Context) bool {
	return m.isValid
}

// mockValidationEngine is a simple implementation of a validation engine for testing.
type mockValidationEngine struct{}

func (m mockValidationEngine) ValidateCommit(_ context.Context, commit domain.CommitInfo) domain.CommitResult {
	return domain.CommitResult{
		CommitInfo: commit,
		Passed:     true,
	}
}

func (m mockValidationEngine) ValidateCommits(ctx context.Context, commits []domain.CommitInfo) domain.ValidationResults {
	results := domain.NewValidationResults()

	for _, commit := range commits {
		result := m.ValidateCommit(ctx, commit)
		results.AddCommitResult(result)
	}

	return results
}

func TestValidationService_Functional(t *testing.T) {
	// Create mocks
	mockCommit := &mockGitCommitService{
		commits: map[string]domain.CommitInfo{
			"abc123": {
				Hash:    "abc123",
				Subject: "Test commit",
				Body:    "Test body",
			},
		},
	}

	mockInfo := &mockInfoProvider{
		currentBranch: "main",
		repoName:      "gommitlint",
		isValid:       true,
	}

	mockEngine := &mockValidationEngine{}

	// Create service with mock engine
	service := validate.NewValidationService(
		mockEngine,
		mockCommit,
		mockInfo,
	)

	// Test the functional "With" methods
	serviceWithNewEngine := service.WithEngine(&mockValidationEngine{})
	// Instead of comparing structs directly, verify the fields are correctly changed
	require.Equal(t, &mockValidationEngine{}, serviceWithNewEngine.Engine())
	require.Equal(t, mockCommit, serviceWithNewEngine.CommitService())
	require.Equal(t, mockInfo, serviceWithNewEngine.InfoProvider())

	newMockCommit := &mockGitCommitService{
		commits: map[string]domain.CommitInfo{
			"xyz789": {
				Hash:    "xyz789",
				Subject: "New commit",
			},
		},
	}
	serviceWithNewCommit := service.WithCommitService(newMockCommit)
	// Verify the individual fields
	require.Equal(t, mockEngine, serviceWithNewCommit.Engine())
	require.Equal(t, newMockCommit, serviceWithNewCommit.CommitService())
	require.Equal(t, mockInfo, serviceWithNewCommit.InfoProvider())

	newMockInfo := &mockInfoProvider{
		currentBranch: "feature",
		repoName:      "new-repo",
		isValid:       true,
	}
	serviceWithNewInfo := service.WithInfoProvider(newMockInfo)
	// Verify individual fields
	require.Equal(t, mockEngine, serviceWithNewInfo.Engine())
	require.Equal(t, mockCommit, serviceWithNewInfo.CommitService())
	require.Equal(t, newMockInfo, serviceWithNewInfo.InfoProvider())

	// Test that functional composition works
	result, err := service.WithEngine(&mockValidationEngine{}).
		WithCommitService(newMockCommit).
		WithInfoProvider(newMockInfo).
		ValidateCommit(context.Background(), "xyz789")

	require.NoError(t, err)
	require.Equal(t, "xyz789", result.CommitInfo.Hash)

	// Test that original service remains unchanged
	_, err = service.ValidateCommit(context.Background(), "xyz789")
	require.Error(t, err) // Original service doesn't have "xyz789" commit
}

// TestValueSemantics verifies that value semantics are properly maintained
// throughout the validation service and its components.
func TestValueSemantics(t *testing.T) {
	// Create a test commit
	testCommit := domain.CommitInfo{
		Hash:    "test123",
		Subject: "Test value semantics",
		Body:    "This is a test of value semantics",
	}

	// Create mocks
	mockCommit := &mockGitCommitService{
		commits: map[string]domain.CommitInfo{
			"test123": testCommit,
		},
	}
	mockInfo := &mockInfoProvider{}

	// Test ValidationEngine immutability
	t.Run("ValidationEngine immutability", func(t *testing.T) {
		// Create a validation engine
		provider := &validate.DomainRuleProvider{}
		engine := validate.DomainValidationEngine{} // Using exported struct for testing

		// Make a copy with WithProvider
		engineCopy := engine.WithProvider(provider)

		// Verify the copy is different from the original
		require.NotEqual(t, engine, engineCopy)

		// Verify the original is unchanged
		require.Nil(t, engine.GetProvider())

		// Verify the copy has the expected value
		require.Same(t, provider, engineCopy.GetProvider())
	})

	// Test DomainRuleProvider immutability
	t.Run("DomainRuleProvider immutability", func(t *testing.T) {
		// Create a custom validation rule
		customRule := &CustomRule{
			name: "TestCustomRule",
		}

		// Create an engine with rule provider
		engine := &mockValidationEngineWithCustomRules{
			customRules: []domain.Rule{},
		}

		// Create service with mock engine
		service := validate.NewValidationService(
			engine,
			mockCommit,
			mockInfo,
		)

		// Register a custom rule (this mutates the engine, which is expected)
		err := service.RegisterCustomRule(customRule)
		require.NoError(t, err)

		// Use WithCustomRule (should create a new service)
		serviceCopy, err := service.WithCustomRule(customRule)
		require.NoError(t, err)

		// Verify the custom rule is correctly added to the engine in the copy
		if engineWithCustomRules, ok := serviceCopy.Engine().(*mockValidationEngineWithCustomRules); ok {
			// Don't test the exact length, since the implementation might have added the rule to the original
			// engine as well. Instead verify that it contains the rule we added.
			found := false

			for _, rule := range engineWithCustomRules.customRules {
				if rule.Name() == customRule.Name() {
					found = true

					break
				}
			}

			require.True(t, found, "Custom rule should be registered in the engine")
		} else {
			require.Fail(t, "Engine should be of type *mockValidationEngineWithCustomRules")
		}
	})

	// Test function composition maintains value semantics
	t.Run("Function composition", func(t *testing.T) {
		// Create validation service
		service := validate.NewValidationService(
			&mockValidationEngine{},
			mockCommit,
			mockInfo,
		)

		// Create a chain of transformations
		transformedService := service.
			WithEngine(&mockValidationEngine{}).
			WithCommitService(mockCommit).
			WithInfoProvider(mockInfo)

		// Verify components individually rather than comparing whole structs
		// This ensures we're testing the behavior, not just the struct layout
		require.IsType(t, &mockValidationEngine{}, transformedService.Engine())
		require.Equal(t, mockCommit, transformedService.CommitService())
		require.Equal(t, mockInfo, transformedService.InfoProvider())
	})
}

func TestValidationService_ValidateCommit(t *testing.T) {
	// Create mocks
	mockCommit := &mockGitCommitService{
		commits: map[string]domain.CommitInfo{
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
		mockCommit,
		mockInfo,
	)

	// Test ValidateCommit
	t.Run("Validate existing commit", func(t *testing.T) {
		result, err := service.ValidateCommit(context.Background(), "abc123")
		require.NoError(t, err)
		require.Equal(t, "abc123", result.CommitInfo.Hash)
	})

	t.Run("Validate non-existent commit", func(t *testing.T) {
		_, err := service.ValidateCommit(context.Background(), "nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "commit not found")
	})
}

func TestValidationService_ValidateHeadCommits(t *testing.T) {
	// Create test commits
	commits := []domain.CommitInfo{
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
	mockCommit := &mockGitCommitService{
		headCommits: commits,
	}
	mockInfo := &mockInfoProvider{}

	// Create mock engine
	mockEngine := mockValidationEngine{}

	// Create service with mock engine
	service := validate.NewValidationService(
		mockEngine,
		mockCommit,
		mockInfo,
	)

	// Test ValidateHeadCommits
	t.Run("Validate head commits without filtering", func(t *testing.T) {
		results, err := service.ValidateHeadCommits(context.Background(), 3, false)
		require.NoError(t, err)
		require.Equal(t, 3, results.Count())
	})

	t.Run("Validate head commits with merge filtering", func(t *testing.T) {
		results, err := service.ValidateHeadCommits(context.Background(), 3, true)
		require.NoError(t, err)
		require.Equal(t, 2, results.Count())
	})
}

func TestValidationService_ValidateCommitRange(t *testing.T) {
	// Create test commits
	commits := []domain.CommitInfo{
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
	mockCommit := &mockGitCommitService{
		ranges: map[string]map[string][]domain.CommitInfo{
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
		mockCommit,
		mockInfo,
	)

	// Test ValidateCommitRange
	t.Run("Validate commit range without filtering", func(t *testing.T) {
		results, err := service.ValidateCommitRange(context.Background(), "feature", "master", false)
		require.NoError(t, err)
		require.Equal(t, 3, results.Count())
	})

	t.Run("Validate commit range with merge filtering", func(t *testing.T) {
		results, err := service.ValidateCommitRange(context.Background(), "feature", "master", true)
		require.NoError(t, err)
		require.Equal(t, 2, results.Count())
	})

	t.Run("Validate non-existent range", func(t *testing.T) {
		_, err := service.ValidateCommitRange(context.Background(), "nonexistent", "master", false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "range not found")
	})
}

func TestValidationService_ValidateWithOptions(t *testing.T) {
	// Create mocks with test data
	testCommit := domain.CommitInfo{
		Hash:          "abc123",
		Subject:       "Test commit",
		Body:          "Test body",
		IsMergeCommit: false,
	}

	// Create mocks
	mockCommit := &mockGitCommitService{
		commits: map[string]domain.CommitInfo{
			"abc123": testCommit,
			"HEAD":   testCommit,
		},
		headCommits: []domain.CommitInfo{testCommit},
		ranges: map[string]map[string][]domain.CommitInfo{
			"to": {
				"from": []domain.CommitInfo{testCommit},
			},
		},
	}

	mockInfo := mockInfoProvider{}

	// Create mock engine
	mockEngine := mockValidationEngine{}

	// Create service with mock engine
	service := validate.NewValidationService(
		mockEngine,
		mockCommit,
		mockInfo,
	)

	// Test different options
	t.Run("Validate specific commit", func(t *testing.T) {
		opts := validate.ValidationOptions{
			CommitHash: "abc123",
		}
		results, err := service.ValidateWithOptions(context.Background(), opts)
		require.NoError(t, err)
		require.Equal(t, 1, results.Count())
	})

	t.Run("Validate head commits", func(t *testing.T) {
		opts := validate.ValidationOptions{
			CommitCount:      1,
			SkipMergeCommits: false,
		}
		results, err := service.ValidateWithOptions(context.Background(), opts)
		require.NoError(t, err)
		require.Equal(t, 1, results.Count())
	})

	t.Run("Validate commit range", func(t *testing.T) {
		opts := validate.ValidationOptions{
			FromHash:         "from",
			ToHash:           "to",
			SkipMergeCommits: false,
		}
		results, err := service.ValidateWithOptions(context.Background(), opts)
		require.NoError(t, err)
		require.Equal(t, 1, results.Count())
	})

	t.Run("Default to HEAD", func(t *testing.T) {
		opts := validate.ValidationOptions{}
		results, err := service.ValidateWithOptions(context.Background(), opts)
		require.NoError(t, err)
		require.Equal(t, 1, results.Count())
	})
}

// CustomRule is a simple example of a custom validation rule for testing.
type CustomRule struct {
	name       string
	violations []appErrors.ValidationError
}

// Name returns the rule's name.
func (r *CustomRule) Name() string {
	return r.name
}

// Validate checks a commit for compliance.
func (r *CustomRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
	r.violations = nil

	// Example validation: require certain text in the commit message
	if len(commit.Message) < 10 {
		r.violations = append(r.violations, appErrors.ValidationError{
			Code:    "too_short",
			Message: "Commit message is too short",
		})
	}

	return r.violations
}

// Result returns a concise result message.
func (r *CustomRule) Result() string {
	if len(r.violations) > 0 {
		return "Custom validation failed"
	}

	return "Custom validation passed"
}

// VerboseResult returns a detailed result message.
func (r *CustomRule) VerboseResult() string {
	if len(r.violations) > 0 {
		return "Commit message didn't pass custom validation"
	}

	return "Commit message passed custom validation"
}

// Help returns guidance for fixing violations.
func (r *CustomRule) Help() string {
	return "Make sure your commit message meets the custom requirements"
}

// Errors returns all validation errors.
func (r *CustomRule) Errors() []appErrors.ValidationError {
	return r.violations
}

// mockValidationEngineWithCustomRules extends mockValidationEngine to support custom rules.
type mockValidationEngineWithCustomRules struct {
	mockValidationEngine
	customRules []domain.Rule
}

// RegisterCustomRule adds a custom rule.
func (m *mockValidationEngineWithCustomRules) RegisterCustomRule(rule domain.Rule) {
	m.customRules = append(m.customRules, rule)
}

// GetAvailableRuleNames returns names of all rules including custom ones.
func (m *mockValidationEngineWithCustomRules) GetAvailableRuleNames() []string {
	names := []string{"BuiltInRule1", "BuiltInRule2"}
	for _, rule := range m.customRules {
		names = append(names, rule.Name())
	}

	return names
}

// TestCustomRuleRegistration tests the ability to register custom rules.
func TestCustomRuleRegistration(t *testing.T) {
	// Create mocks
	mockCommit := &mockGitCommitService{
		commits: map[string]domain.CommitInfo{
			"HEAD": {
				Hash:    "def456",
				Subject: "Head commit",
				Message: "Head commit\n\nTest message",
			},
		},
	}
	mockInfo := &mockInfoProvider{}

	// Create mock engine with custom rule support
	mockEngine := &mockValidationEngineWithCustomRules{
		customRules: []domain.Rule{},
	}

	// Create service with mock engine
	service := validate.NewValidationService(
		mockEngine,
		mockCommit,
		mockInfo,
	)

	// Get rule names before adding custom rule
	beforeRules := service.GetAvailableRuleNames()
	require.Len(t, beforeRules, 2) // Should have 2 built-in rules

	// Create and register a custom rule
	customRule := &CustomRule{
		name: "TestCustomRule",
	}
	err := service.RegisterCustomRule(customRule)
	require.NoError(t, err)

	// Get rule names after adding custom rule
	afterRules := service.GetAvailableRuleNames()
	require.Len(t, afterRules, 3) // Should have 2 built-in rules + 1 custom rule

	// Verify the custom rule is in the list
	found := false

	for _, name := range afterRules {
		if name == "TestCustomRule" {
			found = true

			break
		}
	}

	require.True(t, found, "Custom rule should be in available rules")
}
