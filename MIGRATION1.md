# Configuration System Simplification Plan

This document outlines a plan to simplify the configuration system in Gommitlint using functional programming principles with value semantics.

## Current Challenges

1. **Interface Proliferation**: Many small interfaces (SubjectConfigProvider, BodyConfigProvider, etc.) that fragment the configuration concerns
2. **Adapter Complexity**: ValidationConfigAdapter needs to implement numerous interfaces
3. **Indirection Overhead**: Multiple adapter layers between raw config and domain code
4. **Cognitive Load**: Developers need to understand many small interfaces and their relationships
5. **Configuration Discovery**: Difficult to discover all available configuration options

## Proposed Solution: Immutable Unified Configuration with Value Semantics

### 1. Create an Immutable Configuration Structure

Define a comprehensive immutable structure that contains all configuration options in typed, documented fields:

```go
// Config contains all configuration options in a single immutable structure
// organized by logical domains.
type Config struct {
    // Subject configuration options
    Subject SubjectConfig
    
    // Body configuration options
    Body BodyConfig
    
    // Conventional commit options
    Conventional ConventionalConfig
    
    // Security options
    Security SecurityConfig
    
    // Rule activation options
    Rules RuleConfig
    
    // Other domains...
}

// SubjectConfig contains subject-related configuration.
type SubjectConfig struct {
    MaxLength int
    Case      string
    // Other subject fields...
}

// BodyConfig contains body-related configuration.
type BodyConfig struct {
    Required        bool
    AllowSignOffOnly bool
    // Other body fields...
}

// ConventionalConfig contains conventional commit configuration.
type ConventionalConfig struct {
    Required      bool
    Types         []string
    Scopes        []string
    MaxDescLength int
    // Other conventional fields...
}

// Other configuration domains...
```

### 2. Pure Functions for Configuration Access

Provide accessor functions that return values without side effects:

```go
// SubjectMaxLength returns the maximum subject length.
func (c Config) SubjectMaxLength() int {
    if c.Subject.MaxLength <= 0 {
        return DefaultSubjectMaxLength
    }
    return c.Subject.MaxLength
}

// SubjectCase returns the subject case style.
func (c Config) SubjectCase() string {
    if c.Subject.Case == "" {
        return DefaultSubjectCase
    }
    return c.Subject.Case
}

// BodyRequired returns whether a commit body is required.
func (c Config) BodyRequired() bool {
    return c.Body.Required
}

// Other accessor methods...
```

### 3. Transformation Functions for Configuration Changes

Implement transformation functions that return new configurations:

```go
// WithSubjectMaxLength returns a new Config with the specified subject max length.
func (c Config) WithSubjectMaxLength(maxLength int) Config {
    result := c // Create a copy
    result.Subject.MaxLength = maxLength
    return result
}

// WithBodyRequired returns a new Config with the body required setting.
func (c Config) WithBodyRequired(required bool) Config {
    result := c // Create a copy
    result.Body.Required = required
    return result
}

// WithConventionalTypes returns a new Config with the specified conventional types.
func (c Config) WithConventionalTypes(types []string) Config {
    result := c // Create a copy
    
    // Create a deep copy of the slice to maintain immutability
    result.Conventional.Types = make([]string, len(types))
    copy(result.Conventional.Types, types)
    
    return result
}

// Other transformation functions...
```

### 4. Pure Validation Functions

Implement pure validation functions that don't modify state:

```go
// Validate returns validation errors for the configuration.
func (c Config) Validate() []error {
    var errors []error
    
    // Validate subject configuration
    if c.Subject.MaxLength < 0 {
        errors = append(errors, fmt.Errorf("subject max length cannot be negative"))
    }
    
    // Validate body configuration
    // ...
    
    // Validate conventional configuration
    // ...
    
    return errors
}

// IsValid returns whether the configuration is valid.
func (c Config) IsValid() bool {
    return len(c.Validate()) == 0
}
```

### 5. Functional Builder Pattern for Configuration

Implement a functional builder pattern with value semantics:

