// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package functional

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResult_Ok(t *testing.T) {
	result := Ok(42)

	require.True(t, result.IsOk())
	require.False(t, result.IsErr())

	value, err := result.Unwrap()
	require.NoError(t, err)
	require.Equal(t, 42, value)
}

func TestResult_Err(t *testing.T) {
	testErr := errors.New("test error")
	result := Err[int](testErr)

	require.False(t, result.IsOk())
	require.True(t, result.IsErr())

	value, err := result.Unwrap()
	require.Error(t, err)
	require.Equal(t, testErr, err)
	require.Equal(t, 0, value) // zero value for int
}

func TestResult_UnwrapOr(t *testing.T) {
	tests := []struct {
		name         string
		result       Result[int]
		defaultValue int
		expected     int
	}{
		{
			name:         "Ok result returns value",
			result:       Ok(42),
			defaultValue: 0,
			expected:     42,
		},
		{
			name:         "Err result returns default",
			result:       Err[int](errors.New("error")),
			defaultValue: 99,
			expected:     99,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			actual := testCase.result.UnwrapOr(testCase.defaultValue)
			require.Equal(t, testCase.expected, actual)
		})
	}
}

func TestMap(t *testing.T) {
	tests := []struct {
		name     string
		input    Result[int]
		fn       func(int) string
		expected Result[string]
	}{
		{
			name:     "Ok result maps successfully",
			input:    Ok(42),
			fn:       func(i int) string { return fmt.Sprintf("value: %d", i) },
			expected: Ok("value: 42"),
		},
		{
			name:     "Err result propagates error",
			input:    Err[int](errors.New("test error")),
			fn:       func(i int) string { return fmt.Sprintf("value: %d", i) },
			expected: Err[string](errors.New("test error")),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Map(testCase.input, testCase.fn)

			require.Equal(t, testCase.expected.IsOk(), result.IsOk())

			if testCase.expected.IsOk() {
				expectedValue, _ := testCase.expected.Unwrap()
				actualValue, _ := result.Unwrap()
				require.Equal(t, expectedValue, actualValue)
			} else {
				_, expectedErr := testCase.expected.Unwrap()
				_, actualErr := result.Unwrap()
				require.Equal(t, expectedErr.Error(), actualErr.Error())
			}
		})
	}
}

func TestFlatMap(t *testing.T) {
	tests := []struct {
		name     string
		input    Result[int]
		fn       func(int) Result[string]
		expected Result[string]
	}{
		{
			name:  "Ok result with successful function",
			input: Ok(42),
			fn: func(i int) Result[string] {
				return Ok(fmt.Sprintf("value: %d", i))
			},
			expected: Ok("value: 42"),
		},
		{
			name:  "Ok result with failing function",
			input: Ok(42),
			fn: func(_ int) Result[string] {
				return Err[string](errors.New("function error"))
			},
			expected: Err[string](errors.New("function error")),
		},
		{
			name:  "Err result propagates error",
			input: Err[int](errors.New("input error")),
			fn: func(i int) Result[string] {
				return Ok(fmt.Sprintf("value: %d", i))
			},
			expected: Err[string](errors.New("input error")),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := FlatMap(testCase.input, testCase.fn)

			require.Equal(t, testCase.expected.IsOk(), result.IsOk())

			if testCase.expected.IsOk() {
				expectedValue, _ := testCase.expected.Unwrap()
				actualValue, _ := result.Unwrap()
				require.Equal(t, expectedValue, actualValue)
			} else {
				_, expectedErr := testCase.expected.Unwrap()
				_, actualErr := result.Unwrap()
				require.Equal(t, expectedErr.Error(), actualErr.Error())
			}
		})
	}
}

