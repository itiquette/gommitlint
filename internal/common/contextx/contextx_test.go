// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package contextx_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/common/contextkeys"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

func TestWithValue(t *testing.T) {
	// Setup
	ctx := testcontext.CreateTestContext()
	key := contextkeys.ContextKey("test-key")
	value := "test-value"

	// Test adding value to context
	ctxWithValue := contextx.WithValue(ctx, key, value)

	// Verify
	retrievedValue, ok := ctxWithValue.Value(key).(string)
	require.True(t, ok, "Value should be retrievable and castable to string")
	require.Equal(t, value, retrievedValue, "Retrieved value should match stored value")
}

func TestValue(t *testing.T) {
	// Setup
	ctx := testcontext.CreateTestContext()
	key := contextkeys.ContextKey("test-key")
	value := "test-value"

	// Add value to context
	ctxWithValue := context.WithValue(ctx, key, value)

	// Retrieve with type assertion
	retrievedValue, exists := contextx.Value[string](ctxWithValue, key)
	require.True(t, exists, "Value retrieval should succeed")
	require.Equal(t, value, retrievedValue, "Retrieved value should match stored value")

	// Test retrieval of wrong type
	_, exists = contextx.Value[int](ctxWithValue, key)
	require.False(t, exists, "Value retrieval with wrong type should fail")

	// Test retrieval of non-existent key
	_, exists = contextx.Value[string](ctx, key)
	require.False(t, exists, "Value retrieval with non-existent key should fail")
}

func TestMergeContext(t *testing.T) {
	// Setup - Use only the predefined context keys
	ctx1 := testcontext.CreateTestContext()
	ctx1 = context.WithValue(ctx1, contextkeys.LoggerKey, "logger1")
	ctx1 = context.WithValue(ctx1, contextkeys.CLIOptionsKey, "options1")

	ctx2 := testcontext.CreateTestContext()
	ctx2 = context.WithValue(ctx2, contextkeys.LoggerKey, "logger2")
	// Don't set CLIOptionsKey in ctx2 to test merging

	// Merge contexts
	merged := contextx.MergeContext(ctx1, ctx2)

	// Verify values from both contexts are present, with ctx2 taking precedence for common keys
	require.Equal(t, "logger2", merged.Value(contextkeys.LoggerKey))
	require.Equal(t, "options1", merged.Value(contextkeys.CLIOptionsKey))

	// Test edge cases - verify values rather than context objects
	mergedWithEmpty := contextx.MergeContext(testcontext.CreateTestContext(), ctx2)
	require.Equal(t, "logger2", mergedWithEmpty.Value(contextkeys.LoggerKey))
	require.Nil(t, mergedWithEmpty.Value(contextkeys.CLIOptionsKey))

	mergedWithNil := contextx.MergeContext(ctx1, nil)
	require.Equal(t, "logger1", mergedWithNil.Value(contextkeys.LoggerKey))
	require.Equal(t, "options1", mergedWithNil.Value(contextkeys.CLIOptionsKey))
}

func TestLoggerFunctions(t *testing.T) {
	// Setup
	ctx := testcontext.CreateTestContext()
	// Use a simple mock logger for testing
	log := &testLogger{}

	// Add logger to context
	ctxWithLogger := contextx.WithLogger(ctx, log)

	// Retrieve logger
	retrievedLogger := contextx.GetLogger(ctxWithLogger)
	require.NotNil(t, retrievedLogger, "Logger should be retrievable")
}

type testLogger struct{}

func (l *testLogger) Debug(_ string, _ ...interface{}) {}
func (l *testLogger) Info(_ string, _ ...interface{})  {}
func (l *testLogger) Warn(_ string, _ ...interface{})  {}
func (l *testLogger) Error(_ string, _ ...interface{}) {}
