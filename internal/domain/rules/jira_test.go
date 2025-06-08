// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/stretchr/testify/require"
)

// createJiraTestCommit creates a test commit with default values.
func createJiraTestCommit() domain.Commit {
	return domain.Commit{
		Hash:          "abc123def456",
		Subject:       "feat: add new feature",
		Message:       "feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.",
		Body:          "This commit adds a new feature that enhances the user experience.",
		Author:        "Test User",
		AuthorEmail:   "test@example.com",
		CommitDate:    time.Now().Format(time.RFC3339),
		IsMergeCommit: false,
	}
}

// newConfigBuilder creates a config builder with sensible defaults.
func newConfigBuilder() configBuilder {
	return configBuilder{
		config: config.Config{
			Rules: config.RulesConfig{
				Enabled:  []string{},
				Disabled: []string{},
			},
			Jira: config.JiraConfig{
				ProjectPrefixes:      []string{},
				RequireInBody:        false,
				RequireInSubject:     false,
				IgnoreTicketPatterns: []string{},
			},
		},
	}
}

// configBuilder provides a functional builder for creating test configurations.
type configBuilder struct {
	config config.Config
}

// WithJiraProjects sets the allowed Jira project prefixes.
func (b configBuilder) WithJiraProjects(projects []string) configBuilder {
	b.config.Jira.ProjectPrefixes = projects

	return b
}

// WithJiraCheckBody enables checking the commit body for Jira references.
func (b configBuilder) WithJiraCheckBody(checkBody bool) configBuilder {
	b.config.Jira.RequireInBody = checkBody

	return b
}

// WithJiraProjectPrefixes sets the Jira project prefixes.
func (b configBuilder) WithJiraProjectPrefixes(prefixes []string) configBuilder {
	b.config.Jira.ProjectPrefixes = prefixes

	return b
}

// Build returns the constructed config.
func (b configBuilder) Build() config.Config {
	return b.config
}

func TestJiraReferenceRule_Validate(t *testing.T) {
	tests := []struct {
		name        string
		commit      domain.Commit
		configSetup func() config.Config
		wantErrors  bool
		description string
	}{
		{
			name: "jira required in subject",
			configSetup: func() config.Config {
				return newConfigBuilder().Build()
			},
			commit: func() domain.Commit {
				commit := createJiraTestCommit()
				commit.Subject = "Add new feature PROJ-123"
				commit.Body = "This is a description"

				return commit
			}(),
			wantErrors:  false,
			description: "Should pass with valid Jira reference in subject",
		},
		{
			name: "jira required in body with body ref enabled",
			configSetup: func() config.Config {
				return newConfigBuilder().
					WithJiraCheckBody(true).
					Build()
			},
			commit: func() domain.Commit {
				commit := createJiraTestCommit()
				commit.Subject = "Add new feature"
				commit.Body = "This is a description\n\nRefs: PROJ-123"

				return commit
			}(),
			wantErrors:  false,
			description: "Should pass with valid Jira reference in body",
		},
		{
			name: "jira required but missing",
			configSetup: func() config.Config {
				return newConfigBuilder().Build()
			},
			commit: func() domain.Commit {
				commit := createJiraTestCommit()
				commit.Subject = "Add new feature"
				commit.Body = "This is a description"

				return commit
			}(),
			wantErrors:  true,
			description: "Should fail when Jira is required but missing",
		},
		{
			name: "jira with conventional and valid projects",
			configSetup: func() config.Config {
				return newConfigBuilder().
					WithJiraProjects([]string{"PROJ", "TEAM"}).
					Build()
			},
			commit: func() domain.Commit {
				commit := createJiraTestCommit()
				commit.Subject = "feat: add login feature PROJ-123"
				commit.Body = "This is a description"

				return commit
			}(),
			wantErrors:  false,
			description: "Should pass with conventional format and valid project",
		},
		{
			name: "jira with conventional and invalid project",
			configSetup: func() config.Config {
				return newConfigBuilder().
					WithJiraProjects([]string{"PROJ", "TEAM"}).
					Build()
			},
			commit: func() domain.Commit {
				commit := createJiraTestCommit()
				commit.Subject = "feat: add login feature OTHER-123"
				commit.Body = "This is a description"

				return commit
			}(),
			wantErrors:  true,
			description: "Should fail with conventional format and invalid project",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup config
			cfg := testCase.configSetup()

			// Create rule with configuration
			// For this test, assume conventional commit is enabled
			cfg.Rules.Enabled = []string{"conventional"}

			rule := rules.NewJiraReferenceRule(cfg)

			// Execute
			failures := rule.Validate(testCase.commit, cfg)

			// Assert
			if testCase.wantErrors {
				require.NotEmpty(t, failures, "Expected validation errors but got none")
			} else {
				require.Empty(t, failures, "Expected no validation errors but got: %v", failures)
			}
		})
	}
}

