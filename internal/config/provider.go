// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Provider is an interface for accessing and manipulating configuration.
// It uses value semantics to ensure immutability of the underlying configuration.
type Provider struct {
	config Config
}

// NewProvider creates a new Provider with default configuration.
func NewProvider() (*Provider, error) {
	// Start with default config
	config := NewDefaultConfig()

	// Set default rules: Include all rules from DefaultDisabledRules
	defaultDisabledRules := append(
		[]string{}, // Create new slice to avoid modifying the original
		config.DisabledRules()...,
	)

	// Add rules that should be disabled by default from the central list
	for ruleName := range DefaultDisabledRules {
		// Check if the rule is explicitly enabled in the default config
		isExplicitlyEnabled := false
		for _, enabledRule := range config.Rules.EnabledRules {
			if enabledRule == ruleName {
				isExplicitlyEnabled = true
				break
			}
		}
		
		// Only add to disabled list if not explicitly enabled
		if !isExplicitlyEnabled {
			defaultDisabledRules = append(defaultDisabledRules, ruleName)
		}
	}

	// Note: When a rule is in both enabled_rules and disabled_rules,
	// it should be removed from disabled_rules (explicit enabling overrides disabled status)

	// Create new config with updated disabled rules
	config = config.WithRules(config.Rules.WithDisabledRules(defaultDisabledRules))

	return &Provider{
		config: config,
	}, nil
}

// NewProviderWithConfig creates a new Provider with the specified configuration.
func NewProviderWithConfig(config Config) *Provider {
	return &Provider{
		config: config,
	}
}

// GetConfig returns the current configuration.
func (p *Provider) GetConfig() Config {
	return p.config
}

// UpdateConfig applies a transformation to the current configuration and returns the provider.
func (p *Provider) UpdateConfig(transform func(Config) Config) *Provider {
	p.config = transform(p.config)

	return p
}

// ProviderKey is the key used to store the Provider in the context.
type ProviderKey struct{}

// WithConfigProvider returns a new context with the Provider stored under the ProviderKey.
func WithConfigProvider(ctx context.Context, provider *Provider) context.Context {
	return context.WithValue(ctx, ProviderKey{}, provider)
}

// WithProviderInContext adds a provider to the context.
func WithProviderInContext(ctx context.Context, provider *Provider) context.Context {
	return WithConfigProvider(ctx, provider)
}

// GetProviderFromContext retrieves a provider from the context.
func GetProviderFromContext(ctx context.Context) (*Provider, error) {
	return GetConfigProvider(ctx)
}

// GetConfigProvider retrieves the Provider from the context.
func GetConfigProvider(ctx context.Context) (*Provider, error) {
	provider, ok := ctx.Value(ProviderKey{}).(*Provider)
	if !ok {
		return nil, errors.New("config provider not found in context")
	}

	return provider, nil
}

// WithConfig returns a new context with the Config stored directly for simpler access.
func WithConfig(ctx context.Context, config Config) context.Context {
	provider := NewProviderWithConfig(config)

	return WithConfigProvider(ctx, provider)
}

// GetConfig retrieves the Config from the context via the ConfigProvider.
func GetConfig(ctx context.Context) Config {
	provider, err := GetConfigProvider(ctx)
	if err != nil {
		// Return default config if none is found in context
		return NewDefaultConfig()
	}

	config := provider.GetConfig()

	// Process enabled_rules vs disabled_rules to ensure explicitly enabled rules
	// take precedence over disabled rules
	enabledRules := config.Rules.EnabledRules
	disabledRules := config.Rules.DisabledRules

	// If any rule appears in both enabled and disabled lists, remove it from disabled
	if len(enabledRules) > 0 && len(disabledRules) > 0 {
		newDisabled := make([]string, 0, len(disabledRules))

		for _, disabledRule := range disabledRules {
			// Check if this rule is explicitly enabled
			isEnabled := false

			for _, enabledRule := range enabledRules {
				cleanDisabled := strings.TrimSpace(strings.Trim(disabledRule, "\"'"))
				cleanEnabled := strings.TrimSpace(strings.Trim(enabledRule, "\"'"))

				if cleanDisabled == cleanEnabled {
					isEnabled = true

					break
				}
			}

			// Only keep the rule in disabled_rules if it's not explicitly enabled
			if !isEnabled {
				newDisabled = append(newDisabled, disabledRule)
			}
		}

		// Update the disabled rules list
		if len(newDisabled) != len(disabledRules) {
			config = config.WithRules(config.Rules.WithDisabledRules(newDisabled))
		}
	}

	return config
}

// UpdateConfig applies a transformation to the current configuration in the context.
func UpdateConfig(ctx context.Context, transform func(Config) Config) context.Context {
	config := GetConfig(ctx)
	updatedConfig := transform(config)

	return WithConfig(ctx, updatedConfig)
}

// WithGitRepository configures the repository path in the context's configuration.
func WithGitRepository(ctx context.Context, repoPath string) context.Context {
	return UpdateConfig(ctx, func(cfg Config) Config {
		return cfg.WithRepository(cfg.Repository.WithPath(repoPath))
	})
}

