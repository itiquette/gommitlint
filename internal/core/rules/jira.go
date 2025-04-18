// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// Common regex patterns compiled once at package level.
var (
	jiraKeyRegex  = regexp.MustCompile(`([A-Z]+-\d+)`)
	refsLineRegex = regexp.MustCompile(`^Refs:\s*([A-Z]+-\d+(?:\s*,\s*[A-Z]+-\d+)*)$`)
)

// JiraReferenceRule enforces proper Jira issue references in commit messages.
//
// This rule ensures that commit messages include valid Jira issue keys (e.g., PROJECT-123)
// according to the project's conventions for placement and format. It helps maintain
// traceability between code changes and issue tracking systems, making it easier to
// understand the purpose and context of each commit.
//
// The rule supports two validation modes:
//  1. Subject-based validation - Checks for Jira keys in the commit subject line
//  2. Body-based validation - Checks for Jira keys in dedicated "Refs:" lines in the commit body
type JiraReferenceRule struct {
	*BaseRule

	// Store information for verbose output
	foundKeys       []string
	validateBodyRef bool
	validProjects   []string
	isConventional  bool
}

// JiraReferenceOption is a function that modifies a JiraReferenceRule.
type JiraReferenceOption func(*JiraReferenceRule)

// WithValidProjects sets the list of permitted Jira project keys.
func WithValidProjects(projects []string) JiraReferenceOption {
	return func(rule *JiraReferenceRule) {
		rule.validProjects = projects
	}
}

// WithBodyRefChecking enables validation of Jira references in the commit body.
func WithBodyRefChecking() JiraReferenceOption {
	return func(rule *JiraReferenceRule) {
		rule.validateBodyRef = true
	}
}

// WithConventionalCommit enables conventional commit format handling.
func WithConventionalCommit() JiraReferenceOption {
	return func(rule *JiraReferenceRule) {
		rule.isConventional = true
	}
}

// WithJiraRequired is removed as the Jira requirement is now handled
// by marking the rule as run when disabled

// NewJiraReferenceRule creates a new JiraReferenceRule with the specified options.
func NewJiraReferenceRule(options ...JiraReferenceOption) *JiraReferenceRule {
	rule := &JiraReferenceRule{
		BaseRule:        NewBaseRule("JiraReference"),
		foundKeys:       []string{},
		validateBodyRef: false,
		validProjects:   []string{},
		isConventional:  false,
	}

	// Apply provided options
	for _, option := range options {
		option(rule)
	}

	return rule
}

// The Name method is provided by BaseRule.