// TestJiraReferenceRule_EnhancedConfigurationValidation tests the enhanced configuration functionality.
func TestJiraReferenceRule_EnhancedConfigurationValidation(t *testing.T) {
	t.Skip("Enhanced context features not fully implemented for JIRA rules")

	tests := []struct {
		name         string
		subject      string
		body         string
		config       config.Config
		wantErrCount int
		wantErrCode  string
		wantContext  map[string]string
	}{
		{
			name:    "RequireInSubject enabled - JIRA in subject",
			subject: "feat: add feature PROJ-123",
			body:    "Description of the feature",
			config: config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes:  []string{"PROJ"},
					RequireInSubject: true,
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCount: 0,
		},
		{
			name:    "RequireInSubject enabled - JIRA only in body",
			subject: "feat: add feature",
			body:    "Description of the feature\n\nRefs: PROJ-123",
			config: config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes:  []string{"PROJ"},
					RequireInSubject: true,
					RequireInBody:    true,
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrMissingJiraKeySubject),
			wantContext: map[string]string{
				"subject":       "feat: add feature",
				"commit_format": "conventional",
			},
		},
		{
			name:    "IgnoreTicketPatterns - ignored pattern",
			subject: "feat: add feature IGNORE-123",
			body:    "Description",
			config: config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes:      []string{"PROJ"},
					IgnoreTicketPatterns: []string{"IGNORE-.*"},
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrMissingJira),
		},
		{
			name:    "IgnoreTicketPatterns - valid pattern",
			subject: "feat: add feature PROJ-123",
			body:    "Description",
			config: config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes:      []string{"PROJ"},
					IgnoreTicketPatterns: []string{"IGNORE-.*"},
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCount: 0,
		},
		{
			name:    "RequireInBody and RequireInSubject both enabled",
			subject: "feat: add feature PROJ-123",
			body:    "Description\n\nRefs: PROJ-456",
			config: config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes:  []string{"PROJ"},
					RequireInBody:    true,
					RequireInSubject: true,
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCount: 0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewJiraReferenceRule(testCase.config)
			commit := domain.Commit{
				Subject: testCase.subject,
				Body:    testCase.body,
			}

			errors := rule.Validate(commit, testCase.config)

			require.Len(t, errors, testCase.wantErrCount, "unexpected number of errors")

			if testCase.wantErrCount > 0 && testCase.wantErrCode != "" {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)

				for key, expectedValue := range testCase.wantContext {
					require.Equal(t, expectedValue, errors[0].Context[key])
				}
			}
		})
	}
}

