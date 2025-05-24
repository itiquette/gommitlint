// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package report_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/itiquette/gommitlint/internal/application/report"
	"github.com/itiquette/gommitlint/internal/domain"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

// mockFormatter implements the outgoing.ResultFormatter interface for testing.
type mockFormatter struct {
	formatFunc func(ctx context.Context, results interface{}) string
}

func (m mockFormatter) Format(ctx context.Context, results interface{}) string {
	if m.formatFunc != nil {
		return m.formatFunc(ctx, results)
	}

	return "Mock formatted output"
}

// mockLogger implements the outgoing.Logger interface for testing.
type mockLogger struct{}

func (m mockLogger) Debug(_ string, _ ...interface{}) {}
func (m mockLogger) Info(_ string, _ ...interface{})  {}
func (m mockLogger) Warn(_ string, _ ...interface{})  {}
func (m mockLogger) Error(_ string, _ ...interface{}) {}

// TestGeneratorValueSemantics verifies that the Generator correctly implements
// value semantics and functional programming patterns.
func TestGeneratorValueSemantics(t *testing.T) {
	// Create test dependencies
	buf := &bytes.Buffer{}
	formatter := mockFormatter{}

	// Create options
	options := report.Options{
		Format:    report.FormatText,
		Verbose:   false,
		ShowHelp:  false,
		LightMode: false,
		Writer:    buf,
	}

	// Create a generator
	logger := mockLogger{}
	generator := report.NewGenerator(options, formatter, logger)

	// Test WithVerbose
	t.Run("WithVerbose maintains immutability", func(t *testing.T) {
		// Get a new generator with verbose enabled
		verboseGen, ok := generator.WithVerbose(true).(report.Generator)
		require.True(t, ok, "Should be able to convert to Generator type")

		verboseGenerator := verboseGen

		// Verify the original is unchanged
		require.False(t, generator.Options().Verbose, "Original generator should not be modified")

		// Verify the new generator has the change
		require.True(t, verboseGenerator.Options().Verbose, "New generator should have verbose enabled")

		// Verify they are different instances - using pointer to struct is not appropriate for value types
		require.NotEqual(t, generator, verboseGenerator, "Should be different values")
	})

	// Test WithShowHelp
	t.Run("WithShowHelp maintains immutability", func(t *testing.T) {
		// Get a new generator with show help enabled
		helpGen, ok := generator.WithShowHelp(true).(report.Generator)
		require.True(t, ok, "Should be able to convert to Generator type")

		helpGenerator := helpGen

		// Verify the original is unchanged
		require.False(t, generator.Options().ShowHelp, "Original generator should not be modified")

		// Verify the new generator has the change
		require.True(t, helpGenerator.Options().ShowHelp, "New generator should have show help enabled")

		// Verify they are different instances
		require.NotEqual(t, generator, helpGenerator, "Should be different values")
	})

	// Test WithRuleToShowHelp
	t.Run("WithRuleToShowHelp maintains immutability", func(t *testing.T) {
		ruleName := "TestRule"

		// Get a new generator with rule to show help
		ruleHelpGen, ok := generator.WithRuleToShowHelp(ruleName).(report.Generator)
		require.True(t, ok, "Should be able to convert to Generator type")

		ruleHelpGenerator := ruleHelpGen

		// Verify the original is unchanged
		require.Empty(t, generator.Options().RuleToShowHelp, "Original generator should not be modified")

		// Verify the new generator has the change
		require.Equal(t, ruleName, ruleHelpGenerator.Options().RuleToShowHelp,
			"New generator should have rule to show help set")

		// Verify they are different instances
		require.NotEqual(t, generator, ruleHelpGenerator, "Should be different values")
	})

	// Test chaining transformations
	t.Run("Chaining maintains immutability", func(t *testing.T) {
		// Chain multiple transformations
		transformedGen := generator.
			WithVerbose(true).
			WithShowHelp(true).
			WithRuleToShowHelp("TestRule")
		transformedGenerator, ok := transformedGen.(report.Generator)
		require.True(t, ok, "Should be able to convert to Generator type")

		// Verify the original is unchanged
		require.False(t, generator.Options().Verbose, "Original generator should not be modified")
		require.False(t, generator.Options().ShowHelp, "Original generator should not be modified")
		require.Empty(t, generator.Options().RuleToShowHelp, "Original generator should not be modified")

		// Verify the new generator has all changes
		require.True(t, transformedGenerator.Options().Verbose, "Transformed generator should have verbose enabled")
		require.True(t, transformedGenerator.Options().ShowHelp, "Transformed generator should have show help enabled")
		require.Equal(t, "TestRule", transformedGenerator.Options().RuleToShowHelp,
			"Transformed generator should have rule to show help set")
	})
}

