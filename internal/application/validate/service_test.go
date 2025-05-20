// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validate_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/log"
	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/core/validation"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

// mockCommitRepository implements the domain.CommitRepository interface for testing.
type mockCommitRepository struct {
	commits     map[string]domain.CommitInfo
	headCommits []domain.CommitInfo
	ranges      map[string]map[string][]domain.CommitInfo
}

func (m mockCommitRepository) GetCommit(_ context.Context, hash string) (domain.CommitInfo, error) {
	if commit, ok := m.commits[hash]; ok {
		return commit, nil
	}

	return domain.CommitInfo{}, appErrors.New("MockService", appErrors.ErrCommitNotFound, "commit not found")
}

func (m mockCommitRepository) GetCommits(_ context.Context, count int) ([]domain.CommitInfo, error) {
	if count <= 0 {
		return []domain.CommitInfo{}, nil
	}

	// Take the minimum of count or the available number of commits
	if count > len(m.headCommits) {
		count = len(m.headCommits)
	}

	return m.headCommits[:count], nil
}

func (m mockCommitRepository) GetHeadCommits(_ context.Context, count int) ([]domain.CommitInfo, error) {
	if count <= 0 {
		return []domain.CommitInfo{}, nil
	}

	// Take the minimum of count or the available number of commits
	if count > len(m.headCommits) {
		count = len(m.headCommits)
	}

	return m.headCommits[:count], nil
}

func (m mockCommitRepository) GetCommitRange(_ context.Context, fromHash, toHash string) ([]domain.CommitInfo, error) {
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

func (m mockInfoProvider) IsValid(_ context.Context) (bool, error) {
	return m.isValid, nil
}

// mockValidationEngine implements validation.Engine interface.
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
		results = results.WithResult(result)
	}

	return results
}

