// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	cliAdapter "github.com/itiquette/gommitlint/internal/adapters/cli"
	"github.com/itiquette/gommitlint/internal/adapters/config"
	"github.com/itiquette/gommitlint/internal/adapters/git"
	logadapter "github.com/itiquette/gommitlint/internal/adapters/logging"
	"github.com/itiquette/gommitlint/internal/adapters/signing"
	"github.com/itiquette/gommitlint/internal/domain"
	configTypes "github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/urfave/cli/v3"
)

// These variables are set by the build process.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Create the root context
	ctx := context.Background()

	// Initialize logger early in the application flow
	ctx = logadapter.InitLogger(ctx, nil, "text") // Basic logger setup

	app := &cli.Command{
		Name:    "gommitlint",
		Usage:   "Commit validator",
		Version: fmt.Sprintf("%s (Commit: %s, Build date: %s)", version, commit, date),

		// Enable shell completion for all supported shells
		EnableShellCompletion: true,

		// Global flags
		Flags: []cli.Flag{
			// Configuration flags
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "specify config file location",
				Category: "Configuration",
			},
			&cli.BoolFlag{
				Name:     "ignore-config",
				Usage:    "ignore config files",
				Category: "Configuration",
			},
			&cli.BoolFlag{
				Name:     "print-config",
				Usage:    "show effective configuration and exit",
				Category: "Configuration",
			},

			// Repository flags
			&cli.StringFlag{
				Name:     "repo-path",
				Usage:    "path to the repository (defaults to current directory)",
				Category: "Global",
			},

			// Output flags
			&cli.StringFlag{
				Name:     "format",
				Value:    "text",
				Usage:    "output format (text, json, github, gitlab)",
				Category: "Output",
			},
			&cli.StringFlag{
				Name:     "color",
				Value:    "auto",
				Usage:    "colorize output (auto, always, never)",
				Category: "Output",
			},
			&cli.StringFlag{
				Name:     "verbosity",
				Value:    "brief",
				Usage:    "log level (quiet, brief, trace)",
				Category: "Output",
			},
			&cli.BoolFlag{
				Name:     "quiet",
				Aliases:  []string{"q"},
				Usage:    "suppress all output except errors",
				Category: "Output",
			},
			&cli.BoolFlag{
				Name:     "caller",
				Usage:    "include caller information in logs",
				Category: "Output",
			},
		},

		// Before hook for global setup
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// Setup logging based on flags
			output := cmd.String("format")
			ctx = logadapter.InitLogger(ctx, nil, output) // Pass nil for now

			// Handle print-config flag
			if cmd.Bool("print-config") {
				cfg, err := loadConfigFromCommand(cmd)
				if err != nil {
					return ctx, fmt.Errorf("failed to load configuration: %w", err)
				}

				printConfigAndExit(cfg, cmd.String("format"))

				return ctx, nil
			}

			return ctx, nil
		},

		Action: func(_ context.Context, cmd *cli.Command) error {
			// If no subcommand, show help
			return cli.ShowAppHelp(cmd)
		},

		Commands: []*cli.Command{
			validateCommand(),
			installHookCommand(),
			removeHookCommand(),
		},
	}

	if err := app.Run(ctx, os.Args); err != nil {
		// Get logger from context and handle error
		zerologLogger := logadapter.GetLogger(ctx)
		logger := logadapter.NewDomainLogger(zerologLogger)
		logger.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}

// loadConfigFromCommand loads configuration based on command flags.
func loadConfigFromCommand(cmd *cli.Command) (configTypes.Config, error) {
	configPath := cmd.String("config")
	ignoreConfig := cmd.Bool("ignore-config")

	if configPath != "" && ignoreConfig {
		return configTypes.Config{}, errors.New("cannot specify both --config and --ignore-config flags")
	}

	if ignoreConfig {
		// Load only defaults and environment, no file config
		defaultCfg := config.LoadDefaultConfig()
		envCfg := config.LoadEnvConfig()

		return config.MergeConfigs(defaultCfg, envCfg)
	}

	if configPath != "" {
		// Check if specified config file exists
		if _, err := os.Stat(configPath); err != nil {
			return configTypes.Config{}, fmt.Errorf("config file '%s' not found: %w", configPath, err)
		}
		// Load from specific path
		return config.LoadConfigFromPath(configPath)
	}

	// Default behavior - use existing LoadConfig()
	return config.LoadConfig()
}

