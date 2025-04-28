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
type JiraReferenceRule struct {
	BaseRule        BaseRule
	foundKeys       []string
	validateBodyRef bool
	validProjects   []string
	isConventional  bool
}

// JiraReferenceOption is a function that modifies a JiraReferenceRule.
type JiraReferenceOption func(JiraReferenceRule) JiraReferenceRule

// WithValidProjects sets the list of permitted Jira project keys.
func WithValidProjects(projects []string) JiraReferenceOption {
	return func(rule JiraReferenceRule) JiraReferenceRule {
		result := rule
		result.validProjects = projects

		return result
	}
}

// WithBodyRefChecking enables validation of Jira references in the commit body.
func WithBodyRefChecking() JiraReferenceOption {
	return func(rule JiraReferenceRule) JiraReferenceRule {
		result := rule
		result.validateBodyRef = true

		return result
	}
}

// WithConventionalCommit enables conventional commit format handling.
func WithConventionalCommit() JiraReferenceOption {
	return func(rule JiraReferenceRule) JiraReferenceRule {
		result := rule
		result.isConventional = true

		return result
	}
}

// NewJiraReferenceRule creates a new JiraReferenceRule with the specified options.
func NewJiraReferenceRule(options ...JiraReferenceOption) JiraReferenceRule {
	rule := JiraReferenceRule{
		BaseRule:        NewBaseRule("JiraReference"),
		foundKeys:       []string{},
		validateBodyRef: false,
		validProjects:   []string{},
		isConventional:  false,
	}

	// Apply provided options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// NewJiraReferenceRuleWithConfig creates a JiraReferenceRule using a configuration provider.
func NewJiraReferenceRuleWithConfig(jiraConfig domain.JiraConfigProvider, conventionalConfig domain.ConventionalConfigProvider) JiraReferenceRule {
	// Build options based on the configuration
	var options []JiraReferenceOption

	// Check if conventional commit format is required
	if conventionalConfig.ConventionalRequired() {
		options = append(options, WithConventionalCommit())
	}

	// Check if body reference checking is enabled
	if jiraConfig.JiraBodyRef() {
		options = append(options, WithBodyRefChecking())
	}

	// Add valid projects if provided
	if projects := jiraConfig.JiraProjects(); len(projects) > 0 {
		options = append(options, WithValidProjects(projects))
	}

	return NewJiraReferenceRule(options...)
}

// Result returns a concise rule message.
func (j JiraReferenceRule) Result() string {
	if j.HasErrors() {
		errors := j.Errors()
		if len(errors) > 0 {
			validationErr := errors[0]
			switch validationErr.Code {
			case string(appErrors.ErrMissingJira):
				return "Missing Jira issue key"
			case string(appErrors.ErrInvalidFormat):
				if key, exists := validationErr.Context["key"]; exists {
					return "Invalid Jira key format: " + key
				}

				return "Invalid Jira key format"
			case string(appErrors.ErrInvalidType):
				if project, exists := validationErr.Context["project"]; exists {
					return "Invalid Jira project: " + project
				}

				return "Invalid Jira project"
			default:
				return "Invalid Jira reference"
			}
		}

		return "Missing or invalid Jira reference"
	}

	return "Valid Jira reference"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (j JiraReferenceRule) VerboseResult() string {
	if j.HasErrors() {
		errors := j.Errors()
		if len(errors) == 0 {
			return "Unknown error"
		}
		// errors[0] is already a ValidationError, so no need for type assertion
		validationErr := errors[0]
		// Return a more detailed error message in verbose mode
		switch validationErr.Code {
		case string(appErrors.ErrEmptyMessage):
			return "Commit subject is empty. Cannot validate Jira references."
		case string(appErrors.ErrMissingJira):
			if strings.Contains(validationErr.Message, "body") {
				return "No Jira issue key found in commit body with 'Refs:' prefix."
			}

			return "No Jira issue key found in commit subject. Must include reference like PROJ-123."
		case string(appErrors.ErrInvalidFormat):
			if strings.Contains(validationErr.Message, "end") {
				var key string

				if ctx := validationErr.Context; ctx != nil {
					if v, ok := ctx["key"]; ok {
						key = v
					}
				}

				return "Jira key '" + key + "' must be at the end of the conventional commit subject line."
			} else if strings.Contains(validationErr.Message, "Refs:") {
				var line string

				if ctx := validationErr.Context; ctx != nil {
					if v, ok := ctx["line"]; ok {
						line = v
					}
				}

				return "Invalid 'Refs:' format: '" + line + "'. Should be 'Refs: PROJ-123' or 'Refs: PROJ-123, PROJ-456'."
			} else if strings.Contains(validationErr.Message, "Signed-off-by") {
				return "'Refs:' line appears after 'Signed-off-by' line. 'Refs:' must come first."
			} else if strings.Contains(validationErr.Message, "key format") {
				var key string

				if ctx := validationErr.Context; ctx != nil {
					if v, ok := ctx["key"]; ok {
						key = v
					}
				}

				return "Invalid Jira key format: '" + key + "'. Must follow the pattern PROJECT-123."
			}
		case string(appErrors.ErrInvalidType):
			var project string

			if ctx := validationErr.Context; ctx != nil {
				if v, ok := ctx["project"]; ok {
					project = v
				}
			}

			validProjects := ""
			if len(j.validProjects) > 0 {
				validProjects = " Valid projects: " + strings.Join(j.validProjects, ", ")
			}

			return "Invalid Jira project '" + project + "'. Not in list of valid projects." + validProjects
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

// Help returns a description of how to fix the rule violation.
func (j JiraReferenceRule) Help() string {
	// First check if the rule has errors - this should be the primary check
	if j.HasErrors() {
		errors := j.Errors()
		if len(errors) > 0 {
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
			default:
				// Check message for clues if it's a non-standard validation error
				if strings.Contains(validationErr.Message, "no Jira issue key") {
					return `Include a valid Jira issue key (e.g., PROJECT-123) in your commit.
For conventional commits, place the Jira key at the end of the first line:
- feat(auth): add login feature PROJ-123
- fix: resolve timeout issue [PROJ-123]

If body references are enabled, use the "Refs:" prefix:
- Refs: PROJECT-123
- Refs: PROJECT-123, PROJECT-456`
				}

				// Default error help message if none of the above conditions match
				return `Include a valid Jira issue key (e.g., PROJECT-123) in your commit message.
The Jira issue key should follow the format PROJECT-123.

For conventional commits, place the key at the end of the subject:
- feat(auth): add feature PROJ-123
- fix: resolve timeout issue [PROJ-123]

For body references, use the "Refs:" line in the commit body:
- Refs: PROJ-123`
			}
		}

		// Fallback for when j.HasErrors() is true but there are no specific errors
		return `Include a valid Jira issue key (e.g., PROJECT-123) in your commit message.
The Jira issue key should follow the format PROJECT-123.

For conventional commits, place the key at the end of the subject:
- feat(auth): add feature PROJ-123
- fix: resolve timeout issue [PROJ-123]

For body references, use the "Refs:" line in the commit body:
- Refs: PROJ-123`
	}

	// Success case (when j.HasErrors() is false)
	if j.validateBodyRef {
		return `Commit message contains valid Jira issue reference(s) in the body using the correct "Refs:" format.
This rule checks for properly formatted Jira issue references in the commit body.`
	}

	return `Commit message contains valid Jira issue reference(s) in the subject line with correct format (PROJECT-123).
This rule checks for properly formatted Jira issue references in the commit subject.`
}

// validateJiraWithState validates a commit and returns both errors and an updated rule state.
func validateJiraWithState(rule JiraReferenceRule, commit domain.CommitInfo) ([]appErrors.ValidationError, JiraReferenceRule) {
	// Start with a clean slate and mark as run
	updatedRule := rule
	updatedRule.BaseRule = updatedRule.BaseRule.WithClearedErrors().WithRun()

	subject := commit.Subject
	body := commit.Body

	// Normalize and trim the subject
	subject = strings.TrimSpace(subject)
	if subject == "" {
		// Create a validation error for empty subject
		err := appErrors.New(
			updatedRule.Name(),
			appErrors.ErrEmptyMessage,
			"Commit subject is empty",
			appErrors.WithContextMap(map[string]string{}),
		)

		// Add error and return
		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)

		return []appErrors.ValidationError{err}, updatedRule
	}

	// Validate based on the configured strategy
	var errors []appErrors.ValidationError
	if rule.validateBodyRef {
		errors = rule.validateBodyReferences(body)
	} else {
		errors = rule.validateSubjectReferences(subject)
	}

	// Add errors to the updated rule
	for _, err := range errors {
		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
	}

	// Set found keys if available from validation
	if len(errors) == 0 && subject != "" {
		matches := jiraKeyRegex.FindAllString(subject, -1)
		if len(matches) > 0 {
			updatedRule = updatedRule.SetFoundKeys(matches)
		}
	}

	return errors, updatedRule
}

// Validate performs validation against a commit and returns any errors.
// This uses value semantics and does not modify the rule's state.
func (j JiraReferenceRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
	errors, _ := validateJiraWithState(j, commit)

	return errors
}

// ValidateJiraWithState is the exported version of validateJiraWithState.
// This is needed for testing but follows the same pure function approach.
func ValidateJiraWithState(rule JiraReferenceRule, commit domain.CommitInfo) ([]appErrors.ValidationError, JiraReferenceRule) {
	return validateJiraWithState(rule, commit)
}

// Name returns the rule name.
func (j JiraReferenceRule) Name() string {
	return j.BaseRule.Name()
}

// Errors returns all validation errors.
func (j JiraReferenceRule) Errors() []appErrors.ValidationError {
	return j.BaseRule.Errors()
}

// HasErrors checks if there are any validation errors.
func (j JiraReferenceRule) HasErrors() bool {
	return j.BaseRule.HasErrors()
}

// SetErrors sets the errors for this rule and returns a new instance.
func (j JiraReferenceRule) SetErrors(errors []appErrors.ValidationError) JiraReferenceRule {
	result := j

	// Update BaseRule with errors
	baseRule := j.BaseRule.WithClearedErrors()
	for _, err := range errors {
		baseRule = baseRule.WithError(err)
	}

	result.BaseRule = baseRule

	return result
}

// SetFoundKeys sets the found keys and returns a new instance.
func (j JiraReferenceRule) SetFoundKeys(keys []string) JiraReferenceRule {
	result := j
	result.foundKeys = keys

	return result
}

// createError creates a validation error without modifying the rule's state.
func (j JiraReferenceRule) createError(code appErrors.ValidationErrorCode, message string, context map[string]string) appErrors.ValidationError {
	// Create the validation error
	return appErrors.ValidationError{
		Rule:    j.Name(),
		Code:    string(code),
		Message: message,
		Context: context,
	}
}

// validateSubjectReferences validates Jira references in the commit subject.
func (j JiraReferenceRule) validateSubjectReferences(subject string) []appErrors.ValidationError {
	lines := strings.Split(subject, "\n")

	firstLine := lines[0]
	if j.isConventional {
		return j.validateConventionalCommitSubject(firstLine)
	}

	return j.validateNonConventionalCommitSubject(subject)
}

// validateConventionalCommitSubject validates a conventional commit subject line.
func (j JiraReferenceRule) validateConventionalCommitSubject(firstLine string) []appErrors.ValidationError {
	errors := make([]appErrors.ValidationError, 0)

	matches := jiraKeyRegex.FindAllString(firstLine, -1)
	if len(matches) == 0 {
		// Check for invalid format key like PROJ123 (without hyphen)
		invalidFormatRegex := regexp.MustCompile(`([A-Z]+\d+)`)
		invalidMatches := invalidFormatRegex.FindAllString(firstLine, -1)

		if len(invalidMatches) > 0 {
			// Found something that looks like a Jira key but with invalid format
			err := j.createError(
				appErrors.ErrInvalidFormat,
				"invalid Jira issue key format: "+invalidMatches[0]+" (should be PROJECT-123)",
				map[string]string{
					"key": invalidMatches[0],
				},
			)
			errors = append(errors, err)

			return errors
		}

		// No key-like patterns found, report as missing
		err := j.createError(
			appErrors.ErrMissingJira,
			"no Jira issue key found in the commit subject",
			map[string]string{
				"subject": firstLine,
			},
		)
		errors = append(errors, err)

		return errors
	}

	// Validate position for conventional commit format
	// This will return any errors or empty slice
	errors = j.validateKeyPositionInConventional(firstLine, matches)
	if len(errors) > 0 {
		return errors
	}

	// Extract the last match
	lastMatch := matches[len(matches)-1]

	// Validate the project
	return j.validateJiraProject(lastMatch)
}

// validateKeyPositionInConventional checks if the Jira key is at the end of the subject line.
func (j JiraReferenceRule) validateKeyPositionInConventional(firstLine string, matches []string) []appErrors.ValidationError {
	errors := make([]appErrors.ValidationError, 0)

	// Get the last match
	lastMatch := matches[len(matches)-1]
	// Check if the last match is at the end of the first line
	// Supporting common formats: PROJ-123, [PROJ-123], (PROJ-123)
	validSuffixes := []string{
		lastMatch,
		"[" + lastMatch + "]",
		"(" + lastMatch + ")",
	}

	for _, suffix := range validSuffixes {
		if strings.HasSuffix(firstLine, suffix) {
			// Found at the end - keep track for downstream methods
			foundKeys := append(j.foundKeys, lastMatch)
			jWithKeys := j.SetFoundKeys(foundKeys)
			_ = jWithKeys // In a purely functional approach, we'd return this or pass to a continuation

			// No errors to return
			return errors
		}
	}

	// Not found at the end, report error
	err := j.createError(
		appErrors.ErrInvalidFormat,
		"in conventional commit format, Jira issue key must be at the end of the first line",
		map[string]string{
			"subject": firstLine,
			"key":     lastMatch,
		},
	)
	errors = append(errors, err)

	return errors
}

// validateNonConventionalCommitSubject validates a non-conventional commit subject.
func (j JiraReferenceRule) validateNonConventionalCommitSubject(subject string) []appErrors.ValidationError {
	errors := make([]appErrors.ValidationError, 0)

	matches := jiraKeyRegex.FindAllString(subject, -1)

	// Special handling for invalid format like PROJ123 (without hyphen)
	if len(matches) == 0 {
		// Check if there's a pattern that looks like a Jira key but without the hyphen
		invalidFormatRegex := regexp.MustCompile(`([A-Z]+\d+)`)
		invalidMatches := invalidFormatRegex.FindAllString(subject, -1)

		if len(invalidMatches) > 0 {
			// Found something that looks like a Jira key but with invalid format
			err := j.createError(
				appErrors.ErrInvalidFormat,
				"invalid Jira issue key format: "+invalidMatches[0]+" (should be PROJECT-123)",
				map[string]string{
					"key": invalidMatches[0],
				},
			)
			errors = append(errors, err)

			return errors
		}

		// No key-like patterns found, report as missing
		err := j.createError(
			appErrors.ErrMissingJira,
			"no Jira issue key found in the commit subject",
			map[string]string{
				"subject": subject,
			},
		)
		errors = append(errors, err)

		return errors
	}

	// Store all found keys for downstream methods
	jWithKeys := j.SetFoundKeys(matches)
	_ = jWithKeys // In purely functional code, we'd pass this forward

	// Validate all found project keys
	return j.validateAllFoundProjects(matches)
}

// validateAllFoundProjects validates all Jira project references found.
func (j JiraReferenceRule) validateAllFoundProjects(matches []string) []appErrors.ValidationError {
	errors := make([]appErrors.ValidationError, 0)

	for _, match := range matches {
		projectErrors := j.validateJiraProject(match)
		if len(projectErrors) > 0 {
			errors = append(errors, projectErrors...)

			break // Stop on first error
		}
	}

	return errors
}

// validateBodyReferences validates Jira references in the commit body.
func (j JiraReferenceRule) validateBodyReferences(body string) []appErrors.ValidationError {
	errors := make([]appErrors.ValidationError, 0)

	body = strings.TrimSpace(body)
	if body == "" {
		err := j.createError(
			appErrors.ErrMissingJira,
			"no Jira issue key found in the commit body",
			nil,
		)
		errors = append(errors, err)

		return errors
	}

	// Look for "Refs:" lines
	bodyLines := strings.Split(body, "\n")

	// Check format and locate references
	lineInfo, lineErrors := j.findRefsAndSignoffLines(bodyLines)
	if len(lineErrors) > 0 {
		errors = append(errors, lineErrors...)

		return errors
	}

	// Validate reference ordering
	orderErrors := j.validateReferenceLineOrdering(lineInfo)
	if len(orderErrors) > 0 {
		errors = append(errors, orderErrors...)

		return errors
	}

	// Validate the Jira keys in the Refs: line
	return j.validateJiraKeysInRefLine(bodyLines, lineInfo.refsLine)
}

// RefsLineInfo holds information about Refs and Signoff lines.
type RefsLineInfo struct {
	foundRefs      bool
	signOffFound   bool
	signOffLineNum int
	refsLineNum    int
	refsLine       int // Alias for refsLineNum for readability
}

// findRefsAndSignoffLines locates and validates "Refs:" and "Signed-off-by:" lines.
func (j JiraReferenceRule) findRefsAndSignoffLines(bodyLines []string) (RefsLineInfo, []appErrors.ValidationError) {
	errors := make([]appErrors.ValidationError, 0)

	info := RefsLineInfo{
		foundRefs:      false,
		signOffFound:   false,
		signOffLineNum: -1,
		refsLineNum:    -1,
		refsLine:       -1,
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
			// Check for invalid format keys like PROJ123 (without hyphen)
			invalidFormatRegex := regexp.MustCompile(`Refs:\s*([A-Z]+\d+)`)
			if invalidFormatRegex.MatchString(line) {
				matches := invalidFormatRegex.FindStringSubmatch(line)
				invalidKey := matches[1]
				err := j.createError(
					appErrors.ErrInvalidFormat,
					"invalid Jira issue key format: "+invalidKey+" (should be PROJECT-123)",
					map[string]string{
						"key":  invalidKey,
						"line": line,
					},
				)
				errors = append(errors, err)

				return info, errors
			}

			// Regular invalid format
			err := j.createError(
				appErrors.ErrInvalidFormat,
				"invalid Refs format in commit body, should be 'Refs: PROJ-123' or 'Refs: PROJ-123, PROJ-456'",
				map[string]string{
					"line": line,
				},
			)
			errors = append(errors, err)

			return info, errors
		}
	}

	// Validate that Refs: exists
	if !info.foundRefs {
		err := j.createError(
			appErrors.ErrMissingJira,
			"no Jira issue key found in the commit body with Refs: prefix",
			nil,
		)
		errors = append(errors, err)

		return info, errors
	}

	return info, errors
}

// validateReferenceLineOrdering ensures that Refs: appears before Signed-off-by.
func (j JiraReferenceRule) validateReferenceLineOrdering(info RefsLineInfo) []appErrors.ValidationError {
	errors := make([]appErrors.ValidationError, 0)

	// Validate that Refs: appears before any Signed-off-by lines
	if info.signOffFound && info.refsLineNum > info.signOffLineNum {
		err := j.createError(
			appErrors.ErrInvalidFormat,
			"Refs: line must appear before any Signed-off-by lines",
			map[string]string{
				"refs_line":    strconv.Itoa(info.refsLineNum),
				"signoff_line": strconv.Itoa(info.signOffLineNum),
			},
		)
		errors = append(errors, err)
	}

	return errors
}

// validateJiraKeysInRefLine validates the Jira keys in the first Refs: line.
func (j JiraReferenceRule) validateJiraKeysInRefLine(bodyLines []string, _ int) []appErrors.ValidationError {
	errors := make([]appErrors.ValidationError, 0)

	var foundKeys []string

	for _, line := range bodyLines {
		line = strings.TrimSpace(line)
		if refsLineRegex.MatchString(line) {
			// Extract and validate all Jira keys
			matches := jiraKeyRegex.FindAllString(line, -1)
			foundKeys = matches

			// Store for downstream methods
			jWithKeys := j.SetFoundKeys(foundKeys)
			_ = jWithKeys // In purely functional code, we'd pass this forward

			for _, match := range matches {
				projectErrors := j.validateJiraProject(match)
				if len(projectErrors) > 0 {
					errors = append(errors, projectErrors...)

					return errors
				}
			}

			break // Process only the first Refs: line
		}
	}

	return errors
}

// validateJiraProject checks if a Jira issue key is valid.
func (j JiraReferenceRule) validateJiraProject(jiraKey string) []appErrors.ValidationError {
	errors := make([]appErrors.ValidationError, 0)

	// First, verify the key has the correct format
	if !jiraKeyRegex.MatchString(jiraKey) {
		err := j.createError(
			appErrors.ErrInvalidFormat,
			"invalid Jira issue key format: "+jiraKey+" (should be PROJECT-123)",
			map[string]string{
				"key": jiraKey,
			},
		)
		errors = append(errors, err)

		return errors
	}

	// If no project list is provided, just validate the format
	if len(j.validProjects) == 0 {
		return errors
	}

	// When project list is provided, validate against it
	projectKey := strings.Split(jiraKey, "-")[0]
	if !containsString(j.validProjects, projectKey) {
		err := j.createError(
			appErrors.ErrInvalidType,
			"Jira project "+projectKey+" is not a valid project",
			map[string]string{
				"project":        projectKey,
				"valid_projects": strings.Join(j.validProjects, ","),
			},
		)
		errors = append(errors, err)
	}

	return errors
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