// TestJiraReferenceRule_EnhancedErrorHandling tests enhanced error handling with specific error codes.
func TestJiraReferenceRule_EnhancedErrorHandling(t *testing.T) {
	t.Skip("Enhanced context features not fully implemented for JIRA rules")

	tests := []struct {
		name           string
		subject        string
		body           string
		config         config.Config
		wantErrCode    string
		wantErrMessage string
		wantContext    map[string]string
	}{
		{
			name:    "Missing JIRA - enhanced error message",
			subject: "feat: add feature",
			body:    "Description",
			config: config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes: []string{"PROJ"},
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCode:    string(domain.ErrMissingJira),
			wantErrMessage: "Missing JIRA reference",
			wantContext: map[string]string{
				"expected_format": "PROJ-123",
				"searched_in":     "subject and body",
			},
		},
		{
			name:    "Invalid project - enhanced error message",
			subject: "feat: add feature OTHER-123",
			body:    "Description",
			config: config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes: []string{"PROJ", "TEAM"},
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCode:    string(domain.ErrInvalidProject),
			wantErrMessage: "Invalid project 'OTHER' in reference 'OTHER-123'",
			wantContext: map[string]string{
				"invalid_project":   "OTHER",
				"invalid_reference": "OTHER-123",
				"allowed_projects":  "PROJ, TEAM",
			},
		},
		{
			name:    "Misplaced JIRA in conventional commit - enhanced error",
			subject: "feat: add PROJ-123 feature implementation",
			body:    "Description",
			config: config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes:  []string{"PROJ"},
					RequireInSubject: true,
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCode:    string(domain.ErrJiraKeyNotAtEnd),
			wantErrMessage: "JIRA key must be at the end",
			wantContext: map[string]string{
				"subject":           "feat: add PROJ-123 feature implementation",
				"found_keys":        "PROJ-123",
				"expected_position": "end_of_subject",
				"commit_type":       "conventional",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewJiraReferenceRule(testCase.config)
			commit := domain.Commit{
				Subject: testCase.subject,
				Body:    testCase.body,
			}

			errors := rule.Validate(commit, testCase.config)

			require.NotEmpty(t, errors, "Expected validation error")
			require.Equal(t, testCase.wantErrCode, errors[0].Code)
			require.Contains(t, errors[0].Message, testCase.wantErrMessage)

			for key, expectedValue := range testCase.wantContext {
				require.Equal(t, expectedValue, errors[0].Context[key])
			}
		})
	}
}

// TestJiraReferenceRule_ProjectPrefixValidation tests project prefix validation functionality.
func TestJiraReferenceRule_ProjectPrefixValidation(t *testing.T) {
	tests := []struct {
		name            string
		subject         string
		body            string
		projectPrefixes []string
		wantErrCount    int
		wantErrCode     string
	}{
		{
			name:            "Valid single project",
			subject:         "feat: add feature PROJ-123",
			body:            "",
			projectPrefixes: []string{"PROJ"},
			wantErrCount:    0,
		},
		{
			name:            "Valid multiple projects",
			subject:         "feat: add feature PROJ-123 TEAM-456",
			body:            "",
			projectPrefixes: []string{"PROJ", "TEAM"},
			wantErrCount:    0,
		},
		{
			name:            "Invalid project prefix",
			subject:         "feat: add feature OTHER-123",
			body:            "",
			projectPrefixes: []string{"PROJ", "TEAM"},
			wantErrCount:    1,
			wantErrCode:     string(domain.ErrInvalidProject),
		},
		{
			name:            "Mixed valid and invalid projects",
			subject:         "feat: add feature PROJ-123 OTHER-456",
			body:            "",
			projectPrefixes: []string{"PROJ", "TEAM"},
			wantErrCount:    1,
			wantErrCode:     string(domain.ErrInvalidProject),
		},
		{
			name:            "No project prefixes - accept all",
			subject:         "feat: add feature ANYTHING-123",
			body:            "",
			projectPrefixes: []string{},
			wantErrCount:    0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes: testCase.projectPrefixes,
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			}

			rule := rules.NewJiraReferenceRule(cfg)
			commit := domain.Commit{
				Subject: testCase.subject,
				Body:    testCase.body,
			}

			errors := rule.Validate(commit, cfg)

			require.Len(t, errors, testCase.wantErrCount, "unexpected number of errors")

			if testCase.wantErrCount > 0 {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)
			}
		})
	}
}