func TestMapErr(t *testing.T) {
	tests := []struct {
		name     string
		input    Result[int]
		fn       func(error) error
		expected Result[int]
	}{
		{
			name:     "Ok result unchanged",
			input:    Ok(42),
			fn:       func(err error) error { return fmt.Errorf("wrapped: %w", err) },
			expected: Ok(42),
		},
		{
			name:     "Err result transforms error",
			input:    Err[int](errors.New("original error")),
			fn:       func(err error) error { return fmt.Errorf("wrapped: %w", err) },
			expected: Err[int](fmt.Errorf("wrapped: %w", errors.New("original error"))),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := MapErr(testCase.input, testCase.fn)

			require.Equal(t, testCase.expected.IsOk(), result.IsOk())

			if testCase.expected.IsOk() {
				expectedValue, _ := testCase.expected.Unwrap()
				actualValue, _ := result.Unwrap()
				require.Equal(t, expectedValue, actualValue)
			} else {
				_, expectedErr := testCase.expected.Unwrap()
				_, actualErr := result.Unwrap()
				require.Equal(t, expectedErr.Error(), actualErr.Error())
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name      string
		input     Result[int]
		predicate func(int) bool
		errorMsg  string
		wantOk    bool
	}{
		{
			name:      "Ok result passes filter",
			input:     Ok(42),
			predicate: func(i int) bool { return i > 0 },
			errorMsg:  "value must be positive",
			wantOk:    true,
		},
		{
			name:      "Ok result fails filter",
			input:     Ok(-1),
			predicate: func(i int) bool { return i > 0 },
			errorMsg:  "value must be positive",
			wantOk:    false,
		},
		{
			name:      "Err result propagates error",
			input:     Err[int](errors.New("original error")),
			predicate: func(i int) bool { return i > 0 },
			errorMsg:  "value must be positive",
			wantOk:    false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Filter(testCase.input, testCase.predicate, testCase.errorMsg)
			require.Equal(t, testCase.wantOk, result.IsOk())
		})
	}
}

func TestFromGoResult(t *testing.T) {
	tests := []struct {
		name   string
		value  int
		err    error
		wantOk bool
	}{
		{
			name:   "Success case",
			value:  42,
			err:    nil,
			wantOk: true,
		},
		{
			name:   "Error case",
			value:  0,
			err:    errors.New("test error"),
			wantOk: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := FromGoResult(testCase.value, testCase.err)
			require.Equal(t, testCase.wantOk, result.IsOk())

			if testCase.wantOk {
				value, err := result.Unwrap()
				require.NoError(t, err)
				require.Equal(t, testCase.value, value)
			} else {
				_, err := result.Unwrap()
				require.Error(t, err)
				require.Equal(t, testCase.err.Error(), err.Error())
			}
		})
	}
}

func TestCombine(t *testing.T) {
	tests := []struct {
		name   string
		r1     Result[int]
		r2     Result[string]
		fn     func(int, string) string
		wantOk bool
	}{
		{
			name:   "Both Ok",
			r1:     Ok(42),
			r2:     Ok("hello"),
			fn:     func(i int, s string) string { return fmt.Sprintf("%s: %d", s, i) },
			wantOk: true,
		},
		{
			name:   "First Err",
			r1:     Err[int](errors.New("first error")),
			r2:     Ok("hello"),
			fn:     func(i int, s string) string { return fmt.Sprintf("%s: %d", s, i) },
			wantOk: false,
		},
		{
			name:   "Second Err",
			r1:     Ok(42),
			r2:     Err[string](errors.New("second error")),
			fn:     func(i int, s string) string { return fmt.Sprintf("%s: %d", s, i) },
			wantOk: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Combine(testCase.r1, testCase.r2, testCase.fn)
			require.Equal(t, testCase.wantOk, result.IsOk())
		})
	}
}

// Example of chaining operations.
func ExampleResult_chaining() {
	// Parse a string to int, multiply by 2, and format as string
	parseAndDouble := func(s string) Result[string] {
		// Simulate parsing
		if s == "invalid" {
			return Err[string](errors.New("invalid input"))
		}

		result := Ok(42) // Simulated parse result

		// Chain operations using Map and FlatMap
		doubled := Map(result, func(i int) int { return i * 2 })
		formatted := Map(doubled, func(i int) string { return fmt.Sprintf("result: %d", i) })

		return formatted
	}

	// Usage
	result := parseAndDouble("42")
	value, _ := result.Unwrap()
	fmt.Println(value)
	// Output: result: 84
}
