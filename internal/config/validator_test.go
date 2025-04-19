package config_test

import (
	"os"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test validation of configuration values directly
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		expectErr   bool
	}{
		{
			name: "nil gommitconf",
			yamlContent: `# Empty config without gommitlint key
some_other_key: true
`,
			expectErr: true,
		},
		{
			name: "invalid subject length",
			yamlContent: `gommitlint:
  subject:
    max-length: 0
    case: lower
  body:
    required: false
`,
			expectErr: true,
		},
		{
			name: "invalid subject case",
			yamlContent: `gommitlint:
  subject:
    max-length: 50
    case: invalid_case
  body:
    required: false
`,
			expectErr: true,
		},
		{
			name: "invalid jira configuration",
			yamlContent: `gommitlint:
  subject:
    max-length: 50
    case: lower
    jira:
      required: true
      pattern: ""
  body:
    required: false
`,
			expectErr: true,
		},
		{
			name: "invalid conventional commit",
			yamlContent: `gommitlint:
  subject:
    max-length: 50
    case: lower
  body:
    required: false
  conventional-commit:
    required: true
    types: []
    max-description-length: 0
`,
			expectErr: true,
		},
		{
			name: "valid configuration",
			yamlContent: `gommitlint:
  subject:
    max-length: 50
    case: lower
  body:
    required: false
  conventional-commit:
    required: false
`,
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new manager for each test
			manager, err := config.New()
			require.NoError(t, err, "Manager creation should not fail")

			// Create a temporary file with the test content
			tmpFile, err := os.CreateTemp("", "config-*.yaml")
			require.NoError(t, err, "Failed to create temp file")
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tc.yamlContent)
			require.NoError(t, err, "Failed to write to temp file")
			tmpFile.Close()

			// Try to load the file
			err = manager.LoadFromFile(tmpFile.Name())

			if tc.expectErr {
				assert.Error(t, err, "Expected validation error for test case: %s", tc.name)
			} else {
				assert.NoError(t, err, "Expected no validation errors for test case: %s", tc.name)
			}
		})
	}
}

// Test validation through the LoadFromFile method with various configuration issues
func TestValidationThroughLoadFromFile(t *testing.T) {
	// Each test gets its own manager to avoid interference
	tests := []struct {
		name         string
		fileContent  string
		expectErr    bool
		errorMessage string
	}{
		{
			name: "negative subject length",
			fileContent: `gommitlint:
  subject:
    max-length: -10
    case: lower
  body:
    required: false
`,
			expectErr:    true,
			errorMessage: "subject.max-length must be positive",
		},
		{
			name: "invalid subject case",
			fileContent: `gommitlint:
  subject:
    max-length: 50
    case: WRONG
  body:
    required: false
`,
			expectErr:    true,
			errorMessage: "subject.case must be one of",
		},
		{
			name: "empty types for required conventional commit",
			fileContent: `gommitlint:
  subject:
    max-length: 50
    case: lower
  body:
    required: false
  conventional-commit:
    required: true
    types: []
`,
			expectErr:    true,
			errorMessage: "conventional-commit.types cannot be empty",
		},
		{
			name: "valid configuration",
			fileContent: `gommitlint:
  subject:
    max-length: 50
    case: lower
  body:
    required: false
  conventional-commit:
    required: false
`,
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new manager
			manager, err := config.New()
			require.NoError(t, err, "Manager creation should not fail")

			// Create a temporary file with the test content
			tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
			require.NoError(t, err, "Failed to create temp file")
			defer func() {
				// Safely remove the file when done
				_ = os.Remove(tmpFile.Name())
			}()

			// Write the content and close the file
			_, err = tmpFile.WriteString(tc.fileContent)
			require.NoError(t, err, "Failed to write to temp file")
			tmpFile.Close()

			// Try to load the file
			err = manager.LoadFromFile(tmpFile.Name())

			// Check the outcome
			if tc.expectErr {
				if assert.Error(t, err, "LoadFromFile should return error for invalid configuration") {
					if tc.errorMessage != "" {
						assert.Contains(t, err.Error(), tc.errorMessage,
							"Error message should contain expected text")
					}
				}
			} else {
				assert.NoError(t, err, "LoadFromFile should not return error for valid configuration")
			}
		})
	}
}