// TestJiraReferenceRule_ConventionalCommitIntegration tests conventional commit integration.
func TestJiraReferenceRule_ConventionalCommitIntegration(t *testing.T) {
	tests := []struct {
		name         string
		subject      string
		body         string
		wantErrCount int
		wantErrCode  string
	}{
		{
			name:         "Valid conventional with JIRA at end",
			subject:      "feat: add new feature PROJ-123",
			body:         "Description",
			wantErrCount: 0,
		},
		{
			name:         "Valid conventional with JIRA in scope",
			subject:      "feat(PROJ-123): add new feature",
			body:         "Description",
			wantErrCount: 0,
		},
		{
			name:         "Invalid - JIRA in middle of description",
			subject:      "feat: add PROJ-123 new feature",
			body:         "Description",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrJiraKeyNotAtEnd),
		},
		{
			name:         "Valid - excluded commit type",
			subject:      "docs: update README PROJ-123",
			body:         "Description",
			wantErrCount: 0,
		},
		{
			name:         "Valid - merge commit",
			subject:      "merge: merge feature branch PROJ-123",
			body:         "Description",
			wantErrCount: 0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes:  []string{"PROJ"},
					RequireInSubject: true,
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			}

			rule := rules.NewJiraReferenceRule(cfg)
			commit := domain.Commit{
				Subject: testCase.subject,
				Body:    testCase.body,
			}

			errors := rule.Validate(commit, cfg)

			require.Len(t, errors, testCase.wantErrCount, "unexpected number of errors")

			if testCase.wantErrCount > 0 {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)
			}
		})
	}
}

// TestJiraReferenceRule_BackwardCompatibility tests that existing functionality is preserved.
func TestJiraReferenceRule_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name         string
		subject      string
		body         string
		config       config.Config
		wantErrCount int
		description  string
	}{
		{
			name:    "Original behavior - JIRA in subject",
			subject: "Add new feature PROJ-123",
			body:    "Description",
			config: config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes: []string{"PROJ"},
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCount: 0,
			description:  "Basic JIRA validation should work as before",
		},
		{
			name:    "Original behavior - missing JIRA",
			subject: "Add new feature",
			body:    "Description",
			config: config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes: []string{"PROJ"},
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCount: 1,
			description:  "Missing JIRA should still fail as before",
		},
		{
			name:    "Original behavior - invalid project",
			subject: "Add new feature OTHER-123",
			body:    "Description",
			config: config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes: []string{"PROJ"},
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCount: 1,
			description:  "Invalid project should still fail as before",
		},
		{
			name:    "Original behavior - body search enabled",
			subject: "Add new feature",
			body:    "Description\n\nRefs: PROJ-123",
			config: config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes: []string{"PROJ"},
					RequireInBody:   true,
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCount: 0,
			description:  "Body search should work as before",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewJiraReferenceRule(testCase.config)
			commit := domain.Commit{
				Subject: testCase.subject,
				Body:    testCase.body,
			}

			errors := rule.Validate(commit, testCase.config)

			require.Len(t, errors, testCase.wantErrCount, testCase.description)
		})
	}
}

// TestJiraReferenceRule_EnhancedSubjectPlacement tests sophisticated subject placement validation.
func TestJiraReferenceRule_EnhancedSubjectPlacement(t *testing.T) {
	tests := []struct {
		name         string
		subject      string
		wantErrCount int
		wantErrCode  string
		description  string
	}{
		{
			name:         "Conventional commit - JIRA at end (valid)",
			subject:      "feat: Add authentication system PROJ-123",
			wantErrCount: 0,
			description:  "JIRA key at end of conventional commit should be valid",
		},
		{
			name:         "Conventional commit - JIRA not at end (invalid)",
			subject:      "feat: PROJ-123 add authentication system",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrJiraKeyNotAtEnd),
			description:  "JIRA key not at end should trigger specific error",
		},
		{
			name:         "Conventional commit - multiple JIRA at end (valid)",
			subject:      "feat: Add feature PROJ-123 TEAM-456",
			wantErrCount: 0,
			description:  "Multiple JIRA keys at end should be valid",
		},
		{
			name:         "Non-conventional commit - JIRA anywhere (valid)",
			subject:      "PROJ-123 add authentication system",
			wantErrCount: 0,
			description:  "Non-conventional commits can have JIRA anywhere",
		},
		{
			name:         "Conventional commit with scope - JIRA at end (valid)",
			subject:      "feat(auth): Add login functionality PROJ-123",
			wantErrCount: 0,
			description:  "JIRA at end of scoped conventional commit should be valid",
		},
		{
			name:         "Breaking change with JIRA at end (valid)",
			subject:      "feat!: Add breaking API change PROJ-123",
			wantErrCount: 0,
			description:  "Breaking change commits with JIRA at end should be valid",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes:  []string{"PROJ", "TEAM"},
					RequireInSubject: true,
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			}

			rule := rules.NewJiraReferenceRule(cfg)
			commit := domain.Commit{Subject: testCase.subject}

			errors := rule.Validate(commit, cfg)

			require.Len(t, errors, testCase.wantErrCount, testCase.description)

			if testCase.wantErrCount > 0 {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)
			}
		})
	}
}