// TestGenerateReport verifies that the Generator's GenerateReport method works correctly.
func TestGenerateReport(t *testing.T) {
	// Create validation results
	results := domain.ValidationResults{
		PassedCommits: 1,
		TotalCommits:  1,
	}

	// Test normal report generation
	t.Run("Generate successful report", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := mockFormatter{
			formatFunc: func(_ context.Context, _ interface{}) string {
				return "Test formatted output for successful validation"
			},
		}

		generator := report.NewGenerator(report.Options{
			Writer: buf,
		}, formatter, mockLogger{})

		err := generator.GenerateReport(testcontext.CreateTestContext(), results)
		require.NoError(t, err, "GenerateReport should not return an error for successful validation")
		require.Contains(t, buf.String(), "Test formatted output", "Output should contain formatted text")
	})

	// Test failed validation report
	t.Run("Generate failed report", func(t *testing.T) {
		// We need to use os.Stdout as the writer since the error is only returned
		// when writing to stdout
		originalStdout := os.Stdout
		reader, writer, _ := os.Pipe()
		os.Stdout = writer

		formatter := mockFormatter{
			formatFunc: func(_ context.Context, _ interface{}) string {
				return "Test formatted output for failed validation"
			},
		}

		// Create results with a failure
		failedResults := domain.ValidationResults{
			PassedCommits: 0,
			TotalCommits:  1,
		}

		generator := report.NewGenerator(report.Options{
			Writer: os.Stdout, // Must use os.Stdout to trigger the error condition
		}, formatter, mockLogger{})

		err := generator.GenerateReport(testcontext.CreateTestContext(), failedResults)
		require.Error(t, err, "GenerateReport should return an error for failed validation")
		require.Contains(t, err.Error(), "validation failed", "Error should indicate validation failed")

		// Restore stdout
		writer.Close()

		var buf bytes.Buffer

		_, err = io.Copy(&buf, reader)
		require.NoError(t, err, "Failed to copy from reader")

		os.Stdout = originalStdout

		require.Contains(t, buf.String(), "Test formatted output", "Output should still be written")
	})
}

// TestGenerateSummary verifies that the Generator's GenerateSummary method works correctly.
func TestGenerateSummary(t *testing.T) {
	// Test successful summary
	t.Run("Generate successful summary", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := mockFormatter{}

		results := domain.ValidationResults{
			PassedCommits: 1,
			TotalCommits:  1,
		}

		generator := report.NewGenerator(report.Options{
			Writer: buf,
		}, formatter, mockLogger{})

		err := generator.GenerateSummary(testcontext.CreateTestContext(), results)
		require.NoError(t, err, "GenerateSummary should not return an error for successful validation")
		require.Contains(t, buf.String(), "All commits passed", "Output should indicate success")
	})

	// Test failed summary
	t.Run("Generate failed summary", func(t *testing.T) {
		// We need to use os.Stdout as the writer since the error is only returned
		// when writing to stdout
		originalStdout := os.Stdout
		reader, writer, _ := os.Pipe()
		os.Stdout = writer

		formatter := mockFormatter{}

		// Create results with a failure
		failedResults := domain.ValidationResults{
			PassedCommits: 0,
			TotalCommits:  1,
		}

		generator := report.NewGenerator(report.Options{
			Writer: os.Stdout, // Must use os.Stdout to trigger the error condition
		}, formatter, mockLogger{})

		err := generator.GenerateSummary(testcontext.CreateTestContext(), failedResults)
		require.Error(t, err, "GenerateSummary should return an error for failed validation")
		require.Contains(t, err.Error(), "validation failed", "Error should indicate validation failed")

		// Restore stdout
		writer.Close()

		var buf bytes.Buffer

		_, err = io.Copy(&buf, reader)
		require.NoError(t, err, "Failed to copy from reader")

		os.Stdout = originalStdout

		require.Contains(t, buf.String(), "Some commits failed", "Output should indicate failure")
	})
}
