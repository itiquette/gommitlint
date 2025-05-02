// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
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

// NewJiraReferenceRuleWithConfig creates a JiraReferenceRule using the unified configuration.

// Result returns a concise rule message.
func (j JiraReferenceRule) Result(errors []errors.ValidationError) string {
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
func (j JiraReferenceRule) VerboseResult(errors []errors.ValidationError) string {
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
func (j JiraReferenceRule) Help(errors []errors.ValidationError) string {
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
		// Create error context with rich information
		errorCtx := appErrors.NewContext()

		helpMessage := `Empty Commit Subject Error: Cannot validate Jira references.

Your commit message has an empty subject line, so Jira reference validation cannot be performed.

✅ CORRECT FORMAT:
- A commit message should start with a subject line:
  "feat: add login feature PROJ-123"
  
  This is a descriptive body that explains the change in detail.
  It can span multiple lines.

❌ INCORRECT FORMAT:
- Your commit has an empty subject line

WHY THIS MATTERS:
- Jira references are critical for tracking work
- They provide traceability between code changes and issue tracking
- Without a proper reference, it's difficult to connect commits to tasks

NEXT STEPS:
1. Add a meaningful subject line to your commit with a valid Jira reference
   - Use 'git commit --amend' to edit your most recent commit
   - Follow your project's commit message conventions
   
2. If using conventional commits, place the Jira key at the end:
   feat(scope): descriptive message PROJ-123`

		// Create the error with rich context
		validationErr := appErrors.CreateRichError(
			updatedRule.Name(),
			appErrors.ErrEmptyMessage,
			"Commit subject is empty",
			helpMessage,
			errorCtx,
		)

		// Add error and return
		updatedRule.BaseRule = updatedRule.BaseRule.WithError(validationErr)

		return []appErrors.ValidationError{validationErr}, updatedRule
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
func (j JiraReferenceRule) Validate(_ context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
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
// NOTE: This function is only kept for future reference in case we need similar functionality.
// It is currently unused as we've migrated to more specific error creation functions.
//
// func (j JiraReferenceRule) createError(code appErrors.ValidationErrorCode, message string, context map[string]string) appErrors.ValidationError {
// 	// Create a new error context
// 	errorCtx := appErrors.NewContext()
//
// 	// Store the message in context for later reference
// 	if context == nil {
// 		context = make(map[string]string)
// 	}
//
// 	context["message"] = message
//
// 	// Create a help message based on the error code
// 	helpMessage := getJiraHelpMessage(code, context, j.validateBodyRef, j.validProjects)
//
// 	// Create the validation error with rich context
// 	validationErr := appErrors.CreateRichError(
// 		j.Name(),
// 		code,
// 		message,
// 		helpMessage,
// 		errorCtx,
// 	)
//
// 	// Add detailed context information
// 	for k, v := range context {
// 		validationErr = validationErr.WithContext(k, v)
// 	}
//
// 	return validationErr
// }
//
// // getJiraHelpMessage returns a detailed help message for different error codes.
// // NOTE: This function is only kept for future reference in case we need similar functionality.
// // It is currently unused as we've migrated to more specific error creation functions.
//
// func getJiraHelpMessage(code appErrors.ValidationErrorCode, context map[string]string, isBodyRef bool, validProjects []string) string {
// 	// Get the original error message if present
// 	errorMsg := ""
//
// 	if context != nil {
// 		if msg, ok := context["message"]; ok {
// 			errorMsg = msg
// 		}
// 	}
//
// 	// Only handle error codes that are actually used in this rule
// 	// Use if statements instead of switch to avoid exhaustive linter complaints
// 	if code == appErrors.ErrMissingJira {
// 		if isBodyRef {
// 			return `Missing Jira Reference Error: No Jira issue key found in commit body.
//
// Your commit message is missing the required Jira issue reference in the body.
//
// ✅ CORRECT FORMAT:
// - Include a "Refs:" line in your commit body with one or more Jira keys:
//
//   feat: add login feature
//
//   Implements the login screen with OAuth integration.
//
//   Refs: PROJ-123
//
// - For multiple tickets, separate them with commas:
//
//   Refs: PROJ-123, PROJ-456
//
// ❌ INCORRECT FORMAT:
// - Your commit body is missing the "Refs:" line
// - Informal references like "Works on PROJ-123" aren't properly recognized
//
// WHY THIS MATTERS:
// - "Refs:" lines provide a consistent format for tooling
// - Properly formatted references ensure traceability
// - They connect code changes to specific work items or issues
//
// NEXT STEPS:
// 1. Add a "Refs:" line before any sign-off lines:
//    - Use 'git commit --amend' to edit your most recent commit
//    - Format: Refs: PROJ-123 or Refs: PROJ-123, PROJ-456
//    - Place it at the end of your commit body, before any sign-off lines`
// 		}
//
// 		// Subject validation error
// 		return `Missing Jira Reference Error: No Jira issue key found in commit subject.
//
// Your commit message is missing the required Jira issue reference in the subject line.
//
// ✅ CORRECT FORMAT:
// - Include a Jira key in the subject line:
//   "Add new feature PROJ-123"
//
// - For conventional commits, place it at the end:
//   "feat: add login feature PROJ-123"
//
// - Common formats include:
//   "feat: implement user authentication PROJ-123"
//   "fix: resolve login timeout [PROJ-123]"
//   "docs: update README (PROJ-123)"
//
// ❌ INCORRECT FORMAT:
// - Your commit subject has no Jira issue key
// - Jira keys should follow the pattern PROJECT-123
//
// WHY THIS MATTERS:
// - Jira references connect code changes to specific work items
// - They provide traceability for future reference
// - They help automatically update Jira issues with commit information
//
// NEXT STEPS:
// 1. Add a Jira issue key to your commit subject:
//    - Use 'git commit --amend' to edit your most recent commit
//    - Add the Jira key at the end of your commit subject
//    - Ensure it follows the format PROJECT-123`
// 	} else if code == appErrors.ErrInvalidFormat {
// 		key := ""
//
// 		if context != nil {
// 			if k, ok := context["key"]; ok {
// 				key = k
// 			}
// 		}
//
// 		// Check for specific format errors
// 		if key != "" {
// 			return fmt.Sprintf(`Invalid Jira Key Format Error: "%s" is not a valid Jira key.
//
// Your commit includes a string that looks like a Jira reference but doesn't follow the correct format.
//
// ✅ CORRECT FORMAT:
// - Jira keys must follow the pattern PROJECT-123
// - Project keys are all uppercase letters: PROJ, TEAM, etc.
// - Issue numbers are separated by a hyphen: PROJ-123
//
// ❌ INCORRECT FORMAT:
// - "%s" doesn't match the required format
// - Common mistakes include:
//   - Missing hyphen between project and number (PROJ123)
//   - Lowercase project key (proj-123)
//   - Spaces in the key (PROJ 123)
//
// WHY THIS MATTERS:
// - Correctly formatted keys ensure proper issue linking
// - Automated tools rely on the standard format
// - It maintains consistency across the project
//
// NEXT STEPS:
// 1. Fix the Jira reference format:
//    - Use 'git commit --amend' to edit your most recent commit
//    - Ensure your Jira key follows PROJECT-123 format
//    - Make sure the hyphen is included between the project and issue number`, key, key)
// 		}
//
// 		if errorMsg != "" && strings.Contains(errorMsg, "must be at the end") {
// 			return `Invalid Jira Key Position Error: Jira key must be at the end of the subject.
//
// In conventional commit format, the Jira key must appear at the end of the subject line.
//
// ✅ CORRECT FORMAT:
// - Place the Jira key at the end of the subject:
//   "feat: add login feature PROJ-123"
//   "fix: resolve timeout issue [PROJ-123]"
//   "docs: update installation instructions (PROJ-123)"
//
// ❌ INCORRECT FORMAT:
// - Jira key in the middle of the subject:
//   "feat: PROJ-123 add login feature"
//   "PROJ-123: fix timeout issue"
//
// WHY THIS MATTERS:
// - Consistent formatting makes commit logs easier to read
// - It separates the descriptive content from the reference
// - Automated parsing tools expect this standard format
//
// NEXT STEPS:
// 1. Move the Jira key to the end of your subject line:
//    - Use 'git commit --amend' to edit your most recent commit
//    - The key can be directly at the end or in brackets/parentheses`
// 		}
//
// 		if errorMsg != "" && strings.Contains(errorMsg, "Refs:") {
// 			line := ""
//
// 			if context != nil {
// 				if l, ok := context["line"]; ok {
// 					line = l
// 				}
// 			}
//
// 			return fmt.Sprintf(`Invalid Refs Line Format Error: "%s" is not correctly formatted.
//
// The "Refs:" line in your commit body doesn't follow the expected format.
//
// ✅ CORRECT FORMAT:
// - The Refs line should start with "Refs:" followed by one or more Jira keys:
//   "Refs: PROJ-123"
//
// - For multiple tickets, separate them with commas:
//   "Refs: PROJ-123, TEAM-456"
//
// - The Refs line should appear on its own line in the commit body
//
// ❌ INCORRECT FORMAT:
// - Your current format: "%s"
// - Common mistakes include:
//   - Incorrect spacing
//   - Missing or malformed Jira keys
//   - Using other separators than commas
//
// WHY THIS MATTERS:
// - The "Refs:" format is specifically designed for tooling integration
// - It provides a consistent way to link commits to issues
// - Automated systems parse this format to update issue trackers
//
// NEXT STEPS:
// 1. Fix the Refs line in your commit message:
//    - Use 'git commit --amend' to edit your most recent commit
//    - Format properly: "Refs: PROJ-123" or "Refs: PROJ-123, PROJ-456"
//    - Ensure proper spacing and comma separation for multiple keys`, line, line)
// 		}
//
// 		if errorMsg != "" && strings.Contains(errorMsg, "must appear before") {
// 			return `Invalid Line Order Error: "Refs:" line must appear before sign-off lines.
//
// Your commit message has the "Refs:" line after a "Signed-off-by:" line, which is incorrect.
//
// ✅ CORRECT FORMAT:
// - The correct order for your commit message is:
//   1. Commit subject
//   2. Blank line
//   3. Commit body (if any)
//   4. Refs: line(s)
//   5. Signed-off-by line(s)
//
// Example:
//   feat: add login feature
//
//   Implements OAuth-based authentication flow.
//
//   Refs: PROJ-123
//   Signed-off-by: Dev Name <dev@example.com>
//
// ❌ INCORRECT FORMAT:
// - Your commit has the Refs line after the Signed-off-by line
//
// WHY THIS MATTERS:
// - Consistent ordering helps automated tools parse commit messages
// - It groups related information together: body, references, signatures
// - It's a standard practice in commit message formatting
//
// NEXT STEPS:
// 1. Reorder the lines in your commit message:
//    - Use 'git commit --amend' to edit your most recent commit
//    - Move the "Refs:" line above any "Signed-off-by:" lines`
// 		}
//
// 		// General invalid format fallback
// 		return `Invalid Jira Reference Format Error: The Jira reference format is incorrect.
//
// Your commit's Jira reference doesn't follow the expected format requirements.
//
// ✅ CORRECT FORMAT:
// - For subject references:
//   "Add new feature PROJ-123"
//   "feat: add login feature PROJ-123"
//
// - For body references:
//   "Refs: PROJ-123"
//   "Refs: PROJ-123, TEAM-456"
//
// ❌ INCORRECT FORMAT:
// - Common issues include:
//   - Missing hyphen between project and number (PROJ123)
//   - Incorrect positioning in conventional commits
//   - Incorrectly formatted "Refs:" lines
//   - References in non-standard locations
//
// WHY THIS MATTERS:
// - Standardized formats ensure tooling integration works correctly
// - They make commit logs consistent and readable
// - They enable proper issue tracking and linking
//
// NEXT STEPS:
// 1. Review your commit message format:
//    - Use 'git commit --amend' to edit your most recent commit
//    - Follow the examples above for correct formatting
//    - Ensure Jira keys follow the PROJECT-123 pattern`
// 	} else if code == appErrors.ErrInvalidType {
// 		project := ""
//
// 		if context != nil {
// 			if p, ok := context["project"]; ok {
// 				project = p
// 			}
// 		}
//
// 		validProjectsStr := ""
// 		if len(validProjects) > 0 {
// 			validProjectsStr = "\n\nValid projects: " + strings.Join(validProjects, ", ")
// 		}
//
// 		return fmt.Sprintf(`Invalid Jira Project Error: Project "%s" is not recognized.
//
// Your commit references a Jira project that is not in the list of valid projects for this repository.%s
//
// ✅ CORRECT FORMAT:
// - Use one of the approved project keys in your Jira references
// - Project keys are the prefix before the hyphen: PROJ-123
//
// ❌ INCORRECT FORMAT:
// - "%s" is not a recognized project key
// - Your commit should reference a valid project
//
// WHY THIS MATTERS:
// - Project restrictions ensure references point to relevant Jira projects
// - It prevents typos and references to unrelated projects
// - It maintains consistency across the repository
//
// NEXT STEPS:
// 1. Update your commit with a valid project reference:
//    - Use 'git commit --amend' to edit your most recent commit
//    - Replace "%s" with one of the valid project keys
//    - Keep the same issue number if applicable`, project, validProjectsStr, project, project)
// 	} else if code == appErrors.ErrEmptyMessage {
// 		// This is handled separately in validateJiraWithState
// 		return `Empty Commit Error: Cannot validate Jira references with empty commit.`
// 	}
//
// 	// Default generic help message
// 	return `Jira Reference Error: The commit message doesn't include a valid Jira reference.
//
// Your commit message needs a properly formatted Jira issue reference.
//
// ✅ CORRECT FORMAT:
// - For subject references:
//   "Add new feature PROJ-123"
//   "feat: add login feature PROJ-123"
//
// - For body references (if enabled):
//   "Refs: PROJ-123"
//   "Refs: PROJ-123, TEAM-456"
//
// ❌ INCORRECT FORMAT:
// - Missing Jira references
// - Incorrectly formatted keys
// - Keys in wrong locations
//
// WHY THIS MATTERS:
// - Jira references connect code changes to specific work items
// - They provide critical traceability
// - They help track progress and requirements
//
// NEXT STEPS:
// 1. Add a properly formatted Jira issue key to your commit:
//    - Use 'git commit --amend' to edit your most recent commit
//    - Follow your team's conventions for Jira reference placement
//    - Ensure the key follows the PROJECT-123 format`
// }

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
			invalidKey := invalidMatches[0]

			// Extract project and issue parts when possible
			projectPart := regexp.MustCompile(`([A-Z]+)`).FindString(invalidKey)
			numberPart := regexp.MustCompile(`(\d+)`).FindString(invalidKey)
			suggestedKey := projectPart + "-" + numberPart

			helpText := fmt.Sprintf(`Invalid Jira Key Format Error: "%s" is not a valid Jira key.

✅ CORRECT FORMAT:
- Jira keys must follow the pattern PROJECT-123
- Project keys are all uppercase letters: PROJ, TEAM, etc.
- Issue numbers are separated by a hyphen: PROJ-123

❌ INCORRECT FORMAT:
- "%s" doesn't match the required format (missing hyphen)
- The correct format would be: %s

WHY THIS MATTERS:
- Correctly formatted keys ensure proper issue linking
- Automated tools rely on the standard format
- It maintains consistency across the project

NEXT STEPS:
1. Fix the Jira reference format:
   - Use 'git commit --amend' to edit your most recent commit
   - Ensure your Jira key follows PROJECT-123 format
   - For conventional commits, make sure it's at the end of the line`,
				invalidKey, invalidKey, suggestedKey)

			err := appErrors.JiraError(
				j.Name(),
				appErrors.ErrInvalidFormat,
				"invalid Jira issue key format: "+invalidKey+" (should be PROJECT-123)",
				helpText,
				invalidKey,
				firstLine,
				map[string]string{"suggested_key": suggestedKey},
			)
			errors = append(errors, err)

			return errors
		}

		// No key-like patterns found, report as missing
		helpText := `Missing Jira Reference Error: No Jira issue key found in commit subject.

Your commit message is missing the required Jira issue reference in the subject line.

✅ CORRECT FORMAT:
- Include a Jira key in the subject line:
  "Add new feature PROJ-123"
  
- For conventional commits, place it at the end:
  "feat: add login feature PROJ-123"
  
- Common formats include:
  "feat: implement user authentication PROJ-123"
  "fix: resolve login timeout [PROJ-123]"
  "docs: update README (PROJ-123)"

❌ INCORRECT FORMAT:
- Your commit subject has no Jira issue key
- Jira keys should follow the pattern PROJECT-123

WHY THIS MATTERS:
- Jira references connect code changes to specific work items
- They provide traceability for future reference
- They help automatically update Jira issues with commit information

NEXT STEPS:
1. Add a Jira issue key to your commit subject:
   - Use 'git commit --amend' to edit your most recent commit
   - Add the Jira key at the end of your commit subject
   - Ensure it follows the format PROJECT-123`

		additionalContext := map[string]string{
			"subject":         firstLine,
			"is_conventional": "true",
		}

		err := appErrors.JiraError(
			j.Name(),
			appErrors.ErrMissingJira,
			"no Jira issue key found in the commit subject",
			helpText,
			"", // No key to provide
			firstLine,
			additionalContext,
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
	helpText := fmt.Sprintf(`Invalid Jira Key Position Error: Jira key must be at the end of the subject.

✅ CORRECT FORMAT:
- Place the Jira key at the end of the subject:
  "feat: add login feature %s"
  "fix: resolve timeout issue [%s]"
  "docs: update installation instructions (%s)"

❌ INCORRECT FORMAT:
- Your current format has the key in the wrong position:
  "%s"
- Jira key in the middle of the subject:
  "feat: %s add login feature"
  "%s: fix timeout issue" 

WHY THIS MATTERS:
- Consistent formatting makes commit logs easier to read
- It separates the descriptive content from the reference
- Automated parsing tools expect this standard format

NEXT STEPS:
1. Move the Jira key to the end of your subject line:
   - Use 'git commit --amend' to edit your most recent commit
   - The key can be directly at the end or in brackets/parentheses
   - Sample correct format: "feat: add feature %s"`,
		lastMatch, lastMatch, lastMatch, firstLine, lastMatch, lastMatch, lastMatch)

	err := appErrors.JiraError(
		j.Name(),
		appErrors.ErrInvalidFormat,
		"in conventional commit format, Jira issue key must be at the end of the first line",
		helpText,
		lastMatch,
		firstLine,
		map[string]string{
			"is_conventional": "true",
			"allowed_formats": strings.Join(validSuffixes, ", "),
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
			invalidKey := invalidMatches[0]

			// Extract project and issue parts when possible
			projectPart := regexp.MustCompile(`([A-Z]+)`).FindString(invalidKey)
			numberPart := regexp.MustCompile(`(\d+)`).FindString(invalidKey)
			suggestedKey := projectPart + "-" + numberPart

			helpText := fmt.Sprintf(`Invalid Jira Key Format Error: "%s" is not a valid Jira key.

✅ CORRECT FORMAT:
- Jira keys must follow the pattern PROJECT-123
- Project keys are all uppercase letters: PROJ, TEAM, etc.
- Issue numbers are separated by a hyphen: PROJ-123

❌ INCORRECT FORMAT:
- "%s" doesn't match the required format (missing hyphen)
- The correct format would be: %s

WHY THIS MATTERS:
- Correctly formatted keys ensure proper issue linking
- Automated tools rely on the standard format
- It maintains consistency across the project

NEXT STEPS:
1. Fix the Jira reference format:
   - Use 'git commit --amend' to edit your most recent commit
   - Ensure your Jira key follows PROJECT-123 format
   - Make sure the hyphen is included between the project and issue number`,
				invalidKey, invalidKey, suggestedKey)

			err := appErrors.JiraError(
				j.Name(),
				appErrors.ErrInvalidFormat,
				"invalid Jira issue key format: "+invalidKey+" (should be PROJECT-123)",
				helpText,
				invalidKey,
				subject,
				map[string]string{"suggested_key": suggestedKey},
			)
			errors = append(errors, err)

			return errors
		}

		// No key-like patterns found, report as missing
		helpText := `Missing Jira Reference Error: No Jira issue key found in commit subject.

Your commit message is missing the required Jira issue reference in the subject line.

✅ CORRECT FORMAT:
- Include a Jira key in the subject line:
  "Add new feature PROJ-123"
  
- Common formats include:
  "Implement user authentication PROJ-123"
  "Resolve login timeout issue [PROJ-123]"
  "Update documentation (PROJ-123)"

❌ INCORRECT FORMAT:
- Your commit subject has no Jira issue key
- Jira keys should follow the pattern PROJECT-123

WHY THIS MATTERS:
- Jira references connect code changes to specific work items
- They provide traceability for future reference
- They help automatically update Jira issues with commit information

NEXT STEPS:
1. Add a Jira issue key to your commit subject:
   - Use 'git commit --amend' to edit your most recent commit
   - Add the Jira key anywhere in your commit subject
   - Ensure it follows the format PROJECT-123`

		additionalContext := map[string]string{
			"subject": subject,
		}

		err := appErrors.JiraError(
			j.Name(),
			appErrors.ErrMissingJira,
			"no Jira issue key found in the commit subject",
			helpText,
			"", // No key to provide
			subject,
			additionalContext,
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
		helpText := `Missing Commit Body Error: No Jira issue key found in commit body

✅ CORRECT FORMAT:
- Your commit message should have a body section with a Refs line:
  
  feat: add login feature
  
  This is the commit body describing the changes in detail.
  
  Refs: PROJ-123

❌ INCORRECT FORMAT:
- Your commit has no body section
- Body references require a non-empty commit body with a Refs line

WHY THIS MATTERS:
- The commit body provides important context for changes
- "Refs:" lines in the body standardize Jira references
- Properly structured commits improve project traceability

NEXT STEPS:
1. Update your commit with a proper body section:
   - Use 'git commit --amend' to edit your commit message
   - Add a blank line after the subject
   - Add descriptive text about the change
   - Include a "Refs: PROJ-123" line with the appropriate Jira key`

		err := appErrors.ReferenceError(
			j.Name(),
			"no Jira issue key found in the commit body",
			helpText,
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

				helpText := fmt.Sprintf(`Invalid Jira Key Format in Refs Line: "%s" is not a valid Jira key

✅ CORRECT FORMAT: Refs: PROJECT-123 or Refs: PROJECT-123, ABC-456
❌ INCORRECT FORMAT: Refs: %s (missing hyphen between project code and issue number)

WHY THIS MATTERS:
Jira uses a specific format for issue keys consisting of a project code in uppercase 
letters followed by a hyphen and a number. Without the correct format, automated 
tools cannot link your commit to the proper Jira issue.

NEXT STEPS:
1. Update your commit message with the correct Jira key format
2. Use "git commit --amend" to modify your commit message
3. Add a hyphen between the project code and issue number
4. Ensure the format follows: Refs: PROJECT-123`, invalidKey, invalidKey)

				err := appErrors.FormatError(
					j.Name(),
					"invalid Jira issue key format: "+invalidKey+" (should be PROJECT-123)",
					helpText,
					line,
				)
				errors = append(errors, err)

				return info, errors
			}

			// Regular invalid format
			helpText := fmt.Sprintf(`Invalid Refs Line Format: "%s" is not correctly formatted

✅ CORRECT FORMAT:
- Refs: PROJECT-123
- Refs: PROJECT-123, PROJECT-456, PROJECT-789

❌ INCORRECT FORMAT:
- Your current format: "%s"
- The line must start with "Refs:" followed by a properly formatted Jira key

WHY THIS MATTERS:
- The "Refs:" format is designed for tooling integration
- Consistent formatting ensures traceability between commits and issues
- It enables automated tools to correctly link commits to Jira issues

NEXT STEPS:
1. Fix the format of your Refs line:
   - Use 'git commit --amend' to edit your commit message
   - Format the line as "Refs: PROJECT-123" (with a hyphen in the Jira key)
   - For multiple references, use "Refs: PROJECT-123, PROJECT-456"`, line, line)

			err := appErrors.FormatError(
				j.Name(),
				"invalid Refs format in commit body, should be 'Refs: PROJ-123' or 'Refs: PROJ-123, PROJ-456'",
				helpText,
				line,
			)
			errors = append(errors, err)

			return info, errors
		}
	}

	// Validate that Refs: exists
	if !info.foundRefs {
		helpText := `Missing Jira Reference Error: No Jira issue key found in commit body with Refs: prefix

✅ CORRECT FORMAT:
- Your commit body must include a line starting with "Refs:" followed by Jira keys:
  Refs: PROJECT-123
  
- For multiple references:
  Refs: PROJECT-123, PROJECT-456, PROJECT-789

❌ INCORRECT FORMAT:
- No "Refs:" line found in your commit body
- Informal references (like "Related to PROJECT-123") are not recognized

WHY THIS MATTERS:
- "Refs:" lines provide a consistent format for tooling
- They connect code changes to specific Jira issues
- They enable automated tools to track development progress

NEXT STEPS:
1. Edit your commit message to add a properly formatted "Refs:" line
   - Use 'git commit --amend' to modify your commit
   - Add a line with "Refs: PROJECT-123" (with the appropriate Jira key)
   - Place this before any signature lines`

		err := appErrors.ReferenceError(
			j.Name(),
			"no Jira issue key found in the commit body with Refs: prefix",
			helpText,
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
		helpText := `Invalid Line Order Error: "Refs:" line must appear before sign-off lines

✅ CORRECT ORDER:
1. Commit subject
2. Blank line
3. Commit body (if any)
4. Refs: line(s)
5. Signed-off-by line(s)

Example:
feat: add login feature

Implements OAuth-based authentication flow.

Refs: PROJ-123
Signed-off-by: Dev Name <dev@example.com>

❌ INCORRECT ORDER:
- Your commit has the Refs line after the Signed-off-by line

WHY THIS MATTERS:
- Consistent ordering helps automated tools parse commit messages
- It groups related information together logically
- It's a standard practice in commit message formatting

NEXT STEPS:
1. Reorder the lines in your commit message:
   - Use 'git commit --amend' to edit your most recent commit
   - Move the "Refs:" line above any "Signed-off-by:" lines`

		context := map[string]string{
			"refs_line":    strconv.Itoa(info.refsLineNum),
			"signoff_line": strconv.Itoa(info.signOffLineNum),
		}

		err := appErrors.FormatError(
			j.Name(),
			"Refs: line must appear before any Signed-off-by lines",
			helpText,
			"", // No subject for this error
		)

		// Add context information
		for k, v := range context {
			err = err.WithContext(k, v)
		}

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
		helpText := fmt.Sprintf(`Invalid Jira Key Format Error: "%s" is not a valid Jira key.

✅ CORRECT FORMAT:
- Jira keys must follow the pattern PROJECT-123
- Project keys are all uppercase letters: PROJ, TEAM, etc.
- Issue numbers are separated by a hyphen: PROJ-123

❌ INCORRECT FORMAT:
- "%s" doesn't match the required format
- Common mistakes include:
  - Missing hyphen between project and number (PROJ123)
  - Lowercase project key (proj-123)
  - Spaces in the key (PROJ 123)

WHY THIS MATTERS:
- Correctly formatted keys ensure proper issue linking
- Automated tools rely on the standard format
- It maintains consistency across the project

NEXT STEPS:
1. Fix the Jira reference format:
   - Use 'git commit --amend' to edit your most recent commit
   - Ensure your Jira key follows PROJECT-123 format
   - Make sure the hyphen is included between the project and issue number`, jiraKey, jiraKey)

		err := appErrors.JiraError(
			j.Name(),
			appErrors.ErrInvalidFormat,
			"invalid Jira issue key format: "+jiraKey+" (should be PROJECT-123)",
			helpText,
			jiraKey,
			"",
			nil,
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
		validProjectsList := strings.Join(j.validProjects, ", ")

		// Create a different message format depending on whether we have valid projects to recommend
		var helpText string
		if len(j.validProjects) > 0 {
			helpText = fmt.Sprintf(`Invalid Jira Project Error: "%s" is not in the allowed project list.

✅ CORRECT PROJECTS: %s

❌ INCORRECT PROJECT: %s (not in the allowed list)

WHY THIS MATTERS:
- Using standardized project keys ensures consistency across the repository
- It prevents typos and references to unrelated projects
- It maintains proper traceability between code and issue tracking

NEXT STEPS:
1. Update your commit message with an allowed project key:
   - Use 'git commit --amend' to edit your most recent commit
   - Replace "%s" with one of the allowed project keys
   - Keep the same issue number if applicable: %s-123`,
				projectKey, validProjectsList, projectKey, projectKey, j.validProjects[0])
		} else {
			helpText = fmt.Sprintf(`Invalid Jira Project Error: "%s" is not a valid project key.

✅ CORRECT FORMAT: Standard uppercase Jira project key (e.g., PROJ, TEAM)

❌ INCORRECT PROJECT: %s

WHY THIS MATTERS:
- Using standardized project keys ensures consistency across the repository
- It prevents typos and references to unrelated projects
- It maintains proper traceability between code and issue tracking

NEXT STEPS:
1. Update your commit message with the correct project key:
   - Use 'git commit --amend' to edit your most recent commit
   - Replace "%s" with the appropriate project key for your issue
   - Consult your team's documentation for the list of valid project keys`,
				projectKey, projectKey, projectKey)
		}

		// Additional context to be included
		additionalContext := map[string]string{
			"project":        projectKey,
			"valid_projects": validProjectsList,
		}

		err := appErrors.JiraError(
			j.Name(),
			appErrors.ErrInvalidType,
			"Jira project "+projectKey+" is not a valid project",
			helpText,
			jiraKey,
			"", // No subject needed here
			additionalContext,
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