// Result returns a concise rule message.
func (j *JiraReferenceRule) Result() string {
	if j.HasErrors() {
		return "Missing or invalid Jira reference"
	}

	return "Valid Jira reference"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (j *JiraReferenceRule) VerboseResult() string {
	if j.HasErrors() {
		errors := j.Errors()
		if len(errors) == 0 {
			return "Unknown error"
		}

		// errors[0] is already a ValidationError, so no need for type assertion
		validationErr := errors[0]

		// Return a more detailed error message in verbose mode
		switch validationErr.Code {
		case "empty_subject":
			return "Commit subject is empty. Cannot validate Jira references."
		case "missing_jira_key_body":
			return "No Jira issue key found in commit body with 'Refs:' prefix."
		case "missing_jira_key_subject":
			return "No Jira issue key found in commit subject. Must include reference like PROJ-123."
		case "key_not_at_end":
			var key string

			for k, v := range validationErr.Context {
				if k == "key" {
					key = v

					break
				}
			}

			return "Jira key '" + key + "' must be at the end of the conventional commit subject line."
		case "invalid_project":
			var project string

			for k, v := range validationErr.Context {
				if k == "project" {
					project = v

					break
				}
			}

			validProjects := ""
			if len(j.validProjects) > 0 {
				validProjects = " Valid projects: " + strings.Join(j.validProjects, ", ")
			}

			return "Invalid Jira project '" + project + "'. Not in list of valid projects." + validProjects
		case "invalid_refs_format":
			var line string

			for k, v := range validationErr.Context {
				if k == "line" {
					line = v

					break
				}
			}

			return "Invalid 'Refs:' format: '" + line + "'. Should be 'Refs: PROJ-123' or 'Refs: PROJ-123, PROJ-456'."
		case "refs_after_signoff":
			return "'Refs:' line appears after 'Signed-off-by' line. 'Refs:' must come first."
		case "invalid_key_format":
			var key string

			for k, v := range validationErr.Context {
				if k == "key" {
					key = v

					break
				}
			}

			return "Invalid Jira key format: '" + key + "'. Must follow the pattern PROJECT-123."
		default:
			return validationErr.Error()
		}
	}

	// Success message with more details
	if j.validateBodyRef {
		return "Valid Jira reference(s) found in commit body: " + strings.Join(j.foundKeys, ", ")
	}

	return "Valid Jira reference(s) found in commit subject: " + strings.Join(j.foundKeys, ", ")
}

// The Errors method is provided by BaseRule.

// Removed addError function as part of Phase 3 error handling refactoring
// Direct use of error codes from appErrors is now required

// Validate performs validation against a commit and returns any errors.
func (j *JiraReferenceRule) Validate(commit *domain.CommitInfo) []appErrors.ValidationError {
	// Reset errors and found keys
	j.ClearErrors()
	j.MarkAsRun()
	j.foundKeys = []string{}

	subject := commit.Subject
	body := commit.Body

	// Normalize and trim the subject
	subject = strings.TrimSpace(subject)
	if subject == "" {
		j.AddAppError(
			appErrors.ErrEmptyMessage,
			"Commit subject is empty",
		)

		return j.Errors()
	}

	// Validate based on the configured strategy
	if j.validateBodyRef {
		j.validateBodyReferences(body)
	} else {
		j.validateSubjectReferences(subject)
	}

	return j.Errors()
}

// Help returns a description of how to fix the rule violation.
func (j *JiraReferenceRule) Help() string {
	if !j.HasErrors() {
		return "No errors to fix"
	}

	// Check error code
	errors := j.Errors()
	if len(errors) > 0 {
		// Cast the error to appErrors.ValidationError to access its fields
		// errors[0] is already a ValidationError, so no need for type assertion
		validationErr := errors[0]

		switch validationErr.Code {
		case string(appErrors.ErrMissingJira):
			// For missing Jira references, use template help
			if j.validateBodyRef {
				return `Include a valid Jira issue key (e.g., PROJECT-123) in your commit body with the "Refs:" prefix.

Examples:
- Refs: PROJECT-123
- Refs: PROJECT-123, PROJECT-456
- Refs: PROJECT-123, PROJECT-456, PROJECT-789

The Refs: line should appear at the end of the commit body, before any Signed-off-by lines.`
			}

			return `Include a valid Jira issue key (e.g., PROJECT-123) in your commit subject.

For conventional commits, place the Jira key at the end of the first line:
- feat(auth): add login feature PROJ-123
- fix: resolve timeout issue [PROJ-123]
- docs(readme): update installation steps (PROJ-123)

For other commit formats, include the Jira key anywhere in the subject.`

		case string(appErrors.ErrInvalidType):
			// For invalid project types
			projectKeys := j.validProjects
			if len(projectKeys) > 0 {
				return `The Jira project reference is not recognized as a valid project.

Valid projects: ` + strings.Join(projectKeys, ", ") + `

Please use one of these project keys in your Jira reference.`
			}

			return `The Jira project reference is not valid.
Jira project keys should be uppercase letters followed by a hyphen and numbers (e.g., PROJECT-123).`

		case string(appErrors.ErrInvalidFormat):
			// This handles multiple format errors with specific messages
			if j.validateBodyRef {
				// For body validation errors
				if strings.Contains(validationErr.Message, "must appear before") {
					return `The "Refs:" line must appear before any "Signed-off-by" lines in your commit message.

The correct order is:
1. Commit subject
2. Blank line
3. Commit body (if any)
4. Refs: line(s)
5. Signed-off-by line(s)`
				}

				if strings.Contains(validationErr.Message, "invalid Refs format") {
					return `The "Refs:" line in your commit body has an invalid format.

The correct format is:
Refs: PROJECT-123
or for multiple references:
Refs: PROJECT-123, PROJECT-456

Make sure:
- "Refs:" is at the beginning of the line
- Project keys follow the format PROJECT-123
- Multiple references are separated by commas
- The Refs line appears before any Signed-off-by lines`
				}
			} else if strings.Contains(validationErr.Message, "must be at the end") {
				return `In conventional commit format, place the Jira issue key at the end of the first line.

Examples:
- feat(auth): add login feature PROJ-123
- fix: resolve timeout issue [PROJ-123]
- docs(readme): update installation steps (PROJ-123)

Avoid putting the Jira key in the middle of the line:
- INCORRECT: feat(PROJ-123): add login feature
- INCORRECT: fix: PROJ-123 resolve timeout issue`
			} else if strings.Contains(validationErr.Message, "empty") {
				return "Provide a non-empty commit message with a Jira issue reference"
			} else if strings.Contains(validationErr.Message, "invalid Jira issue key format") {
				// Use the general format template
				return `Invalid Jira issue key format. Make sure it follows the pattern PROJECT-123.

Jira keys should be uppercase letters followed by a hyphen and numbers (e.g., PROJECT-123).`
			}

			// General invalid format
			return `The commit message format is invalid. Make sure it follows the expected pattern.

For conventional commits:
- feat(scope): description PROJ-123

For body references:
- Subject line
- 
- Refs: PROJ-123`
		}
	}

	// Default help
	return `Ensure your commit message contains a valid Jira issue reference.
The Jira issue key should follow the format PROJECT-123.`
}

// This is now handled by the new addError method above

// validateSubjectReferences validates Jira references in the commit subject.
func (j *JiraReferenceRule) validateSubjectReferences(subject string) {
	lines := strings.Split(subject, "\n")
	firstLine := lines[0]

	if j.isConventional {
		j.validateConventionalCommitSubject(firstLine)
	} else {
		j.validateNonConventionalCommitSubject(subject)
	}
}

// validateConventionalCommitSubject validates a conventional commit subject line.
func (j *JiraReferenceRule) validateConventionalCommitSubject(firstLine string) {
	matches := jiraKeyRegex.FindAllString(firstLine, -1)
	if len(matches) == 0 {
		j.AddErrorWithContext(
			appErrors.ErrMissingJira,
			"no Jira issue key found in the commit subject",
			map[string]string{
				"subject": firstLine,
			},
		)

		return
	}

	// Validate position for conventional commit format
	if !j.validateKeyPositionInConventional(firstLine, matches) {
		return
	}

	// Extract the last match
	lastMatch := matches[len(matches)-1]

	// Validate the project
	j.validateJiraProject(lastMatch)
}

// validateKeyPositionInConventional checks if the Jira key is at the end of the subject line.
func (j *JiraReferenceRule) validateKeyPositionInConventional(firstLine string, matches []string) bool {
	// Get the last match
	lastMatch := matches[len(matches)-1]

	// Check if the last match is at the end of the first line
	// Supporting common formats: PROJ-123, [PROJ-123], (PROJ-123)
	validSuffixes := []string{
		lastMatch,
		"[" + lastMatch + "]",
		"(" + lastMatch + ")",
	}

	found := false

	for _, suffix := range validSuffixes {
		if strings.HasSuffix(firstLine, suffix) {
			found = true

			j.foundKeys = append(j.foundKeys, lastMatch)

			break
		}
	}

	if !found {
		j.AddErrorWithContext(
			appErrors.ErrInvalidFormat,
			"in conventional commit format, Jira issue key must be at the end of the first line",
			map[string]string{
				"subject": firstLine,
				"key":     lastMatch,
			},
		)

		return false
	}

	return true
}

// validateNonConventionalCommitSubject validates a non-conventional commit subject.
func (j *JiraReferenceRule) validateNonConventionalCommitSubject(subject string) {
	matches := jiraKeyRegex.FindAllString(subject, -1)
	if len(matches) == 0 {
		j.AddErrorWithContext(
			appErrors.ErrMissingJira,
			"no Jira issue key found in the commit subject",
			map[string]string{
				"subject": subject,
			},
		)

		return
	}

	// Store all found keys
	j.foundKeys = matches

	// Validate all found project keys
	j.validateAllFoundProjects(matches)
}

// validateAllFoundProjects validates all Jira project references found.
func (j *JiraReferenceRule) validateAllFoundProjects(matches []string) {
	for _, match := range matches {
		if !j.validateJiraProject(match) {
			return // Stop on first invalid project
		}
	}
}

// validateBodyReferences validates Jira references in the commit body.
func (j *JiraReferenceRule) validateBodyReferences(body string) {
	body = strings.TrimSpace(body)
	if body == "" {
		j.AddErrorWithContext(
			appErrors.ErrMissingJira,
			"no Jira issue key found in the commit body",
			nil,
		)

		return
	}

	// Look for "Refs:" lines
	bodyLines := strings.Split(body, "\n")

	// Check format and locate references
	lineInfo := j.findRefsAndSignoffLines(bodyLines)
	if !lineInfo.success {
		return // Error was already added
	}

	// Validate reference ordering
	if !j.validateReferenceLineOrdering(lineInfo) {
		return // Error was already added
	}

	// Validate the Jira keys in the Refs: line
	j.validateJiraKeysInRefLine(bodyLines, lineInfo.refsLine)
}

// RefsLineInfo holds information about Refs and Signoff lines.
type RefsLineInfo struct {
	foundRefs      bool
	signOffFound   bool
	signOffLineNum int
	refsLineNum    int
	refsLine       int // Alias for refsLineNum for readability
	success        bool
}

// findRefsAndSignoffLines locates and validates "Refs:" and "Signed-off-by:" lines.
func (j *JiraReferenceRule) findRefsAndSignoffLines(bodyLines []string) RefsLineInfo {
	info := RefsLineInfo{
		foundRefs:      false,
		signOffFound:   false,
		signOffLineNum: -1,
		refsLineNum:    -1,
		refsLine:       -1,
		success:        true,
	}

	// First pass: find Refs: and Signed-off-by lines
	for ind, line := range bodyLines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Signed-off-by:") {
			info.signOffFound = true
			info.signOffLineNum = ind
		}

		if refsLineRegex.MatchString(line) {
			info.foundRefs = true
			info.refsLineNum = ind
			info.refsLine = ind
		} else if strings.HasPrefix(line, "Refs:") {
			// Line starts with Refs: but doesn't match the expected format
			j.AddErrorWithContext(
				appErrors.ErrInvalidFormat,
				"invalid Refs format in commit body, should be 'Refs: PROJ-123' or 'Refs: PROJ-123, PROJ-456'",
				map[string]string{
					"line": line,
				},
			)

			info.success = false

			return info
		}
	}

	// Validate that Refs: exists
	if !info.foundRefs {
		j.AddErrorWithContext(
			appErrors.ErrMissingJira,
			"no Jira issue key found in the commit body with Refs: prefix",
			nil,
		)

		info.success = false

		return info
	}

	return info
}