// printConfigAndExit prints configuration and exits.
func printConfigAndExit(cfg configTypes.Config, format string) {
	// Determine which rules are enabled
	commitRules := rules.CreateCommitRules(cfg)
	repoRules := rules.CreateRepositoryRules(cfg)

	// Build a structure showing effective configuration
	effectiveConfig := buildEffectiveConfig(cfg, commitRules, repoRules)

	// Format and output based on requested format
	var exitCode int

	var err error

	switch format {
	case "json":
		exitCode, err = printConfigJSON(effectiveConfig, os.Stdout)
	default:
		exitCode, err = printConfigText(effectiveConfig, os.Stdout)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(exitCode)
}

// validateCommand creates the validate subcommand.
func validateCommand() *cli.Command {
	return &cli.Command{
		Name:    "validate",
		Aliases: []string{"val", "v"},
		Usage:   "Validate commit messages",
		Description: `Validates commit message/s against a set of rules.

Examples:
  # Validate commits in the current branch against main
  gommitlint validate --base-branch=main
  
  # Validate a specific commit
  gommitlint validate --ref=HEAD~1
  
  # Validate a commit message from a file
  gommitlint validate --message-file=/path/to/commit-msg.txt
  
  # Validate a range of commits
  gommitlint validate --range=main..feature
  
  # Validate last 5 commits
  gommitlint validate --count=5`,

		Flags: []cli.Flag{
			// Validation Target flags (choose one)
			&cli.StringFlag{
				Name:     "message-file",
				Aliases:  []string{"f"},
				Usage:    "commit message file path to validate",
				Category: "Validation Target (choose one)",
			},
			&cli.StringFlag{
				Name:     "ref",
				Aliases:  []string{"r"},
				Usage:    "git reference to validate (defaults to HEAD)",
				Category: "Validation Target (choose one)",
			},
			&cli.IntFlag{
				Name:     "count",
				Aliases:  []string{"n"},
				Value:    1,
				Usage:    "number of commits from HEAD to validate",
				Category: "Validation Target (choose one)",
			},
			&cli.StringFlag{
				Name:     "range",
				Usage:    "range of commits to validate (<commit1>..<commit2>)",
				Category: "Validation Target (choose one)",
			},
			&cli.StringFlag{
				Name:     "base-branch",
				Usage:    "base branch to compare with (validates commits in <base-branch>..HEAD)",
				Category: "Validation Target (choose one)",
			},

			// Output flags
			&cli.IntFlag{
				Name:     "verbose",
				Aliases:  []string{"v"},
				Usage:    "show detailed validation results (1 for verbose, 2 for extra verbose)",
				Category: "Output Options",
			},
			&cli.StringFlag{
				Name:     "rule-help",
				Usage:    "show detailed help for a specific rule",
				Category: "Output Options",
			},
			&cli.StringFlag{
				Name:     "report-file",
				Usage:    "write validation results to file instead of stdout",
				Category: "Output Options",
			},
		},

		Action: func(ctx context.Context, cmd *cli.Command) error {
			return executeValidation(ctx, cmd)
		},
	}
}

// installHookCommand creates the install-hook subcommand.
func installHookCommand() *cli.Command {
	return &cli.Command{
		Name:    "install-hook",
		Aliases: []string{"install"},
		Usage:   "Install Git commit-msg hook for validation",
		Description: `Installs a Git commit-msg hook to automatically validate commit messages.

Examples:
  # Install commit-msg hook in the current repository
  gommitlint install-hook

  # Install commit-msg hook with force
  gommitlint install-hook --force`,

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "overwrite existing hook if it exists",
			},
		},

		Action: func(ctx context.Context, cmd *cli.Command) error {
			return executeInstallHook(ctx, cmd)
		},
	}
}