func (m mockValidationEngine) GetRegistry() *domain.RuleRegistry {
	return domain.NewRuleRegistry()
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
func (r *CustomRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := log.Logger(ctx)
	logger.Trace().Str("rule", r.name).Msg("Entering CustomRule.Validate")

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
func (r *CustomRule) Result(_ []appErrors.ValidationError) string {
	if len(r.violations) > 0 {
		return "Custom validation failed"
	}

	return "Custom validation passed"
}

// VerboseResult returns a detailed result message.
func (r *CustomRule) VerboseResult(_ []appErrors.ValidationError) string {
	if len(r.violations) > 0 {
		return "Commit message didn't pass custom validation"
	}

	return "Commit message passed custom validation"
}

// Help returns guidance for fixing violations.
func (r *CustomRule) Help(_ []appErrors.ValidationError) string {
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

// WithCustomRule returns a new engine with the custom rule added.
func (m *mockValidationEngineWithCustomRules) WithCustomRule(rule domain.Rule) validation.Engine {
	// Create a new rules array directly with capacity for the new rule
	newCustomRules := make([]domain.Rule, 0, len(m.customRules)+1)

	// Add existing rules
	newCustomRules = append(newCustomRules, m.customRules...)

	// Add new rule
	newCustomRules = append(newCustomRules, rule)

	// Return new instance
	return &mockValidationEngineWithCustomRules{
		customRules: newCustomRules,
	}
}

// GetAvailableRuleNames returns names of all rules including custom ones.
func (m *mockValidationEngineWithCustomRules) GetAvailableRuleNames(ctx context.Context) []string {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering mockValidationEngineWithCustomRules.GetAvailableRuleNames")

	names := []string{"BuiltInRule1", "BuiltInRule2"}
	for _, rule := range m.customRules {
		names = append(names, rule.Name())
	}

	return names
}

// TestValidationService_ValidateCommit tests the ValidateCommit method using table-driven tests.
func TestValidationService_ValidateCommit(t *testing.T) {
	tests := []struct {
		name         string
		commitHash   string
		setupService func() validate.ValidationService
		wantErr      bool
		errContains  string
		assertResult func(t *testing.T, result domain.CommitResult)
	}{
		{
			name:       "Valid existing commit",
			commitHash: "abc123",
			setupService: func() validate.ValidationService {
				mockCommit := &mockCommitRepository{
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

				return validate.NewValidationService(mockEngine, mockCommit, mockInfo)
			},
			wantErr: false,
			assertResult: func(t *testing.T, result domain.CommitResult) { //nolint:thelper
				require.Equal(t, "abc123", result.CommitInfo.Hash)
				require.True(t, result.Passed)
			},
		},
		{
			name:       "Non-existent commit",
			commitHash: "nonexistent",
			setupService: func() validate.ValidationService {
				mockCommit := &mockCommitRepository{
					commits: map[string]domain.CommitInfo{},
				}
				mockInfo := &mockInfoProvider{}
				mockEngine := &mockValidationEngine{}

				return validate.NewValidationService(mockEngine, mockCommit, mockInfo)
			},
			wantErr:     true,
			errContains: "commit not found",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			service := testCase.setupService()
			result, err := service.ValidateCommit(testcontext.CreateTestContext(), testCase.commitHash)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errContains != "" {
					require.Contains(t, err.Error(), testCase.errContains)
				}

				return
			}

			require.NoError(t, err)

			if testCase.assertResult != nil {
				testCase.assertResult(t, result)
			}
		})
	}
}

// TestValidationService_ValidateLastNCommits tests the ValidateLastNCommits method using table-driven tests.
func TestValidationService_ValidateLastNCommits(t *testing.T) {
	// Define test commits
	testCommits := []domain.CommitInfo{
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

	tests := []struct {
		name             string
		count            int
		skipMergeCommits bool
		setupService     func() validate.ValidationService
		wantCount        int
	}{
		{
			name:             "All commits without filtering",
			count:            3,
			skipMergeCommits: false,
			setupService: func() validate.ValidationService {
				mockCommit := &mockCommitRepository{
					headCommits: testCommits,
				}
				mockInfo := &mockInfoProvider{}
				mockEngine := &mockValidationEngine{}

				return validate.NewValidationService(mockEngine, mockCommit, mockInfo)
			},
			wantCount: 3,
		},
		{
			name:             "All commits with merge filtering",
			count:            3,
			skipMergeCommits: true,
			setupService: func() validate.ValidationService {
				mockCommit := &mockCommitRepository{
					headCommits: testCommits,
				}
				mockInfo := &mockInfoProvider{}
				mockEngine := &mockValidationEngine{}

				return validate.NewValidationService(mockEngine, mockCommit, mockInfo)
			},
			wantCount: 2, // Excluding merge commit
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			service := testCase.setupService()
			results, err := service.ValidateLastNCommits(testcontext.CreateTestContext(), testCase.count, testCase.skipMergeCommits)

			require.NoError(t, err)
			require.Equal(t, testCase.wantCount, results.Count())
		})
	}
}

// TestValidationService_ValidateCommitRange tests the ValidateCommitRange method using table-driven tests.
func TestValidationService_ValidateCommitRange(t *testing.T) {
	// Define test commits
	testCommits := []domain.CommitInfo{
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

	tests := []struct {
		name             string
		fromHash         string
		toHash           string
		skipMergeCommits bool
		setupService     func() validate.ValidationService
		wantCount        int
		wantErr          bool
		errContains      string
	}{
		{
			name:             "Valid range without filtering",
			fromHash:         "feature",
			toHash:           "master",
			skipMergeCommits: false,
			setupService: func() validate.ValidationService {
				mockCommit := &mockCommitRepository{
					ranges: map[string]map[string][]domain.CommitInfo{
						"master": {
							"feature": testCommits,
						},
					},
				}
				mockInfo := &mockInfoProvider{}
				mockEngine := &mockValidationEngine{}

				return validate.NewValidationService(mockEngine, mockCommit, mockInfo)
			},
			wantCount: 3,
		},
		{
			name:             "Valid range with merge filtering",
			fromHash:         "feature",
			toHash:           "master",
			skipMergeCommits: true,
			setupService: func() validate.ValidationService {
				mockCommit := &mockCommitRepository{
					ranges: map[string]map[string][]domain.CommitInfo{
						"master": {
							"feature": testCommits,
						},
					},
				}
				mockInfo := &mockInfoProvider{}
				mockEngine := &mockValidationEngine{}

				return validate.NewValidationService(mockEngine, mockCommit, mockInfo)
			},
			wantCount: 2, // Excluding merge commit
		},
		{
			name:             "Non-existent range",
			fromHash:         "nonexistent",
			toHash:           "master",
			skipMergeCommits: false,
			setupService: func() validate.ValidationService {
				mockCommit := &mockCommitRepository{
					ranges: map[string]map[string][]domain.CommitInfo{},
				}
				mockInfo := &mockInfoProvider{}
				mockEngine := &mockValidationEngine{}

				return validate.NewValidationService(mockEngine, mockCommit, mockInfo)
			},
			wantErr:     true,
			errContains: "range not found",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			service := testCase.setupService()
			results, err := service.ValidateCommitRange(testcontext.CreateTestContext(), testCase.fromHash, testCase.toHash, testCase.skipMergeCommits)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errContains != "" {
					require.Contains(t, err.Error(), testCase.errContains)
				}

				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.wantCount, results.Count())
		})
	}
}

// TestValidationService_ValidateWithOptions tests the ValidateWithOptions method using table-driven tests.
func TestValidationService_ValidateWithOptions(t *testing.T) {
	// Define test data
	testCommit := domain.CommitInfo{
		Hash:    "abc123",
		Subject: "Test commit",
		Body:    "Test body",
	}

	tests := []struct {
		name         string
		options      validate.ValidationOptions
		setupService func() validate.ValidationService
		wantCount    int
		wantErr      bool
	}{
		{
			name: "Validate specific commit",
			options: validate.ValidationOptions{
				CommitHash: "abc123",
			},
			setupService: func() validate.ValidationService {
				mockCommit := &mockCommitRepository{
					commits: map[string]domain.CommitInfo{
						"abc123": testCommit,
					},
				}
				mockInfo := &mockInfoProvider{}
				mockEngine := &mockValidationEngine{}

				return validate.NewValidationService(mockEngine, mockCommit, mockInfo)
			},
			wantCount: 1,
		},
		{
			name: "Validate head commits",
			options: validate.ValidationOptions{
				CommitCount:      1,
				SkipMergeCommits: false,
			},
			setupService: func() validate.ValidationService {
				mockCommit := &mockCommitRepository{
					headCommits: []domain.CommitInfo{testCommit},
				}
				mockInfo := &mockInfoProvider{}
				mockEngine := &mockValidationEngine{}

				return validate.NewValidationService(mockEngine, mockCommit, mockInfo)
			},
			wantCount: 1,
		},
		{
			name: "Validate commit range",
			options: validate.ValidationOptions{
				FromHash:         "from",
				ToHash:           "to",
				SkipMergeCommits: false,
			},
			setupService: func() validate.ValidationService {
				mockCommit := &mockCommitRepository{
					ranges: map[string]map[string][]domain.CommitInfo{
						"to": {
							"from": []domain.CommitInfo{testCommit},
						},
					},
				}
				mockInfo := &mockInfoProvider{}
				mockEngine := &mockValidationEngine{}

				return validate.NewValidationService(mockEngine, mockCommit, mockInfo)
			},
			wantCount: 1,
		},
		{
			name:    "Default to HEAD",
			options: validate.ValidationOptions{},
			setupService: func() validate.ValidationService {
				mockCommit := &mockCommitRepository{
					headCommits: []domain.CommitInfo{testCommit},
				}
				mockInfo := &mockInfoProvider{}
				mockEngine := &mockValidationEngine{}

				return validate.NewValidationService(mockEngine, mockCommit, mockInfo)
			},
			wantCount: 1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			service := testCase.setupService()
			results, err := service.ValidateWithOptions(testcontext.CreateTestContext(), testCase.options)

			if testCase.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.wantCount, results.Count())
		})
	}
}

