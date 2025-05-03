// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package contextx_test

import (
	"strconv"
	"testing"

	"github.com/itiquette/gommitlint/internal/contextx"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expected  []int
	}{
		{
			name:      "filter even numbers",
			input:     []int{1, 2, 3, 4, 5, 6},
			predicate: func(n int) bool { return n%2 == 0 },
			expected:  []int{2, 4, 6},
		},
		{
			name:      "filter positive numbers",
			input:     []int{-2, -1, 0, 1, 2},
			predicate: func(n int) bool { return n > 0 },
			expected:  []int{1, 2},
		},
		{
			name:      "empty input slice",
			input:     []int{},
			predicate: func(n int) bool { return n > 0 },
			expected:  []int{},
		},
		{
			name:      "nil input slice",
			input:     nil,
			predicate: func(n int) bool { return n > 0 },
			expected:  nil,
		},
		{
			name:      "filter all",
			input:     []int{1, 2, 3},
			predicate: func(_ int) bool { return false },
			expected:  []int{},
		},
		{
			name:      "keep all",
			input:     []int{1, 2, 3},
			predicate: func(_ int) bool { return true },
			expected:  []int{1, 2, 3},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := contextx.Filter(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expected, result)

			// Ensure the original slice wasn't modified
			if len(testCase.input) > 0 {
				originalCopy := make([]int, len(testCase.input))
				copy(originalCopy, testCase.input)
				require.Equal(t, originalCopy, testCase.input)
			}
		})
	}
}

func TestDeepCopy(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "copy string slice",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := contextx.DeepCopy(testCase.input)
			require.Equal(t, testCase.expected, result)

			// Check that it's a different slice with the same content
			if testCase.input != nil {
				require.NotSame(t, &testCase.input, &result)

				// Modify result and ensure original is unchanged
				if len(result) > 0 {
					result[0] = "modified"
					if len(testCase.input) > 0 {
						require.NotEqual(t, result[0], testCase.input[0])
					}
				}
			}
		})
	}
}