// validateReferenceLineOrdering ensures that Refs: appears before Signed-off-by.
func (j *JiraReferenceRule) validateReferenceLineOrdering(info RefsLineInfo) bool {
	// Validate that Refs: appears before any Signed-off-by lines
	if info.signOffFound && info.refsLineNum > info.signOffLineNum {
		j.AddErrorWithContext(
			appErrors.ErrInvalidFormat,
			"Refs: line must appear before any Signed-off-by lines",
			map[string]string{
				"refs_line":    strconv.Itoa(info.refsLineNum),
				"signoff_line": strconv.Itoa(info.signOffLineNum),
			},
		)

		return false
	}

	return true
}

// validateJiraKeysInRefLine validates the Jira keys in the first Refs: line.
func (j *JiraReferenceRule) validateJiraKeysInRefLine(bodyLines []string, _ int) {
	for _, line := range bodyLines {
		line = strings.TrimSpace(line)
		if refsLineRegex.MatchString(line) {
			// Extract and validate all Jira keys
			matches := jiraKeyRegex.FindAllString(line, -1)
			j.foundKeys = matches

			for _, match := range matches {
				if !j.validateJiraProject(match) {
					return // Stop on first invalid project
				}
			}

			break // Process only the first Refs: line
		}
	}
}

// validateJiraProject checks if a Jira issue key is valid.
func (j *JiraReferenceRule) validateJiraProject(jiraKey string) bool {
	// First, verify the key has the correct format
	if !jiraKeyRegex.MatchString(jiraKey) {
		j.AddErrorWithContext(
			appErrors.ErrInvalidFormat,
			"invalid Jira issue key format: "+jiraKey+" (should be PROJECT-123)",
			map[string]string{
				"key": jiraKey,
			},
		)

		return false
	}

	// If no project list is provided, just validate the format
	if len(j.validProjects) == 0 {
		return true
	}

	// When project list is provided, validate against it
	projectKey := strings.Split(jiraKey, "-")[0]
	if !containsString(j.validProjects, projectKey) {
		j.AddErrorWithContext(
			appErrors.ErrInvalidType,
			"Jira project "+projectKey+" is not a valid project",
			map[string]string{
				"project":        projectKey,
				"valid_projects": strings.Join(j.validProjects, ","),
			},
		)

		return false
	}

	return true
}

// containsString checks if a string is present in a slice of strings.
func containsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}

	return false
}
