// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/itiquette/gommitlint/internal/adapters/config"
	"github.com/itiquette/gommitlint/internal/domain"
	configTypes "github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

// NewConfigCommand creates the config subcommand.
func NewConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Configuration operations",
		Description: `Operations for viewing and managing gommitlint configuration.

Examples:
  # Show current effective configuration
  gommitlint config show
  
  # Generate complete config template with all options
  gommitlint config init > .gommitlint.yaml
  
  # Show current config in YAML format  
  gommitlint config show --format=yaml
  
  # Show configuration in JSON format
  gommitlint config show --format=json`,

		Commands: []*cli.Command{
			{
				Name:  "show",
				Usage: "Show effective configuration",
				Description: `Shows the effective configuration after merging defaults, 
config files, and environment variables.

Output formats:
  text - Human-readable format (default)
  json - Machine-readable JSON format  
  yaml - Config file template format`,

				Action: func(ctx context.Context, cmd *cli.Command) error {
					return ExecuteConfigShow(ctx, cmd)
				},
			},
			{
				Name:  "init",
				Usage: "Generate complete configuration file template",
				Description: `Generates a complete .gommitlint.yaml template file with all 
available options set to their default values.

This is useful for:
  - Discovering all available configuration options
  - Creating a starting point for customization
  - Understanding the complete configuration structure

Examples:
  # Generate config file with all defaults
  gommitlint config init > .gommitlint.yaml
  
  # Generate and immediately edit
  gommitlint config init > .gommitlint.yaml && $EDITOR .gommitlint.yaml`,

				Action: func(ctx context.Context, cmd *cli.Command) error {
					return ExecuteConfigInit(ctx, cmd)
				},
			},
		},
	}
}

// ExecuteConfigShow handles the config show subcommand.
func ExecuteConfigShow(_ context.Context, cmd *cli.Command) error {
	// Load configuration
	cfgResult, err := LoadConfigFromCommand(cmd.Root())
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get format from root command flags
	format := cmd.Root().String("format")

	// Build effective configuration
	cfg := cfgResult.Config
	commitRules := rules.CreateCommitRules(cfg)
	repoRules := rules.CreateRepositoryRules(cfg)
	effectiveConfig := BuildEffectiveConfig(cfg, commitRules, repoRules, cfgResult.Source)

	// Format and output
	var exitCode int

	switch format {
	case "json":
		exitCode, err = printConfigJSON(effectiveConfig, os.Stdout)
	case "yaml":
		exitCode, err = printConfigYAML(effectiveConfig, os.Stdout)
	default:
		exitCode, err = printConfigText(effectiveConfig, os.Stdout)
	}

	if err != nil {
		return fmt.Errorf("failed to print configuration: %w", err)
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}

// ExecuteConfigInit handles the config init subcommand.
func ExecuteConfigInit(_ context.Context, _ *cli.Command) error {
	// Load default configuration (ignore any existing config files)
	defaultCfg := config.LoadDefaultConfig()

	// Generate a complete template with all defaults
	if err := printConfigInitYAML(defaultCfg, os.Stdout); err != nil {
		return fmt.Errorf("failed to generate config template: %w", err)
	}

	return nil
}

// EffectiveConfig represents the resolved configuration with enabled rules.
type EffectiveConfig struct {
	Config       configTypes.Config `json:"config"`
	EnabledRules []string           `json:"enabled_rules"`
	ConfigSource string             `json:"config_source"`
}

// BuildEffectiveConfig creates an EffectiveConfig from components.
func BuildEffectiveConfig(cfg configTypes.Config, commitRules []domain.CommitRule, repoRules []domain.RepositoryRule, configSource string) EffectiveConfig {
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
		ConfigSource: configSource,
	}
}

func printConfigJSON(effectiveConfig EffectiveConfig, output *os.File) (int, error) {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(effectiveConfig); err != nil {
		return 1, fmt.Errorf("failed to encode config as JSON: %w", err)
	}

	return 0, nil
}