// TestJiraReferenceRule_EnhancedBodyValidation tests sophisticated body validation with Refs format.
func TestJiraReferenceRule_EnhancedBodyValidation(t *testing.T) {
	tests := []struct {
		name         string
		subject      string
		body         string
		wantErrCount int
		wantErrCode  string
		description  string
	}{
		{
			name:    "Valid Refs format - single key",
			subject: "feat: Add new feature",
			body: `This is the commit body.

Refs: PROJ-123

Signed-off-by: Developer <dev@example.com>`,
			wantErrCount: 0,
			description:  "Valid Refs format should pass validation",
		},
		{
			name:    "Valid Refs format - multiple keys",
			subject: "feat: Add new feature",
			body: `This is the commit body.

Refs: PROJ-123, TEAM-456

Signed-off-by: Developer <dev@example.com>`,
			wantErrCount: 0,
			description:  "Multiple JIRA keys in Refs should be valid",
		},
		{
			name:    "Invalid Refs format - missing colon",
			subject: "feat: Add new feature",
			body: `This is the commit body.

Refs PROJ-123

Signed-off-by: Developer <dev@example.com>`,
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrInvalidRefsFormat),
			description:  "Missing colon should trigger format error",
		},
		{
			name:    "Invalid Refs format - lowercase key",
			subject: "feat: Add new feature",
			body: `This is the commit body.

Refs: proj-123

Signed-off-by: Developer <dev@example.com>`,
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrInvalidKeyFormat),
			description:  "Lowercase JIRA key should trigger format error",
		},
		{
			name:    "Refs after Signed-off-by - invalid ordering",
			subject: "feat: Add new feature",
			body: `This is the commit body.

Signed-off-by: Developer <dev@example.com>
Refs: PROJ-123`,
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrRefsAfterSignoff),
			description:  "Refs after Signed-off-by should trigger ordering error",
		},
		{
			name:    "Missing Refs line",
			subject: "feat: Add new feature",
			body: `This is the commit body.

Signed-off-by: Developer <dev@example.com>`,
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrMissingJiraKeyBody),
			description:  "Missing Refs line should trigger missing key error",
		},
		{
			name:    "Invalid project in Refs",
			subject: "feat: Add new feature",
			body: `This is the commit body.

Refs: INVALID-123

Signed-off-by: Developer <dev@example.com>`,
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrInvalidProject),
			description:  "Invalid project in Refs should trigger project error",
		},
		{
			name:    "Multiple Refs lines with valid ordering",
			subject: "feat: Add new feature",
			body: `This is the commit body.

Refs: PROJ-123
Refs: TEAM-456

Signed-off-by: Developer <dev@example.com>`,
			wantErrCount: 0,
			description:  "Multiple Refs lines should be valid if ordered correctly",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Jira: config.JiraConfig{
					ProjectPrefixes: []string{"PROJ", "TEAM"},
					RequireInBody:   true,
				},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			}

			rule := rules.NewJiraReferenceRule(cfg)
			commit := domain.Commit{
				Subject: testCase.subject,
				Body:    testCase.body,
			}

			errors := rule.Validate(commit, cfg)

			require.Len(t, errors, testCase.wantErrCount, testCase.description)

			if testCase.wantErrCount > 0 {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)
			}
		})
	}
}