```go
// ConfigOption is a function that transforms a Config.
type ConfigOption func(Config) Config

// WithSubjectMaxLength creates an option to set the subject max length.
func WithSubjectMaxLength(maxLength int) ConfigOption {
    return func(c Config) Config {
        return c.WithSubjectMaxLength(maxLength)
    }
}

// WithBodyRequired creates an option to set the body required flag.
func WithBodyRequired(required bool) ConfigOption {
    return func(c Config) Config {
        return c.WithBodyRequired(required)
    }
}

// Other option functions...

// NewConfig creates a new Config with the default values.
func NewConfig(options ...ConfigOption) Config {
    config := DefaultConfig()
    
    // Apply all options in sequence
    for _, option := range options {
        config = option(config)
    }
    
    return config
}

// Usage:
// config := NewConfig(
//     WithSubjectMaxLength(80),
//     WithBodyRequired(true),
//     WithConventionalTypes([]string{"feat", "fix", "docs"}),
// )
```

### 6. Serialization and Deserialization

Pure functions for serialization and deserialization:

```go
// FromYAML creates a Config from YAML data.
func FromYAML(data []byte) (Config, error) {
    var raw map[string]interface{}
    if err := yaml.Unmarshal(data, &raw); err != nil {
        return Config{}, fmt.Errorf("invalid YAML: %w", err)
    }
    
    return fromMap(raw)
}

// ToYAML converts a Config to YAML data.
func (c Config) ToYAML() ([]byte, error) {
    return yaml.Marshal(toMap(c))
}

// fromMap creates a Config from a map representation.
func fromMap(data map[string]interface{}) (Config, error) {
    config := DefaultConfig()
    
    // Process subject configuration
    if subjectMap, ok := data["subject"].(map[string]interface{}); ok {
        if maxLength, ok := subjectMap["max_length"].(int); ok {
            config = config.WithSubjectMaxLength(maxLength)
        }
        // Process other subject fields...
    }
    
    // Process other configuration sections...
    
    return config, nil
}

// toMap converts a Config to a map representation.
func toMap(config Config) map[string]interface{} {
    result := make(map[string]interface{})
    
    // Convert subject configuration
    subject := make(map[string]interface{})
    subject["max_length"] = config.Subject.MaxLength
    subject["case"] = config.Subject.Case
    // Convert other subject fields...
    
    result["subject"] = subject
    
    // Convert other configuration sections...
    
    return result
}
```

## Benefits of this Approach

1. **Simplicity**: Single configuration structure with clear organization
2. **Discoverability**: All options visible in one place with documentation
3. **Type Safety**: Strong typing ensures configuration is used correctly
4. **Centralized Validation**: Validation happens in one place
5. **Easier Testing**: Builder pattern makes test setup simpler
6. **Forward Compatibility**: Easy to add new configuration options
7. **Reduced Indirection**: Fewer layers between configuration and usage

## Migration Strategy

1. Create the new unified configuration structure
2. Implement the new Config interface and ConfigImpl
3. Add adapters to translate from old interfaces to new ones
4. Gradually update rule constructors to use the new configuration
5. Eventually remove the old interfaces when all code is migrated

## Implementation Phases

### Phase 1: Core Configuration Structure

- [ ] Create UnifiedConfig structure with all configuration options
- [ ] Define Config interface with accessor methods
- [ ] Implement ConfigImpl with proper defaults
- [ ] Add validation logic for the configuration
- [ ] Create configuration loading functionality

### Phase 2: Builder and Test Support

- [ ] Implement ConfigBuilder for test configuration
- [ ] Update test helpers to use the new configuration
- [ ] Create adapters to translate between old and new interfaces
- [ ] Add documentation for the new configuration system

### Phase 3: Rule Migration

- [ ] Update one rule at a time to use the new configuration system
- [ ] Add proper tests for each migrated rule
- [ ] Ensure backward compatibility during the transition
- [ ] Update documentation for each migrated rule

### Phase 4: Completion and Cleanup

- [ ] Remove old configuration interfaces once all code is migrated
- [ ] Update all documentation to reflect the new system
- [ ] Perform final testing to ensure correctness
- [ ] Add examples and guides for the new configuration system