func printConfigYAML(effectiveConfig EffectiveConfig, output *os.File) (int, error) {
	// For YAML output, we output the config wrapped in gommitlint: root key
	// We add a header comment to make it clear this is a generated config
	fmt.Fprintln(output, "# Gommitlint configuration file")
	fmt.Fprintln(output, "# Generated by: gommitlint --print-config --format=yaml")
	fmt.Fprintf(output, "# Configuration source: %s\n", effectiveConfig.ConfigSource)
	fmt.Fprintln(output, "#")
	fmt.Fprintln(output, "# You can use this as a starting point for your .gommitlint.yaml")
	fmt.Fprintln(output, "# Customize the values below according to your project's needs")
	fmt.Fprintln(output, "")

	// Wrap the config in the gommitlint root structure
	configWithRoot := map[string]interface{}{
		"gommitlint": effectiveConfig.Config,
	}

	encoder := yaml.NewEncoder(output)
	encoder.SetIndent(2)
	defer encoder.Close()

	if err := encoder.Encode(configWithRoot); err != nil {
		return 1, fmt.Errorf("failed to encode config as YAML: %w", err)
	}

	return 0, nil
}

func printConfigText(effectiveConfig EffectiveConfig, output *os.File) (int, error) {
	cfg := effectiveConfig.Config
	rules := effectiveConfig.EnabledRules

	fmt.Fprintln(output, "Gommitlint Configuration")
	fmt.Fprintln(output, "========================")
	fmt.Fprintln(output)

	// Show configuration source
	fmt.Fprintf(output, "Configuration Source: %s\n", effectiveConfig.ConfigSource)
	fmt.Fprintln(output)

	// Sort rules for consistent output
	sort.Strings(rules)

	fmt.Fprintf(output, "Enabled Rules (%d):\n", len(rules))

	if len(rules) == 0 {
		fmt.Fprintln(output, "  (none)")
	} else {
		for _, rule := range rules {
			fmt.Fprintf(output, "  - %s\n", rule)
		}
	}

	fmt.Fprintln(output)

	// Rules Configuration
	fmt.Fprintln(output, "Rules Configuration:")

	if len(cfg.Rules.Enabled) > 0 {
		fmt.Fprintf(output, "  Explicitly Enabled: %v\n", cfg.Rules.Enabled)
	} else {
		fmt.Fprintln(output, "  Explicitly Enabled: (none)")
	}

	if len(cfg.Rules.Disabled) > 0 {
		fmt.Fprintf(output, "  Explicitly Disabled: %v\n", cfg.Rules.Disabled)
	} else {
		fmt.Fprintln(output, "  Explicitly Disabled: (none)")
	}

	fmt.Fprintln(output)

	// Message Configuration
	fmt.Fprintln(output, "Message Configuration:")
	fmt.Fprintf(output, "  Subject Max Length: %d\n", cfg.Message.Subject.MaxLength)
	fmt.Fprintf(output, "  Subject Case: %s\n", cfg.Message.Subject.Case)
	fmt.Fprintf(output, "  Require Imperative: %t\n", cfg.Message.Subject.RequireImperative)

	if len(cfg.Message.Subject.ForbidEndings) > 0 {
		fmt.Fprintf(output, "  Forbid Endings: %v\n", cfg.Message.Subject.ForbidEndings)
	}

	fmt.Fprintf(output, "  Body Required: %t\n", cfg.Message.Body.Required)
	fmt.Fprintf(output, "  Body Min Length: %d\n", cfg.Message.Body.MinLength)
	fmt.Fprintf(output, "  Allow Signoff Only: %t\n", cfg.Message.Body.AllowSignoffOnly)
	fmt.Fprintf(output, "  Min Signoff Count: %d\n", cfg.Message.Body.MinSignoffCount)
	fmt.Fprintln(output)

	// Conventional Commit Configuration
	fmt.Fprintln(output, "Conventional Commit Configuration:")
	fmt.Fprintf(output, "  Require Scope: %t\n", cfg.Conventional.RequireScope)
	fmt.Fprintf(output, "  Max Description Length: %d\n", cfg.Conventional.MaxDescriptionLength)
	fmt.Fprintf(output, "  Allow Breaking Changes: %t\n", cfg.Conventional.AllowBreaking)

	if len(cfg.Conventional.Types) > 0 {
		fmt.Fprintf(output, "  Allowed Types: %v\n", cfg.Conventional.Types)
	}

	if len(cfg.Conventional.Scopes) > 0 {
		fmt.Fprintf(output, "  Allowed Scopes: %v\n", cfg.Conventional.Scopes)
	}

	fmt.Fprintln(output)

	// Signature Configuration
	fmt.Fprintln(output, "Signature Configuration:")
	fmt.Fprintf(output, "  Required: %t\n", cfg.Signature.Required)
	fmt.Fprintf(output, "  Verify Format: %t\n", cfg.Signature.VerifyFormat)

	if cfg.Signature.KeyDirectory != "" {
		fmt.Fprintf(output, "  Key Directory: %s\n", cfg.Signature.KeyDirectory)
	}

	if len(cfg.Signature.AllowedSigners) > 0 {
		fmt.Fprintf(output, "  Allowed Signers: %v\n", cfg.Signature.AllowedSigners)
	}

	fmt.Fprintln(output)

	// Identity Configuration
	fmt.Fprintln(output, "Identity Configuration:")

	if len(cfg.Identity.AllowedAuthors) > 0 {
		fmt.Fprintf(output, "  Allowed Authors: %v\n", cfg.Identity.AllowedAuthors)
	} else {
		fmt.Fprintln(output, "  Allowed Authors: (any)")
	}

	fmt.Fprintln(output)

	// Repository Configuration
	fmt.Fprintln(output, "Repository Configuration:")
	fmt.Fprintf(output, "  Max Commits Ahead: %d\n", cfg.Repo.MaxCommitsAhead)
	fmt.Fprintf(output, "  Reference Branch: %s\n", cfg.Repo.ReferenceBranch)
	fmt.Fprintf(output, "  Allow Merge Commits: %t\n", cfg.Repo.AllowMergeCommits)
	fmt.Fprintln(output)

	// JIRA Configuration
	fmt.Fprintln(output, "JIRA Configuration:")
	fmt.Fprintf(output, "  Require In Subject: %t\n", cfg.Jira.RequireInSubject)
	fmt.Fprintf(output, "  Require In Body: %t\n", cfg.Jira.RequireInBody)

	if len(cfg.Jira.ProjectPrefixes) > 0 {
		fmt.Fprintf(output, "  Project Prefixes: %v\n", cfg.Jira.ProjectPrefixes)
	}

	if len(cfg.Jira.IgnoreTicketPatterns) > 0 {
		fmt.Fprintf(output, "  Ignore Ticket Patterns: %v\n", cfg.Jira.IgnoreTicketPatterns)
	}

	fmt.Fprintln(output)

	// Spell Configuration
	fmt.Fprintln(output, "Spell Configuration:")
	fmt.Fprintf(output, "  Locale: %s\n", cfg.Spell.Locale)

	if len(cfg.Spell.IgnoreWords) > 0 {
		fmt.Fprintf(output, "  Ignore Words: %v\n", cfg.Spell.IgnoreWords)
	}

	fmt.Fprintln(output)

	// Output Configuration
	fmt.Fprintln(output, "Output Configuration:")
	fmt.Fprintf(output, "  Format: %s\n", cfg.Output)
	fmt.Fprintln(output)

	if _, err := fmt.Fprintln(output, "Use --format=json for machine-readable output"); err != nil {
		return 1, fmt.Errorf("failed to write output: %w", err)
	}

	if _, err := fmt.Fprintln(output, "Use --format=yaml to generate a config file output"); err != nil {
		return 1, fmt.Errorf("failed to write output: %w", err)
	}

	return 0, nil
}

