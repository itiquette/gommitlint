// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/application/report"
	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/infrastructure/git"
	"github.com/itiquette/gommitlint/internal/infrastructure/output"
	"github.com/spf13/cobra"
)

// newValidateCmd creates a new validate command.
func newValidateCmd() *cobra.Command {
	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validates commit messages",
		Long: `Validates commit message/s against a set of rules.

Examples:
  # Validate commits in the current branch against main
  gommitlint validate --base-branch=main
  
  # Validate a specific commit
  gommitlint validate --git-reference=HEAD
  
  # Validate a commit message from a file
  gommitlint validate --message-file=/path/to/commit-msg.txt
  
  # Validate a range of commits
  gommitlint validate --revision-range=main..HEAD`,
		Run: func(cmd *cobra.Command, _ []string) {
			// Process validation request
			exitCode, err := runNewValidation(cmd)
			if err != nil {
				cmd.PrintErrf("Error: %v\n", err)
				os.Exit(1)
			}

			// Special exit codes:
			// 0 - success
			// 1 - system error
			// 2 - validation failure
			if exitCode > 0 {
				os.Exit(exitCode)
			}
		},
	}

	// Add flags to the command
	validateCmd.Flags().String("message-file", "", "commit message file path to validate")
	validateCmd.Flags().String("git-reference", "", "git reference to validate (defaults to HEAD)")
	validateCmd.Flags().Int("commit-count", 1, "number of commits from HEAD to validate")
	validateCmd.Flags().String("revision-range", "", "range of commits to validate (<commit1>..<commit2>)")
	validateCmd.Flags().String("base-branch", "", "base branch to compare with (sets revision-range to <base-branch>..HEAD)")
	validateCmd.Flags().BoolP("verbose", "v", false, "show detailed validation results")
	validateCmd.Flags().Bool("extra-verbose", false, "show extra detailed validation results")
	validateCmd.Flags().Bool("light-mode", false, "use light background color scheme")
	validateCmd.Flags().String("rulehelp", "", "show detailed help for a specific rule")
	validateCmd.Flags().String("format", "text", "output format: text, json, github, or gitlab")
	validateCmd.Flags().Bool("skip-merge-commits", true, "skip merge commits in validation")
	validateCmd.Flags().String("repo-path", "", "path to the repository to validate")

	return validateCmd
}

// We'll use a functional approach instead of storing context in a struct
// to avoid the containedctx linter error

type ValidationParameters struct {
	// No context is stored directly in the struct - we'll pass it as a parameter
	OutWriter io.Writer

	// Validation options
	MessageFile      string
	GitReference     string
	CommitCount      int
	RevisionRange    string
	BaseBranch       string
	SkipMergeCommits bool
	RepoPath         string

	// Reporting options
	Verbose      bool
	ExtraVerbose bool
	LightMode    bool
	RuleHelp     string
	Format       string

	// Dependencies
	Dependencies *AppDependencies
}

// NewValidationParameters creates ValidationParameters from a cobra command.
func NewValidationParameters(cmd *cobra.Command) ValidationParameters {
	// Extract all parameters from the command
	messageFile, _ := cmd.Flags().GetString("message-file")
	gitReference, _ := cmd.Flags().GetString("git-reference")
	commitCount, _ := cmd.Flags().GetInt("commit-count")
	revisionRange, _ := cmd.Flags().GetString("revision-range")
	baseBranch, _ := cmd.Flags().GetString("base-branch")
	verbose, _ := cmd.Flags().GetBool("verbose")
	extraVerbose, _ := cmd.Flags().GetBool("extra-verbose")
	lightMode, _ := cmd.Flags().GetBool("light-mode")
	ruleHelp, _ := cmd.Flags().GetString("rulehelp")
	format, _ := cmd.Flags().GetString("format")
	skipMergeCommits, _ := cmd.Flags().GetBool("skip-merge-commits")
	repoPath, _ := cmd.Flags().GetString("repo-path")

	// Get dependencies from context if available
	var deps *AppDependencies

	if cmd.Context() != nil {
		if d, ok := cmd.Context().Value(dependenciesKey).(*AppDependencies); ok {
			deps = d
		}
	}

	return ValidationParameters{
		OutWriter:        cmd.OutOrStdout(),
		MessageFile:      messageFile,
		GitReference:     gitReference,
		CommitCount:      commitCount,
		RevisionRange:    revisionRange,
		BaseBranch:       baseBranch,
		Verbose:          verbose,
		ExtraVerbose:     extraVerbose,
		LightMode:        lightMode,
		RuleHelp:         ruleHelp,
		Format:           format,
		SkipMergeCommits: skipMergeCommits,
		RepoPath:         repoPath,
		Dependencies:     deps,
	}
}