// WithGitRepository sets the Git repository path in the configuration.
func (p *Provider) WithGitRepository(path string) *Provider {
	p.config = p.config.WithRepository(p.config.Repository.WithPath(path))

	return p
}

// Load loads the configuration from default paths.
func (p *Provider) Load() error {
	// Search for configuration in standard locations
	configPaths := []string{
		".gommitlint.yaml",
		".gommitlint.yml",
		".config/gommitlint/config.yaml",
		".config/gommitlint/config.yml",
	}

	// Try to load from each path
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			// Log successful find
			fmt.Fprintf(os.Stderr, "Found configuration file at %s\n", path)

			return p.LoadFromPath(path)
		}
	}

	// No config file found, but this is not an error - use defaults
	return nil
}

// loadConfigFile loads and parses the config file, returning file content and koanf instance.
func loadConfigFile(configPath string) (string, *koanf.Koanf, error) {
	koanfInstance := koanf.New(".")

	// Log file details
	fmt.Fprintf(os.Stderr, "Loading configuration from %s\n", configPath)

	// Show file content for debug
	content, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)

		return "", nil, fmt.Errorf("error reading config file %s: %w", configPath, err)
	}

	contentStr := string(content)
	fmt.Fprintf(os.Stderr, "Config file content:\n%s\n", contentStr)

	// Load YAML configuration
	if err := koanfInstance.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)

		return contentStr, nil, fmt.Errorf("error loading config file %s: %w", configPath, err)
	}

	// Log found keys
	fmt.Fprintf(os.Stderr, "Found keys in config: %v\n", koanfInstance.Keys())

	// Check for gommitlint.rules.enabled_rules key specifically
	if koanfInstance.Exists("gommitlint.rules.enabled_rules") {
		var rules []string
		if err := koanfInstance.Unmarshal("gommitlint.rules.enabled_rules", &rules); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling gommitlint.rules.enabled_rules: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "FOUND gommitlint.rules.enabled_rules = %+q\n", rules)
		}
	} else {
		fmt.Fprintf(os.Stderr, "WARNING: Key gommitlint.rules.enabled_rules not found in config\n")
	}

	return contentStr, koanfInstance, nil
}

// processEnabledRules handles the enabled_rules configuration.
func processEnabledRules(koanfInstance *koanf.Koanf) ([]string, bool) {
	var enabledRules []string

	haveEnabledRules := false

	// Extract enabled rules if present in the YAML configuration
	if koanfInstance.Exists("gommitlint.rules.enabled_rules") {
		if err := koanfInstance.Unmarshal("gommitlint.rules.enabled_rules", &enabledRules); err == nil {
			haveEnabledRules = true

			// Print the enabled rules for debugging
			fmt.Fprintf(os.Stderr, "CONFIG: Found enabled_rules in config: %v\n", enabledRules)
		}
	}

	return enabledRules, haveEnabledRules
}

// processDisabledRules extracts and processes the disabled_rules configuration.
func processDisabledRules(koanfInstance *koanf.Koanf) []string {
	var disabledRules []string

	if koanfInstance.Exists("gommitlint.rules.disabled_rules") {
		if err := koanfInstance.Unmarshal("gommitlint.rules.disabled_rules", &disabledRules); err == nil {
			// Print the disabled rules for debugging
			fmt.Fprintf(os.Stderr, "CONFIG: Found disabled_rules in config: %v\n", disabledRules)
		}
	}

	return disabledRules
}

// resolveRuleConflicts ensures enabled rules take precedence over disabled rules.
// The disabledRules parameter is not directly used because we rely on the
// current disabled rules from the config, which may already have been updated.
func resolveRuleConflicts(enabledRules []string, cfg Config) Config {
	// Create a cleaned disabled rules list that removes any explicitly enabled rules
	currentDisabled := cfg.Rules.DisabledRules
	newDisabled := make([]string, 0, len(currentDisabled))

	// Debug log for rules processing
	fmt.Fprintf(os.Stderr, "CONFIG: Processing rule enablement. Enabled rules: %v, Disabled rules: %v\n",
		enabledRules, currentDisabled)

	// Process all disabled rules
	for _, disabledRule := range currentDisabled {
		// Clean rule name for comparison
		cleanDisabled := strings.TrimSpace(strings.Trim(disabledRule, "\"'"))

		// Check if this rule is explicitly enabled (and thus should be removed from disabled)
		isExplicitlyEnabled := false

		for _, enabledRule := range enabledRules {
			cleanEnabled := strings.TrimSpace(strings.Trim(enabledRule, "\"'"))
			if cleanDisabled == cleanEnabled {
				// Important: This rule is explicitly enabled, so it should NOT be in disabled list
				isExplicitlyEnabled = true

				fmt.Fprintf(os.Stderr, "CONFIG: Removing '%s' from disabled_rules (it's explicitly enabled)\n",
					cleanDisabled)

				break
			}
		}

		if !isExplicitlyEnabled {
			newDisabled = append(newDisabled, disabledRule)
		}
	}

	// Check for rules that are in DefaultDisabledRules but explicitly enabled
	// These rules shouldn't be in the disabled list
	for ruleName := range DefaultDisabledRules {
		// Skip if already in disabled list (would have been processed above)
		alreadyInDisabled := false
		for _, disabledRule := range currentDisabled {
			cleanDisabled := strings.TrimSpace(strings.Trim(disabledRule, "\"'"))
			if cleanDisabled == ruleName {
				alreadyInDisabled = true
				break
			}
		}
		
		if alreadyInDisabled {
			continue // Already processed above
		}
		
		// Check if explicitly enabled - if not, add to disabled list
		isExplicitlyEnabled := false
		for _, enabledRule := range enabledRules {
			cleanEnabled := strings.TrimSpace(strings.Trim(enabledRule, "\"'"))
			if cleanEnabled == ruleName {
				isExplicitlyEnabled = true
				fmt.Fprintf(os.Stderr, "CONFIG: Not adding default-disabled rule '%s' to disabled list (it's explicitly enabled)\n",
					ruleName)
				break
			}
		}
		
		if !isExplicitlyEnabled {
			// Add to disabled list since it's default-disabled and not explicitly enabled
			newDisabled = append(newDisabled, ruleName)
			fmt.Fprintf(os.Stderr, "CONFIG: Adding default-disabled rule '%s' to disabled list\n", ruleName)
		}
	}

	// Apply the cleaned disabled rules (with explicitly enabled rules removed)
	cfg = cfg.WithRules(cfg.Rules.WithDisabledRules(newDisabled))
	fmt.Fprintf(os.Stderr, "CONFIG: After resolving rule conflicts - Disabled rules: %v\n",
		newDisabled)

	return cfg
}