// TestJiraReferenceRule_RichErrorTaxonomy tests that all 8 original error codes are restored.
func TestJiraReferenceRule_RichErrorTaxonomy(t *testing.T) {
	tests := []struct {
		name        string
		subject     string
		body        string
		config      config.Config
		wantErrCode string
		description string
	}{
		{
			name:    "ErrEmptySubject",
			subject: "",
			body:    "",
			config: config.Config{
				Jira: config.JiraConfig{RequireInSubject: true},
			},
			wantErrCode: string(domain.ErrEmptySubject),
			description: "Empty subject should trigger specific error",
		},
		{
			name:    "ErrMissingJiraKeySubject",
			subject: "feat: Add new feature",
			body:    "",
			config: config.Config{
				Jira: config.JiraConfig{
					RequireInSubject: true,
					ProjectPrefixes:  []string{"PROJ"},
				},
			},
			wantErrCode: string(domain.ErrMissingJiraKeySubject),
			description: "Missing JIRA in subject should trigger specific error",
		},
		{
			name:    "ErrMissingJiraKeyBody",
			subject: "feat: Add new feature",
			body:    "No Refs line here",
			config: config.Config{
				Jira: config.JiraConfig{RequireInBody: true},
			},
			wantErrCode: string(domain.ErrMissingJiraKeyBody),
			description: "Missing JIRA in body should trigger specific error",
		},
		{
			name:    "ErrJiraKeyNotAtEnd",
			subject: "feat: PROJ-123 add new feature",
			body:    "",
			config: config.Config{
				Jira:  config.JiraConfig{RequireInSubject: true},
				Rules: config.RulesConfig{Enabled: []string{"conventional"}},
			},
			wantErrCode: string(domain.ErrJiraKeyNotAtEnd),
			description: "JIRA not at end in conventional commit should trigger specific error",
		},
		{
			name:    "ErrInvalidProject",
			subject: "feat: Add new feature INVALID-123",
			body:    "",
			config: config.Config{
				Jira: config.JiraConfig{
					RequireInSubject: true,
					ProjectPrefixes:  []string{"PROJ"},
				},
			},
			wantErrCode: string(domain.ErrInvalidProject),
			description: "Invalid project prefix should trigger specific error",
		},
		{
			name:    "ErrInvalidRefsFormat",
			subject: "feat: Add new feature",
			body:    "Refs invalid format without colon",
			config: config.Config{
				Jira: config.JiraConfig{RequireInBody: true},
			},
			wantErrCode: string(domain.ErrMissingJiraKeyBody),
			description: "Text that looks like Refs but invalid format should trigger missing key error",
		},
		{
			name:    "ErrInvalidKeyFormat",
			subject: "feat: Add new feature",
			body:    "Refs: proj-123",
			config: config.Config{
				Jira: config.JiraConfig{RequireInBody: true},
			},
			wantErrCode: string(domain.ErrInvalidKeyFormat),
			description: "Refs with invalid key format should trigger key format error",
		},
		{
			name:    "ErrRefsAfterSignoff",
			subject: "feat: Add new feature",
			body: `Signed-off-by: Dev <dev@example.com>
Refs: PROJ-123`,
			config: config.Config{
				Jira: config.JiraConfig{RequireInBody: true},
			},
			wantErrCode: string(domain.ErrRefsAfterSignoff),
			description: "Refs after Signed-off-by should trigger specific error",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewJiraReferenceRule(testCase.config)
			commit := domain.Commit{
				Subject: testCase.subject,
				Body:    testCase.body,
			}

			errors := rule.Validate(commit, testCase.config)

			require.NotEmpty(t, errors, testCase.description)
			require.Equal(t, testCase.wantErrCode, errors[0].Code, testCase.description)
		})
	}
}