// removeHookCommand creates the remove-hook subcommand.
func removeHookCommand() *cli.Command {
	return &cli.Command{
		Name:    "remove-hook",
		Aliases: []string{"remove", "uninstall"},
		Usage:   "Remove Git hooks for commit validation",
		Description: `Removes Git hooks previously installed by gommitlint.

Examples:
  # Remove commit-msg hook from the current repository
  gommitlint remove-hook
  
  # Remove hook without confirmation prompt
  gommitlint remove-hook --yes`,

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "skip confirmation prompt",
			},
		},

		Action: func(ctx context.Context, cmd *cli.Command) error {
			return executeRemoveHook(ctx, cmd)
		},
	}
}

// executeValidation orchestrates the validation process.
func executeValidation(ctx context.Context, cmd *cli.Command) error {
	// Load configuration
	cfg, err := loadConfigFromCommand(cmd.Root())
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create logger from context
	zerologLogger := logadapter.GetLogger(ctx)
	logger := logadapter.NewDomainLogger(zerologLogger)

	// Create validation target from CLI flags
	target, err := createValidationTarget(cmd)
	if err != nil {
		return fmt.Errorf("failed to create validation target: %w", err)
	}

	// Create output options from CLI flags
	outputOptions, err := createOutputOptions(cmd)
	if err != nil {
		return fmt.Errorf("failed to create output options: %w", err)
	}

	// Handle rule help if requested
	if outputOptions.ShowRuleHelp() {
		return handleRuleHelp(outputOptions, cfg)
	}

	// Create Git repository
	repoPath := getRepoPath(cmd)

	repo, err := git.NewRepository(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Create rules from configuration
	commitRules := rules.CreateCommitRules(cfg)
	repoRules := rules.CreateRepositoryRules(cfg)

	// Execute validation
	report, err := cliAdapter.ValidateTarget(ctx, target, commitRules, repoRules, repo, cfg, logger)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Write output
	err = outputOptions.WriteReport(report)
	if err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Return non-zero exit code if validation failed
	if !report.Summary.AllPassed {
		os.Exit(1)
	}

	return nil
}

// createValidationTarget creates a ValidationTarget from CLI flags.
func createValidationTarget(cmd *cli.Command) (cliAdapter.ValidationTarget, error) {
	messageFile := cmd.String("message-file")
	gitRef := cmd.String("ref")
	commitRange := cmd.String("range")
	baseBranch := cmd.String("base-branch")
	commitCount := cmd.Int("count")

	return cliAdapter.NewValidationTarget(messageFile, gitRef, commitRange, baseBranch, commitCount)
}

// createOutputOptions creates OutputOptions from CLI flags.
func createOutputOptions(cmd *cli.Command) (cliAdapter.OutputOptions, error) {
	// Determine output writer
	var writer *os.File

	reportFile := cmd.String("report-file")
	if reportFile != "" {
		file, err := os.Create(reportFile)
		if err != nil {
			return cliAdapter.OutputOptions{}, fmt.Errorf("failed to create report file: %w", err)
		}

		writer = file
	} else {
		writer = os.Stdout
	}

	// Get format from root command flags
	format := cmd.Root().String("format")
	color := cmd.Root().String("color")
	quiet := cmd.Root().Bool("quiet")

	// Create base options
	options := cliAdapter.NewOutputOptions(writer).
		WithFormat(format).
		WithColor(color)

	// Handle verbose flags
	verboseLevel := cmd.Int("verbose")
	if verboseLevel > 0 && !quiet {
		options = options.WithVerbose(true)
	}

	// Handle rule help
	ruleHelp := cmd.String("rule-help")
	if ruleHelp != "" {
		options = options.WithRuleHelp(ruleHelp)
		if err := options.ValidateRuleHelp(); err != nil {
			return cliAdapter.OutputOptions{}, err
		}
	}

	return options, nil
}

// getRepoPath gets the repository path from CLI flags or defaults to current directory.
func getRepoPath(cmd *cli.Command) string {
	repoPath := cmd.Root().String("repo-path")
	if repoPath == "" {
		repoPath = "."
	}

	return repoPath
}

// handleRuleHelp shows help for a specific rule and exits.
func handleRuleHelp(options cliAdapter.OutputOptions, _ configTypes.Config) error {
	// For rule help, we create a minimal report showing rule information
	// This is a simplified implementation - in a full implementation you might
	// want to create a dedicated help system
	fmt.Printf("Help for rule: %s\n", options.GetRuleHelp())
	fmt.Println("(Rule help display not yet fully implemented)")

	return nil
}

// EffectiveConfig represents the resolved configuration with enabled rules.
type EffectiveConfig struct {
	Config       configTypes.Config `json:"config"`
	EnabledRules []string           `json:"enabled_rules"`
}

func buildEffectiveConfig(cfg configTypes.Config, commitRules []domain.CommitRule, repoRules []domain.RepositoryRule) EffectiveConfig {
	// Collect enabled rule names
	enabledRules := make([]string, 0, len(commitRules)+len(repoRules))

	for _, rule := range commitRules {
		enabledRules = append(enabledRules, rule.Name())
	}

	for _, rule := range repoRules {
		enabledRules = append(enabledRules, rule.Name())
	}

	return EffectiveConfig{
		Config:       cfg,
		EnabledRules: enabledRules,
	}
}

func printConfigJSON(_ EffectiveConfig, output *os.File) (int, error) {
	fmt.Fprintln(output, "JSON config printing not yet implemented")

	return 0, nil
}

func printConfigText(_ EffectiveConfig, output *os.File) (int, error) {
	fmt.Fprintln(output, "Text config printing not yet implemented")

	return 0, nil
}

// executeInstallHook orchestrates the hook installation process.
func executeInstallHook(ctx context.Context, cmd *cli.Command) error {
	// Get flags
	force := cmd.Bool("force")
	repoPath := getRepoPath(cmd)

	// Create logger from context
	zerologLogger := logadapter.GetLogger(ctx)
	logger := logadapter.NewDomainLogger(zerologLogger)

	// Install the hook
	if err := installHook(force, repoPath); err != nil {
		logger.Error("Hook installation failed", "error", err)

		return err
	}

	fmt.Fprintln(cmd.Writer, "✅ Git commit-msg hook installed successfully!")

	return nil
}

// executeRemoveHook orchestrates the hook removal process.
func executeRemoveHook(ctx context.Context, cmd *cli.Command) error {
	// Get flags
	skipConfirm := cmd.Bool("yes")
	repoPath := getRepoPath(cmd)

	// Create logger from context
	zerologLogger := logadapter.GetLogger(ctx)
	logger := logadapter.NewDomainLogger(zerologLogger)

	// Remove the hook
	if err := removeHook(cmd, repoPath, skipConfirm); err != nil {
		logger.Error("Hook removal failed", "error", err)

		return err
	}

	fmt.Fprintln(cmd.Writer, "✅ Git hook removed successfully!")

	return nil
}

// installHook installs a Git commit-msg hook in the specified repository.
// It uses atomic file operations and proper permission management to ensure security.
func installHook(force bool, repoPath string) error {
	// Validate and normalize the repository path using signing utilities
	validatedPath, err := signing.ValidateGitRepoPath(repoPath)
	if err != nil {
		return fmt.Errorf("invalid repository path: %w", err)
	}

	// Create parameters with defaults
	params := NewHookInstallationParameters(force, validatedPath)

	// Ensure hooks directory exists
	if err := params.EnsureHooksDirectory(); err != nil {
		return err
	}

	// Check if we can install the hook
	if err := params.CanInstallHook(); err != nil {
		return err
	}

	// Get the hook path
	hookPath, err := params.FindHookPath()
	if err != nil {
		return err
	}

	// Get the hook content
	hookContent := params.GetHookContent()

	// Write the hook file using our secure file writing function
	// This handles creating a temporary file, setting permissions, and atomically
	// renaming it to the final destination to prevent TOCTOU vulnerabilities
	if err := signing.SafeWriteFile(hookPath, []byte(hookContent), 0700); err != nil {
		return fmt.Errorf("could not write hook file: %w", err)
	}

	return nil
}

// removeHook removes a Git hook from the specified repository.
func removeHook(cmd *cli.Command, repoPath string, skipConfirm bool) error {
	// Validate and normalize the repository path using signing utilities
	validatedPath, err := signing.ValidateGitRepoPath(repoPath)
	if err != nil {
		return fmt.Errorf("invalid repository path: %w", err)
	}

	// Create parameters with defaults
	params := NewHookRemovalParameters(cmd, validatedPath, skipConfirm)

	// Verify the hook exists
	if err := params.VerifyHookExists(); err != nil {
		return err
	}

	// Check if the hook was installed by gommitlint
	isGommitlintHook, err := params.IsGommitlintHook()
	if err != nil {
		return err
	}

	// If not a gommitlint hook, ask for confirmation
	if !isGommitlintHook {
		confirmed, err := params.ConfirmRemoval()
		if err != nil {
			return err
		}

		if !confirmed {
			return errors.New("operation cancelled by user")
		}
	}

	// Get hook path (pure function)
	hookPath, err := params.GetHookPath()
	if err != nil {
		return err
	}

	// Remove the hook file (side effect isolated)
	return RemoveHookFile(hookPath)
}

// HookInstallationParameters contains all parameters needed for hook installation.
// This structure supports safe operations.
type HookInstallationParameters struct {
	Force         bool
	RepoPath      string
	HookType      string
	PathValidator cliAdapter.PathValidator
}

// NewHookInstallationParameters creates HookInstallationParameters with defaults.
func NewHookInstallationParameters(force bool, repoPath string) HookInstallationParameters {
	// Default to commit-msg hook type (currently the only supported type)
	return HookInstallationParameters{
		Force:         force,
		RepoPath:      repoPath,
		HookType:      "commit-msg",
		PathValidator: cliAdapter.DefaultPathValidator(),
	}
}

// FindHookPath determines the hook file path based on the parameters.
// Implements security best practices for safe path handling.
func (p HookInstallationParameters) FindHookPath() (string, error) {
	if p.PathValidator == nil {
		p.PathValidator = cliAdapter.DefaultPathValidator()
	}

	return p.PathValidator.ValidateHookPath(p.RepoPath, p.HookType)
}

// EnsureHooksDirectory ensures the hooks directory exists.
// Implements security best practices for safe path handling.
func (p HookInstallationParameters) EnsureHooksDirectory() error {
	if p.PathValidator == nil {
		p.PathValidator = cliAdapter.DefaultPathValidator()
	}

	return p.PathValidator.EnsureHooksDirectory(p.RepoPath)
}

// GetHookContent returns the content for the hook based on its type.
func (p HookInstallationParameters) GetHookContent() string {
	// Currently, only commit-msg hooks are supported
	return generateCommitMsgHook()
}

// CanInstallHook checks if a hook can be installed based on parameters.
func (p HookInstallationParameters) CanInstallHook() error {
	hookPath, err := p.FindHookPath()
	if err != nil {
		return err
	}

	// Check if hook already exists using file descriptor to prevent TOCTOU
	file, err := os.Open(hookPath)
	if err == nil {
		// File exists
		file.Close()

		if !p.Force {
			return fmt.Errorf("hook already exists at %s (use --force to overwrite)", hookPath)
		}
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return fmt.Errorf("cannot check hook existence: %w", err)
	}

	return nil
}

// HookRemovalParameters contains all parameters needed for hook removal.
type HookRemovalParameters struct {
	RepoPath      string
	SkipConfirm   bool
	HookType      string
	Output        io.Writer
	Input         io.Reader
	PathValidator cliAdapter.PathValidator
}

// NewHookRemovalParameters creates HookRemovalParameters with defaults.
func NewHookRemovalParameters(cmd *cli.Command, repoPath string, skipConfirm bool) HookRemovalParameters {
	return HookRemovalParameters{
		RepoPath:      repoPath,
		SkipConfirm:   skipConfirm,
		HookType:      "commit-msg", // Currently the only supported type
		Output:        cmd.Writer,
		Input:         cmd.Reader,
		PathValidator: cliAdapter.DefaultPathValidator(),
	}
}

// FindHookPath determines the hook file path based on the parameters.
// Applies security best practices for safe path handling using signing.
func (p HookRemovalParameters) FindHookPath() (string, error) {
	if p.PathValidator == nil {
		p.PathValidator = cliAdapter.DefaultPathValidator()
	}

	return p.PathValidator.ValidateHookPath(p.RepoPath, p.HookType)
}

// VerifyHookExists checks if the hook file exists.
func (p HookRemovalParameters) VerifyHookExists() error {
	hookPath, err := p.FindHookPath()
	if err != nil {
		return err
	}

	// Check if hook exists using file descriptor to prevent TOCTOU
	file, err := os.Open(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("hook does not exist at %s", hookPath)
		}

		return fmt.Errorf("cannot check hook existence: %w", err)
	}

	file.Close()

	return nil
}

