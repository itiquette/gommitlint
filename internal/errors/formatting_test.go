// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package errors

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithHelp(t *testing.T) {
	err := New("TestRule", ErrInvalidFormat, "Test message")

	// Enhance with help
	enhanced := WithHelp(err, "This is a help message")

	// Check that enhanced has help in the Help field
	require.Equal(t, "This is a help message", enhanced.Help)
}

func TestGetHelp(t *testing.T) {
	// Error without help
	err1 := New("TestRule", ErrInvalidFormat, "Test message")
	require.Equal(t, "", err1.GetHelp())

	// Error with help
	err2 := err1.WithHelp("This is a help message")
	require.Equal(t, "This is a help message", err2.GetHelp())
}

func TestFormatAsText(t *testing.T) {
	err := New("TestRule", ErrInvalidFormat, "Test message")
	err = err.WithContext("key1", "value1")
	err = err.WithHelp("This is a help message")

	// Test non-verbose format
	text := err.FormatAsText(false)
	require.Equal(t, "[TestRule] Test message", text)

	// Test verbose format
	verbose := err.FormatAsText(true)
	require.Contains(t, verbose, "Rule:    TestRule")
	require.Contains(t, verbose, "Code:    invalid_format")
	require.Contains(t, verbose, "Message: Test message")
	require.Contains(t, verbose, "Help:    This is a help message")
	require.Contains(t, verbose, "Context:")
	require.Contains(t, verbose, "  key1: value1")

	// Help should not be duplicated in context
	require.NotContains(t, verbose, "  help: This is a help message")
}

func TestFormatAsJSON(t *testing.T) {
	err := New("TestRule", ErrInvalidFormat, "Test message")
	err = err.WithContext("key1", "value1")
	err = err.WithHelp("This is a help message")

	// Format as JSON
	data, jsonErr := err.FormatAsJSON()
	require.NoError(t, jsonErr)

	// Parse back
	var parsed map[string]interface{}

	require.NoError(t, json.Unmarshal(data, &parsed))

	// Check fields
	require.Equal(t, "TestRule", parsed["rule"])
	require.Equal(t, "invalid_format", parsed["code"])
	require.Equal(t, "Test message", parsed["message"])
	require.Equal(t, "This is a help message", parsed["help"])

	// Check context - should not include help
	context, ok := parsed["context"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "value1", context["key1"])
	_, hasHelp := context["help"]
	require.False(t, hasHelp)
}

func TestFormatAsMarkdown(t *testing.T) {
	err := New("TestRule", ErrInvalidFormat, "Test message")
	err = err.WithContext("key1", "value1")
	err = err.WithHelp("This is a help message")

	// Format as Markdown
	markdown := err.FormatAsMarkdown()

	// Check content
	require.Contains(t, markdown, "### TestRule: Test message")
	require.Contains(t, markdown, "**Code:** `invalid_format`")
	require.Contains(t, markdown, "**Help:** This is a help message")
	require.Contains(t, markdown, "**Context:**")
	require.Contains(t, markdown, "- `key1`: value1")

	// Help should not be duplicated in context
	require.NotContains(t, markdown, "- `help`: This is a help message")
}

func TestNewFormatValidationError(t *testing.T) {
	// Test the simplified NewFormatValidationError helper
	err := NewFormatValidationError("TestRule", "Test error message", "This is help text", "Subject line")

	require.Equal(t, "TestRule", err.Rule)
	require.Equal(t, "invalid_format", err.Code)
	require.Equal(t, "Test error message", err.Message)
	require.Equal(t, "This is help text", err.GetHelp())
	require.Equal(t, "Subject line", err.Context["subject"])
}

func TestFormatterInterface(t *testing.T) {
	err := New("TestRule", ErrInvalidFormat, "Test message")
	err = err.WithContext("key1", "value1")
	err = err.WithHelp("This is a help message")

	// Test text formatter
	textFormatter := NewTextFormatter(true)
	textOutput := textFormatter.FormatError(err)
	require.Contains(t, textOutput, "Rule:    TestRule")
	require.Contains(t, textOutput, "Help:    This is a help message")

	// Test JSON formatter
	jsonFormatter := NewJSONFormatter()
	jsonOutput := jsonFormatter.FormatError(err)
	require.Contains(t, jsonOutput, `"rule":"TestRule"`)
	require.Contains(t, jsonOutput, `"help":"This is a help message"`)

	// Test Markdown formatter
	mdFormatter := NewMarkdownFormatter()
	mdOutput := mdFormatter.FormatError(err)
	require.Contains(t, mdOutput, "### TestRule: Test message")
	require.Contains(t, mdOutput, "**Help:** This is a help message")
}