// LoadFromPath loads the configuration from the specified path.
func (p *Provider) LoadFromPath(configPath string) error {
	// Load the config file
	_, koanfInstance, err := loadConfigFile(configPath)
	if err != nil {
		return err
	}

	// Apply the loaded configuration
	p.UpdateConfig(func(cfg Config) Config {
		// Process rules in the correct order:
		// 1. Process enabled rules
		enabledRules, haveEnabledRules := processEnabledRules(koanfInstance)

		if haveEnabledRules {
			cfg = cfg.WithRules(cfg.Rules.WithEnabledRules(enabledRules))
		}

		// 3. Process disabled rules
		disabledRules := processDisabledRules(koanfInstance)
		if len(disabledRules) > 0 {
			cfg = cfg.WithRules(cfg.Rules.WithDisabledRules(disabledRules))
		}

		// 4. Resolve conflicts between enabled and disabled rules
		if haveEnabledRules {
			cfg = resolveRuleConflicts(enabledRules, cfg)
		}

		return cfg
	})

	return nil
}

// Save saves the configuration to the specified path.
// Currently a stub implementation.
func (p *Provider) Save(_ string) error {
	// TODO: Implement saving to a specific path
	return nil
}

// DefaultConfig creates a new Config with default values.
// This is an alias for NewDefaultConfig for backward compatibility.
func DefaultConfig() Config {
	return NewDefaultConfig()
}

// NewDefaultConfig creates a new Config with default values.
func NewDefaultConfig() Config {
	return Config{
		Subject: SubjectConfig{
			Case:               "sentence",
			MaxLength:          72,
			RequireImperative:  false,
			DisallowedSuffixes: []string{"."},
		},
		Body: BodyConfig{
			Required:         true,
			MinLength:        10,
			MinimumLines:     3,
			AllowSignOffOnly: false,
		},
		Conventional: ConventionalConfig{
			Required:             true,
			RequireScope:         false,
			Types:                []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"},
			AllowBreakingChanges: true,
			MaxDescriptionLength: 72,
		},
		Rules: RulesConfig{
			EnabledRules: []string{
				"SubjectLength",
				"Body",
				"Conventional",
				"Imperative",
				"SubjectCase",
				"SubjectSuffix",
				// "JiraReference", // Disabled by default
				// "CommitBody", // Disabled by default
				"SignOff",
				"Signature",
				"Spell",
				"CommitsAhead",
				"SignedIdentity",
			},
			DisabledRules: []string{},
		},
		Security: SecurityConfig{
			SignOffRequired:       true,
			GPGRequired:           false,
			KeyDirectory:          "",
			AllowedSignatureTypes: []string{"GPG", "SSH"},
			AllowedKeyrings:       []string{},
			AllowedIdentities:     []string{},
			AllowMultipleSignOffs: true,
		},
		Repository: RepositoryConfig{
			Path:               "",
			ReferenceBranch:    "main",
			MaxCommitsAhead:    5,
			MaxHistoryDays:     365,
			OutputFormat:       "text",
			IgnoreMergeCommits: false,
		},
		Output: OutputConfig{
			Format:  "text",
			Verbose: false,
			Quiet:   false,
			Color:   true,
		},
		SpellCheck: SpellCheckConfig{
			Enabled:          false,
			Language:         "en-US",
			IgnoreCase:       false,
			CustomDictionary: nil,
		},
		Jira: JiraConfig{
			Pattern:  "[A-Z]+-\\d+",
			Projects: []string{},
			BodyRef:  false,
		},
	}
}