func TestMap(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		transform func(int) string
		expected  []string
	}{
		{
			name:      "int to string",
			input:     []int{1, 2, 3},
			transform: func(n int) string { return strconv.Itoa(n) },
			expected:  []string{"1", "2", "3"},
		},
		{
			name:      "multiply by 2",
			input:     []int{1, 2, 3},
			transform: func(n int) string { return strconv.Itoa(n * 2) },
			expected:  []string{"2", "4", "6"},
		},
		{
			name:      "empty slice",
			input:     []int{},
			transform: func(n int) string { return strconv.Itoa(n) },
			expected:  []string{},
		},
		{
			name:      "nil slice",
			input:     nil,
			transform: func(n int) string { return strconv.Itoa(n) },
			expected:  nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := contextx.Map(testCase.input, testCase.transform)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestReduce(t *testing.T) {
	tests := []struct {
		name         string
		input        []int
		initialValue int
		reducer      func(int, int) int
		expected     int
	}{
		{
			name:         "sum",
			input:        []int{1, 2, 3, 4, 5},
			initialValue: 0,
			reducer:      func(acc, n int) int { return acc + n },
			expected:     15,
		},
		{
			name:         "product",
			input:        []int{1, 2, 3, 4, 5},
			initialValue: 1,
			reducer:      func(acc, n int) int { return acc * n },
			expected:     120,
		},
		{
			name:         "max",
			input:        []int{3, 1, 7, 4, 2},
			initialValue: 0,
			reducer: func(acc, n int) int {
				if n > acc {
					return n
				}

				return acc
			},
			expected: 7,
		},
		{
			name:         "empty slice",
			input:        []int{},
			initialValue: 0,
			reducer:      func(acc, n int) int { return acc + n },
			expected:     0,
		},
		{
			name:         "nil slice",
			input:        nil,
			initialValue: 42,
			reducer:      func(acc, n int) int { return acc + n },
			expected:     42, // Initial value is returned
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := contextx.Reduce(testCase.input, testCase.initialValue, testCase.reducer)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestEvery(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expected  bool
	}{
		{
			name:      "all positive",
			input:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n > 0 },
			expected:  true,
		},
		{
			name:      "not all even",
			input:     []int{2, 4, 5, 6, 8},
			predicate: func(n int) bool { return n%2 == 0 },
			expected:  false,
		},
		{
			name:      "empty slice",
			input:     []int{},
			predicate: func(n int) bool { return n > 0 },
			expected:  true, // Vacuously true
		},
		{
			name:      "nil slice",
			input:     nil,
			predicate: func(n int) bool { return n > 0 },
			expected:  true, // Vacuously true
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := contextx.Every(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestSome(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expected  bool
	}{
		{
			name:      "has even",
			input:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n%2 == 0 },
			expected:  true,
		},
		{
			name:      "none negative",
			input:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n < 0 },
			expected:  false,
		},
		{
			name:      "empty slice",
			input:     []int{},
			predicate: func(n int) bool { return n > 0 },
			expected:  false, // Empty slice has no elements that satisfy
		},
		{
			name:      "nil slice",
			input:     nil,
			predicate: func(n int) bool { return n > 0 },
			expected:  false, // Nil slice has no elements
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := contextx.Some(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		value    string
		expected bool
	}{
		{
			name:     "contains value",
			input:    []string{"a", "b", "c"},
			value:    "b",
			expected: true,
		},
		{
			name:     "does not contain value",
			input:    []string{"a", "b", "c"},
			value:    "d",
			expected: false,
		},
		{
			name:     "empty slice",
			input:    []string{},
			value:    "a",
			expected: false,
		},
		{
			name:     "nil slice",
			input:    nil,
			value:    "a",
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := contextx.Contains(testCase.input, testCase.value)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestDeepCopyMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]int
	}{
		{
			name: "copy map with values",
			input: map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
			},
		},
		{
			name:  "empty map",
			input: map[string]int{},
		},
		{
			name:  "nil map",
			input: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := contextx.DeepCopyMap(testCase.input)

			// Check nil case
			if testCase.input == nil {
				require.Nil(t, result)

				return
			}

			// Check contents are equal
			require.Equal(t, testCase.input, result)

			// Check that it's a different map
			require.NotSame(t, &testCase.input, &result)

			// Modify result and check original is unchanged
			if len(result) > 0 {
				for k := range result {
					result[k] = 999

					break
				}

				// Find the same key in the original and ensure it's unchanged
				for k, v := range testCase.input {
					if result[k] == 999 {
						require.NotEqual(t, 999, v)
					}
				}
			}
		})
	}
}

func TestForEach(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "process array elements",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{2, 4, 6, 8, 10}, // Double each value
		},
		{
			name:     "empty slice",
			input:    []int{},
			expected: []int(nil), // ForEach with empty slice results in nil slice when using var result []int
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: nil, // Nothing to process
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a result slice to collect the processed items
			var result []int

			// Use ForEach to transform the input
			contextx.ForEach(testCase.input, func(n int) {
				result = append(result, n*2) // Double each number
			})

			// Verify the result
			require.Equal(t, testCase.expected, result)

			// Ensure the original slice wasn't modified
			if testCase.input != nil {
				originalCopy := make([]int, len(testCase.input))
				copy(originalCopy, testCase.input)
				require.Equal(t, originalCopy, testCase.input)
			}
		})
	}
}

func TestFindIndex(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expected  int
	}{
		{
			name:      "find first even number",
			input:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n%2 == 0 },
			expected:  1, // index of 2
		},
		{
			name:      "find number greater than 3",
			input:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n > 3 },
			expected:  3, // index of 4
		},
		{
			name:      "element not found",
			input:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n > 10 },
			expected:  -1, // not found
		},
		{
			name:      "empty slice",
			input:     []int{},
			predicate: func(n int) bool { return n > 0 },
			expected:  -1, // not found in empty slice
		},
		{
			name:      "nil slice",
			input:     nil,
			predicate: func(n int) bool { return n > 0 },
			expected:  -1, // not found in nil slice
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := contextx.FindIndex(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		found     bool
		expected  int
	}{
		{
			name:      "find first even number",
			input:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n%2 == 0 },
			found:     true,
			expected:  2,
		},
		{
			name:      "find number greater than 3",
			input:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n > 3 },
			found:     true,
			expected:  4,
		},
		{
			name:      "element not found",
			input:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n > 10 },
			found:     false,
			expected:  0, // zero value
		},
		{
			name:      "empty slice",
			input:     []int{},
			predicate: func(n int) bool { return n > 0 },
			found:     false,
			expected:  0, // zero value
		},
		{
			name:      "nil slice",
			input:     nil,
			predicate: func(n int) bool { return n > 0 },
			found:     false,
			expected:  0, // zero value
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, found := contextx.Find(testCase.input, testCase.predicate)
			require.Equal(t, testCase.found, found)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestRange(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		expected []int
	}{
		{
			name:     "range 5",
			n:        5,
			expected: []int{0, 1, 2, 3, 4},
		},
		{
			name:     "range 1",
			n:        1,
			expected: []int{0},
		},
		{
			name:     "range 0",
			n:        0,
			expected: []int{},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := contextx.Range(testCase.n)
			require.Equal(t, testCase.expected, result)
		})
	}
}