// Maximum allowed values for CLI parameters.
const (
	MaxCommitCount = 1000 // Maximum reasonable number of commits to validate
	MaxPathLength  = 4096 // Maximum path length
	MaxRefLength   = 256  // Maximum Git reference length
)

// validateParameterLength checks if a parameter value exceeds the maximum safe length.
func validateParameterLength(name, value string, maxLength int) error {
	if len(value) > maxLength {
		return fmt.Errorf("%s exceeds maximum allowed length (%d)", name, maxLength)
	}

	return nil
}

// validateGitReference checks if a Git reference value is safe.
func validateGitReference(ref string) error {
	if ref == "" {
		return nil
	}

	// Check length
	if err := validateParameterLength("Git reference", ref, MaxRefLength); err != nil {
		return err
	}

	// Validate git reference format - allowing common git reference formats
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_\-./~^{}[\]@]+$`)
	if !validPattern.MatchString(ref) {
		return errors.New("git reference contains invalid characters")
	}

	return nil
}

// validateFilePath checks if a file path is safe.
func validateFilePath(path string) error {
	if path == "" {
		return nil
	}

	// Check length
	if err := validateParameterLength("File path", path, MaxPathLength); err != nil {
		return err
	}

	// Validate basic path safety
	// This doesn't check if the path exists, just if it's a valid path format
	path = filepath.Clean(path)
	if strings.Contains(path, "..") {
		return errors.New("path traversal detected in file path")
	}

	return nil
}

// validateCommitCount checks if a commit count is reasonable.
func validateCommitCount(count int) error {
	if count < 0 {
		return errors.New("commit count cannot be negative")
	}

	if count > MaxCommitCount {
		return fmt.Errorf("commit count exceeds maximum allowed value (%d)", MaxCommitCount)
	}

	return nil
}

// ToValidationOptions converts ValidationParameters to validate.ValidationOptions.
func (p ValidationParameters) ToValidationOptions() (validate.ValidationOptions, error) {
	// Validate all input parameters for security
	if err := validateFilePath(p.MessageFile); err != nil {
		return validate.ValidationOptions{}, fmt.Errorf("invalid message file: %w", err)
	}

	if err := validateGitReference(p.GitReference); err != nil {
		return validate.ValidationOptions{}, fmt.Errorf("invalid git reference: %w", err)
	}

	if err := validateGitReference(p.BaseBranch); err != nil {
		return validate.ValidationOptions{}, fmt.Errorf("invalid base branch: %w", err)
	}

	if err := validateCommitCount(p.CommitCount); err != nil {
		return validate.ValidationOptions{}, fmt.Errorf("invalid commit count: %w", err)
	}

	// Process validation flags with precedence
	opts := validate.ValidationOptions{
		SkipMergeCommits: p.SkipMergeCommits,
	}

	// Default to validating HEAD if no other option is provided
	if p.GitReference == "" && p.RevisionRange == "" && p.BaseBranch == "" && p.MessageFile == "" && p.CommitCount <= 0 {
		// Default behavior uses HEAD
		opts.CommitHash = "HEAD"
	}

	// Apply validation source with precedence order
	if p.MessageFile != "" {
		// 1. Message from file (highest priority)
		opts.MessageFile = filepath.Clean(p.MessageFile)
	} else if p.BaseBranch != "" {
		// 2. Base branch comparison
		opts.FromHash = p.BaseBranch
		opts.ToHash = "HEAD"
	} else if p.RevisionRange != "" {
		// 3. Revision range
		// Validate revision range
		if err := validateParameterLength("Revision range", p.RevisionRange, MaxRefLength); err != nil {
			return validate.ValidationOptions{}, err
		}

		// Parse revision range (format: from..to)
		parts := parseRevisionRange(p.RevisionRange)
		if len(parts) == 2 {
			// Validate both parts
			if err := validateGitReference(parts[0]); err != nil {
				return validate.ValidationOptions{}, fmt.Errorf("invalid revision range start: %w", err)
			}

			if err := validateGitReference(parts[1]); err != nil {
				return validate.ValidationOptions{}, fmt.Errorf("invalid revision range end: %w", err)
			}

			opts.FromHash = parts[0]
			opts.ToHash = parts[1]
		} else {
			return validate.ValidationOptions{}, fmt.Errorf("invalid revision range format: %s (expected format: from..to)", p.RevisionRange)
		}
	} else if p.GitReference != "" {
		// 4. Single git reference
		opts.CommitHash = p.GitReference
	} else if p.CommitCount > 0 {
		// 5. Commit count
		opts.CommitCount = p.CommitCount
	}

	return opts, nil
}

// ToReportOptions converts ValidationParameters to report.Options.
func (p ValidationParameters) ToReportOptions() report.Options {
	return report.Options{
		Format:         getReportFormat(p.Format),
		Verbose:        p.Verbose,
		ShowHelp:       p.ExtraVerbose || p.RuleHelp != "",
		RuleToShowHelp: p.RuleHelp,
		LightMode:      p.LightMode,
		Writer:         p.OutWriter,
	}
}

// CreateFormatter creates a formatter based on the parameters.
func (p ValidationParameters) CreateFormatter() domain.ResultFormatter {
	reportOptions := p.ToReportOptions()

	switch reportOptions.Format {
	case report.FormatJSON:
		return output.NewJSONFormatter()
	case report.FormatGitHubActions:
		return output.NewGitHubFormatter().
			WithVerbose(reportOptions.Verbose).
			WithShowHelp(reportOptions.ShowHelp)
	case report.FormatGitLabCI:
		return output.NewGitLabFormatter().
			WithVerbose(reportOptions.Verbose).
			WithShowHelp(reportOptions.ShowHelp)
	case report.FormatText:
		fallthrough
	default:
		return output.NewTextFormatter().
			WithVerbose(reportOptions.Verbose).
			WithShowHelp(reportOptions.ShowHelp).
			WithLightMode(reportOptions.LightMode)
	}
}

// CreateValidationService creates a validation service based on parameters.
func (p ValidationParameters) CreateValidationService() (validate.ValidationService, error) {
	// Validate repository path
	if err := validateFilePath(p.RepoPath); err != nil {
		return validate.ValidationService{}, fmt.Errorf("invalid repository path: %w", err)
	}

	if p.Dependencies != nil {
		fmt.Println("AAAAA")
		// Create validation service with injected dependencies
		return createValidationServiceWithDeps(p.Dependencies, p.RepoPath)
	}

	fmt.Println("BBBBB")
	// Fall back to default service creation
	return validate.CreateDefaultValidationService(p.RepoPath)
}

// runNewValidation handles the core validation logic and returns an exit code.
func runNewValidation(cmd *cobra.Command) (int, error) {
	// Create parameters object to encapsulate all inputs
	params := NewValidationParameters(cmd)

	// Create validation service
	service, err := params.CreateValidationService()
	if err != nil {
		return 1, fmt.Errorf("failed to create validation service: %w", err)
	}

	// Convert parameters to validation options
	opts, err := params.ToValidationOptions()
	if err != nil {
		return 1, err
	}

	// Validate according to options - pass the context from the command
	ctx := cmd.Context()
	if ctx == nil {
		// Use background context as fallback for tests
		ctx = context.Background()
	}

	results, err := service.ValidateWithOptions(ctx, opts)
	if err != nil {
		return 1, fmt.Errorf("validation failed: %w", err)
	}

	// Create formatter and report generator
	formatter := params.CreateFormatter()
	reportOptions := params.ToReportOptions()
	reportGenerator := report.NewGenerator(reportOptions, formatter)

	// Generate report using the domain interface
	err = reportGenerator.GenerateReport(results)
	if err != nil {
		return 1, fmt.Errorf("failed to generate report: %w", err)
	}

	// Return success if all rules passed, validation failure code otherwise
	if results.AllPassed() {
		return 0, nil
	}

	return 2, nil
}

// parseRevisionRange parses a revision range string (e.g., "main..HEAD") into a slice of parts.
func parseRevisionRange(revisionRange string) []string {
	// Split by ".." to get from and to
	parts := []string{}

	// Handle both ".." and "..." formats
	if idx := indexOf(revisionRange, "..."); idx >= 0 {
		parts = append(parts, revisionRange[:idx], revisionRange[idx+3:])
	} else if idx := indexOf(revisionRange, ".."); idx >= 0 {
		parts = append(parts, revisionRange[:idx], revisionRange[idx+2:])
	}

	return parts
}

// indexOf returns the index of the first occurrence of substring in s, or -1 if not found.
func indexOf(s, substring string) int {
	for i := 0; i <= len(s)-len(substring); i++ {
		if s[i:i+len(substring)] == substring {
			return i
		}
	}

	return -1
}

// getReportFormat converts a string format to the report.Format type.
func getReportFormat(format string) report.Format {
	switch format {
	case "json":
		return report.FormatJSON
	case "github":
		return report.FormatGitHubActions
	case "gitlab":
		return report.FormatGitLabCI
	default:
		return report.FormatText
	}
}

// ConfigAdapter adapts Config to the ValidationConfig interface needed by the validation service.
type ConfigAdapter struct {
	config config.Config
}

// NewConfigAdapter creates a new adapter for Config.
func NewConfigAdapter(cfg config.Config) *ConfigAdapter {
	return &ConfigAdapter{
		config: cfg,
	}
}

// This adapter implements all the ValidationConfig methods by delegating to the underlying Config

// SubjectRequireImperative delegates to the underlying config.
func (a *ConfigAdapter) SubjectRequireImperative() bool {
	return a.config.SubjectRequireImperative()
}

// SubjectMaxLength delegates to the underlying config.
func (a *ConfigAdapter) SubjectMaxLength() int {
	return a.config.SubjectMaxLength()
}

// SubjectCase delegates to the underlying config.
func (a *ConfigAdapter) SubjectCase() string {
	return a.config.SubjectCase()
}

// SubjectInvalidSuffixes delegates to the underlying config.
func (a *ConfigAdapter) SubjectInvalidSuffixes() string {
	// The domain interface expects a string
	return a.config.SubjectInvalidSuffixes()
}

// BodyRequired delegates to the underlying config.
func (a *ConfigAdapter) BodyRequired() bool {
	return a.config.BodyRequired()
}

// BodyAllowSignOffOnly delegates to the underlying config.
func (a *ConfigAdapter) BodyAllowSignOffOnly() bool {
	return a.config.BodyAllowSignOffOnly()
}

// ConventionalRequired delegates to the underlying config.
func (a *ConfigAdapter) ConventionalRequired() bool {
	return a.config.ConventionalRequired()
}

// ConventionalTypes delegates to the underlying config.
func (a *ConfigAdapter) ConventionalTypes() []string {
	return a.config.ConventionalTypes()
}

// ConventionalScopes delegates to the underlying config.
func (a *ConfigAdapter) ConventionalScopes() []string {
	return a.config.ConventionalScopes()
}

// ConventionalMaxDescriptionLength delegates to the underlying config.
func (a *ConfigAdapter) ConventionalMaxDescriptionLength() int {
	return a.config.ConventionalMaxDescriptionLength()
}

// JiraRequired delegates to the underlying config.
func (a *ConfigAdapter) JiraRequired() bool {
	return a.config.JiraRequired()
}

// JiraProjects delegates to the underlying config.
func (a *ConfigAdapter) JiraProjects() []string {
	return a.config.JiraProjects()
}

// JiraPattern delegates to the underlying config.
func (a *ConfigAdapter) JiraPattern() string {
	return a.config.JiraPattern()
}

// JiraBodyRef delegates to the underlying config.
func (a *ConfigAdapter) JiraBodyRef() bool {
	return a.config.JiraBodyRef()
}

// JiraStrict delegates to the underlying config.
func (a *ConfigAdapter) JiraStrict() bool {
	return a.config.JiraStrict()
}

// SignatureRequired delegates to the underlying config.
func (a *ConfigAdapter) SignatureRequired() bool {
	return a.config.SignatureRequired()
}

// SignOffRequired delegates to the underlying config.
func (a *ConfigAdapter) SignOffRequired() bool {
	return a.config.SignOffRequired()
}

// AllowMultipleSignOffs delegates to the underlying config.
func (a *ConfigAdapter) AllowMultipleSignOffs() bool {
	return a.config.AllowMultipleSignOffs()
}

// AllowedSignatureTypes delegates to the underlying config.
func (a *ConfigAdapter) AllowedSignatureTypes() []string {
	return a.config.AllowedSignatureTypes()
}

// IdentityPublicKeyURI delegates to the underlying config.
func (a *ConfigAdapter) IdentityPublicKeyURI() string {
	return a.config.IdentityPublicKeyURI()
}

// SpellEnabled delegates to the underlying config.
func (a *ConfigAdapter) SpellEnabled() bool {
	return a.config.SpellEnabled()
}

// SpellLocale delegates to the underlying config.
func (a *ConfigAdapter) SpellLocale() string {
	return a.config.SpellLocale()
}

// SpellMaxErrors delegates to the underlying config.
func (a *ConfigAdapter) SpellMaxErrors() int {
	return a.config.SpellMaxErrors()
}

// SpellIgnoreWords delegates to the underlying config.
func (a *ConfigAdapter) SpellIgnoreWords() []string {
	return a.config.SpellIgnoreWords()
}

// SpellCustomWords delegates to the underlying config.
func (a *ConfigAdapter) SpellCustomWords() map[string]string {
	// Create a copy of the custom words map to maintain immutability
	if a.config.SpellCheck.CustomWords == nil {
		return make(map[string]string)
	}

	// Return a copy of the map
	result := make(map[string]string, len(a.config.SpellCheck.CustomWords))
	for word, replacement := range a.config.SpellCheck.CustomWords {
		result[word] = replacement
	}

	return result
}

// ReferenceBranch delegates to the underlying config.
func (a *ConfigAdapter) ReferenceBranch() string {
	return a.config.ReferenceBranch()
}

// IgnoreMergeCommits delegates to the underlying config.
func (a *ConfigAdapter) IgnoreMergeCommits() bool {
	return a.config.IgnoreMergeCommits()
}

// MaxCommitsAhead delegates to the underlying config.
func (a *ConfigAdapter) MaxCommitsAhead() int {
	return a.config.MaxCommitsAhead()
}

// CheckCommitsAhead delegates to the underlying config.
func (a *ConfigAdapter) CheckCommitsAhead() bool {
	return a.config.CheckCommitsAhead()
}

// EnabledRules delegates to the underlying config.
func (a *ConfigAdapter) EnabledRules() []string {
	return a.config.EnabledRules()
}

// DisabledRules delegates to the underlying config.
func (a *ConfigAdapter) DisabledRules() []string {
	return a.config.DisabledRules()
}

// GetActiveRules implements the domain.RuleConfigProvider interface.
// We don't need active rules here as the validation service will calculate them as needed.
func (a *ConfigAdapter) GetActiveRules() []string {
	// Return an empty slice or enabled rules if available
	if len(a.config.EnabledRules()) > 0 {
		return a.config.EnabledRules()
	}

	return []string{}
}

// GetAvailableRules implements the domain.RuleConfigProvider interface.
func (a *ConfigAdapter) GetAvailableRules() []string {
	// Return an empty slice - the rule provider will determine available rules
	return []string{}
}

// These methods are used by the validation service to set rules
// We'll update the underlying config to maintain immutability

// SetEnabledRules sets the enabled rules.
func (a *ConfigAdapter) SetEnabledRules(ruleNames []string) {
	a.config = a.config.WithEnabledRules(ruleNames)
}

// SetDisabledRules sets the disabled rules.
func (a *ConfigAdapter) SetDisabledRules(ruleNames []string) {
	a.config = a.config.WithDisabledRules(ruleNames)
}

// createValidationServiceWithDeps creates a validation service using injected dependencies.
// This function follows the hexagonal architecture pattern with explicit dependency injection.
func createValidationServiceWithDeps(deps *AppDependencies, repoPath string) (validate.ValidationService, error) {
	// Get config from manager
	unifiedConfig := deps.ConfigManager.GetValidationConfig()

	// Adapt the config to match the ValidationConfig interface
	validationConfig := NewConfigAdapter(unifiedConfig)

	// Create a repository adapter
	repoAdapter, err := git.NewRepositoryAdapter(repoPath)
	if err != nil {
		return validate.ValidationService{}, fmt.Errorf("failed to create repository adapter: %w", err)
	}

	// Return the validation service with dependencies
	return validate.CreateValidationServiceWithDependencies(
		validationConfig,
		repoAdapter, // GitCommitService
		repoAdapter, // RepositoryInfoProvider
		repoAdapter, // CommitAnalyzer
	), nil
}