// IsGommitlintHook checks if the hook was installed by gommitlint.
func (p HookRemovalParameters) IsGommitlintHook() (bool, error) {
	hookPath, err := p.FindHookPath()
	if err != nil {
		return false, err
	}

	// Open file to check size first
	file, err := os.Open(hookPath)
	if err != nil {
		return false, fmt.Errorf("could not open hook file: %w", err)
	}
	defer file.Close()

	// Check file size to prevent reading huge files
	info, err := file.Stat()
	if err != nil {
		return false, fmt.Errorf("could not stat hook file: %w", err)
	}

	// Limit to 10KB - hooks should be small scripts
	const maxHookSize = 10 * 1024
	if info.Size() > maxHookSize {
		return false, fmt.Errorf("hook file too large: %d bytes (max %d)", info.Size(), maxHookSize)
	}

	// Now safe to read the file
	content := make([]byte, info.Size())

	_, err = io.ReadFull(file, content)
	if err != nil {
		return false, fmt.Errorf("could not read hook file: %w", err)
	}

	return strings.Contains(string(content), "gommitlint"), nil
}

// ConfirmRemoval asks the user to confirm removing a non-gommitlint hook.
func (p HookRemovalParameters) ConfirmRemoval() (bool, error) {
	if p.SkipConfirm {
		return true, nil
	}

	fmt.Fprintln(p.Output, "⚠️ Warning: The hook doesn't appear to be installed by gommitlint.")
	fmt.Fprint(p.Output, "Do you still want to remove it? [y/N]: ")

	// Read user input
	var response string

	_, err := fmt.Fscanln(p.Input, &response)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	// Check if response is affirmative
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes", nil
}

// GetHookPath returns the hook file path to be removed.
func (p HookRemovalParameters) GetHookPath() (string, error) {
	return p.FindHookPath()
}

// RemoveHookFile removes the hook file from the filesystem.
func RemoveHookFile(hookPath string) error {
	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("could not remove hook file: %w", err)
	}

	return nil
}

// generateCommitMsgHook generates content for the commit-msg hook.
func generateCommitMsgHook() string {
	return createDefaultHookScript()
}

// createDefaultHookScript creates a shell script for the commit-msg hook.
// It is used internally by generateCommitMsgHook.
func createDefaultHookScript() string {
	return `#!/bin/sh
#
# gommitlint commit-msg hook for validating commit messages.
# Generated by gommitlint install-hook command.
#

# Get the commit message file path
COMMIT_MSG_FILE=$1

# Run validation using gommitlint
gommitlint validate --message-file="$COMMIT_MSG_FILE"
RESULT=$?

# Return the validation result
exit $RESULT
`
}