// TestValidationService_Functional tests the functional "With" methods.
func TestValidationService_Functional(t *testing.T) {
	// Create base mocks
	baseMockCommit := &mockCommitRepository{
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

	tests := []struct {
		name     string
		testFunc func(t *testing.T, service validate.ValidationService)
	}{
		{
			name: "WithEngine creates new service",
			testFunc: func(t *testing.T, service validate.ValidationService) { //nolint:thelper
				newEngine := &mockValidationEngine{}
				newService := service.WithEngine(newEngine)

				require.Equal(t, newEngine, newService.Engine())
				require.Equal(t, service.CommitService(), newService.CommitService())
				require.Equal(t, service.InfoProvider(), newService.InfoProvider())
			},
		},
		{
			name: "WithCommitService creates new service",
			testFunc: func(t *testing.T, service validate.ValidationService) { //nolint:thelper
				newCommitService := &mockCommitRepository{
					commits: map[string]domain.CommitInfo{
						"xyz789": {
							Hash:    "xyz789",
							Subject: "New commit",
						},
					},
				}
				newService := service.WithCommitService(newCommitService)

				require.Equal(t, service.Engine(), newService.Engine())
				require.Equal(t, newCommitService, newService.CommitService())
				require.Equal(t, service.InfoProvider(), newService.InfoProvider())
			},
		},
		{
			name: "WithInfoProvider creates new service",
			testFunc: func(t *testing.T, service validate.ValidationService) { //nolint:thelper
				newInfoProvider := &mockInfoProvider{
					currentBranch: "feature",
					repoName:      "new-repo",
					isValid:       true,
				}
				newService := service.WithInfoProvider(newInfoProvider)

				require.Equal(t, service.Engine(), newService.Engine())
				require.Equal(t, service.CommitService(), newService.CommitService())
				require.Equal(t, newInfoProvider, newService.InfoProvider())
			},
		},
		{
			name: "Functional composition maintains immutability",
			testFunc: func(t *testing.T, service validate.ValidationService) { //nolint:thelper
				newCommitService := &mockCommitRepository{
					commits: map[string]domain.CommitInfo{
						"xyz789": {
							Hash:    "xyz789",
							Subject: "New commit",
						},
					},
				}

				// Compose new service
				composedService := service.
					WithEngine(&mockValidationEngine{}).
					WithCommitService(newCommitService)

				// Test composed service
				result, err := composedService.ValidateCommit(testcontext.CreateTestContext(), "xyz789")
				require.NoError(t, err)
				require.Equal(t, "xyz789", result.CommitInfo.Hash)

				// Test original service remains unchanged
				_, err = service.ValidateCommit(testcontext.CreateTestContext(), "xyz789")
				require.Error(t, err) // Original doesn't have xyz789
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			service := validate.NewValidationService(mockEngine, baseMockCommit, mockInfo)
			testCase.testFunc(t, service)
		})
	}
}

// TestValueSemantics verifies that value semantics are properly maintained.
func TestValueSemantics(t *testing.T) {
	// Need to update this test to properly handle context-based validation
	// The test expects ValidationResults but ValidateCommit returns CommitResult
	// This test would need to be restructured based on actual API
	t.Run("WithCustomRule maintains immutability", func(t *testing.T) {
		// Create a custom validation rule
		customRule := &CustomRule{
			name: "TestCustomRule",
		}

		// Create an engine that supports custom rules
		engine := &mockValidationEngineWithCustomRules{
			customRules: []domain.Rule{},
		}

		// Create service
		service := validate.NewValidationService(
			engine,
			&mockCommitRepository{},
			&mockInfoProvider{},
		)

		// Add a custom rule
		ctx := testcontext.CreateTestContext() // Create a test context
		serviceCopy, err := service.WithCustomRule(ctx, customRule)
		require.NoError(t, err)

		// Verify the custom rule is correctly added to the engine in the copy
		if engineWithCustomRules, ok := serviceCopy.Engine().(*mockValidationEngineWithCustomRules); ok {
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
}