// printConfigInitYAML generates a complete configuration template with all defaults.
func printConfigInitYAML(cfg configTypes.Config, output *os.File) error {
	// Generate comprehensive header comment
	fmt.Fprintln(output, "# Gommitlint Configuration Template")
	fmt.Fprintln(output, "# Generated by: gommitlint config init")
	fmt.Fprintln(output, "#")
	fmt.Fprintln(output, "# This file contains ALL available configuration options with their default values.")
	fmt.Fprintln(output, "# Uncomment and modify any settings you want to customize.")
	fmt.Fprintln(output, "# Remove sections you don't need to keep the config file clean.")
	fmt.Fprintln(output, "#")
	fmt.Fprintln(output, "# For more information, see: https://github.com/itiquette/gommitlint")
	fmt.Fprintln(output, "")

	// Wrap the config in the gommitlint root structure
	configWithRoot := map[string]interface{}{
		"gommitlint": cfg,
	}

	encoder := yaml.NewEncoder(output)
	encoder.SetIndent(2)
	defer encoder.Close()

	if err := encoder.Encode(configWithRoot); err != nil {
		return fmt.Errorf("failed to encode config template as YAML: %w", err)
	}

	// Add footer comment
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "# Configuration complete!")
	fmt.Fprintln(output, "# Save this file as .gommitlint.yaml in your project root.")

	return nil
}
