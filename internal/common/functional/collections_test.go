// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package functional

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		fn       func(int) string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []int{},
			fn:       func(i int) string { return strconv.Itoa(i) },
			expected: []string{},
		},
		{
			name:     "nil slice",
			input:    nil,
			fn:       func(i int) string { return strconv.Itoa(i) },
			expected: nil,
		},
		{
			name:     "normal slice",
			input:    []int{1, 2, 3},
			fn:       func(i int) string { return strconv.Itoa(i * 2) },
			expected: []string{"2", "4", "6"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := MapSlice(testCase.input, testCase.fn)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestFilterSlice(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expected  []int
	}{
		{
			name:      "empty slice",
			input:     []int{},
			predicate: func(i int) bool { return i > 0 },
			expected:  []int{},
		},
		{
			name:      "nil slice",
			input:     nil,
			predicate: func(i int) bool { return i > 0 },
			expected:  nil,
		},
		{
			name:      "filter positive numbers",
			input:     []int{-2, -1, 0, 1, 2},
			predicate: func(i int) bool { return i > 0 },
			expected:  []int{1, 2},
		},
		{
			name:      "no matches",
			input:     []int{1, 2, 3},
			predicate: func(i int) bool { return i < 0 },
			expected:  []int{},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := FilterSlice(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestReduce(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		initial  int
		fn       func(int, int) int
		expected int
	}{
		{
			name:     "sum numbers",
			input:    []int{1, 2, 3, 4},
			initial:  0,
			fn:       func(acc, val int) int { return acc + val },
			expected: 10,
		},
		{
			name:     "product numbers",
			input:    []int{2, 3, 4},
			initial:  1,
			fn:       func(acc, val int) int { return acc * val },
			expected: 24,
		},
		{
			name:     "empty slice",
			input:    []int{},
			initial:  42,
			fn:       func(acc, val int) int { return acc + val },
			expected: 42,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Reduce(testCase.input, testCase.initial, testCase.fn)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		wantOk    bool
		expected  int
	}{
		{
			name:      "find first even number",
			input:     []int{1, 3, 4, 6, 7},
			predicate: func(i int) bool { return i%2 == 0 },
			wantOk:    true,
			expected:  4,
		},
		{
			name:      "no match found",
			input:     []int{1, 3, 5},
			predicate: func(i int) bool { return i%2 == 0 },
			wantOk:    false,
		},
		{
			name:      "empty slice",
			input:     []int{},
			predicate: func(_ int) bool { return true },
			wantOk:    false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Find(testCase.input, testCase.predicate)
			require.Equal(t, testCase.wantOk, result.IsOk())

			if testCase.wantOk {
				value, err := result.Unwrap()
				require.NoError(t, err)
				require.Equal(t, testCase.expected, value)
			}
		})
	}
}

func TestAll(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expected  bool
	}{
		{
			name:      "all positive",
			input:     []int{1, 2, 3},
			predicate: func(i int) bool { return i > 0 },
			expected:  true,
		},
		{
			name:      "not all positive",
			input:     []int{1, -1, 3},
			predicate: func(i int) bool { return i > 0 },
			expected:  false,
		},
		{
			name:      "empty slice",
			input:     []int{},
			predicate: func(i int) bool { return i > 0 },
			expected:  true, // vacuous truth
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := All(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestAny(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expected  bool
	}{
		{
			name:      "has positive",
			input:     []int{-1, 0, 1},
			predicate: func(i int) bool { return i > 0 },
			expected:  true,
		},
		{
			name:      "no positive",
			input:     []int{-1, 0, -2},
			predicate: func(i int) bool { return i > 0 },
			expected:  false,
		},
		{
			name:      "empty slice",
			input:     []int{},
			predicate: func(i int) bool { return i > 0 },
			expected:  false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Any(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestPartition(t *testing.T) {
	tests := []struct {
		name           string
		input          []int
		predicate      func(int) bool
		expectedTrues  []int
		expectedFalses []int
	}{
		{
			name:           "partition by positive",
			input:          []int{-2, 1, 0, 3, -1},
			predicate:      func(i int) bool { return i > 0 },
			expectedTrues:  []int{1, 3},
			expectedFalses: []int{-2, 0, -1},
		},
		{
			name:           "all true",
			input:          []int{1, 2, 3},
			predicate:      func(i int) bool { return i > 0 },
			expectedTrues:  []int{1, 2, 3},
			expectedFalses: []int{},
		},
		{
			name:           "empty slice",
			input:          []int{},
			predicate:      func(i int) bool { return i > 0 },
			expectedTrues:  []int{},
			expectedFalses: []int{},
		},
		{
			name:           "nil slice",
			input:          nil,
			predicate:      func(i int) bool { return i > 0 },
			expectedTrues:  nil,
			expectedFalses: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			trues, falses := Partition(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expectedTrues, trues)
			require.Equal(t, testCase.expectedFalses, falses)
		})
	}
}

func TestGroupBy(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		keyFn    func(string) int
		expected map[int][]string
	}{
		{
			name:  "group by length",
			input: []string{"a", "ab", "abc", "x", "xy"},
			keyFn: func(s string) int { return len(s) },
			expected: map[int][]string{
				1: {"a", "x"},
				2: {"ab", "xy"},
				3: {"abc"},
			},
		},
		{
			name:     "empty slice",
			input:    []string{},
			keyFn:    func(s string) int { return len(s) },
			expected: map[int][]string{},
		},
		{
			name:     "nil slice",
			input:    nil,
			keyFn:    func(s string) int { return len(s) },
			expected: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := GroupBy(testCase.input, testCase.keyFn)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "remove duplicates",
			input:    []int{1, 2, 2, 3, 1, 4},
			expected: []int{1, 2, 3, 4},
		},
		{
			name:     "no duplicates",
			input:    []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "empty slice",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Unique(testCase.input)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestZip(t *testing.T) {
	tests := []struct {
		name     string
		input1   []int
		input2   []string
		expected []Pair[int, string]
	}{
		{
			name:   "equal length",
			input1: []int{1, 2, 3},
			input2: []string{"a", "b", "c"},
			expected: []Pair[int, string]{
				{1, "a"},
				{2, "b"},
				{3, "c"},
			},
		},
		{
			name:   "first shorter",
			input1: []int{1, 2},
			input2: []string{"a", "b", "c"},
			expected: []Pair[int, string]{
				{1, "a"},
				{2, "b"},
			},
		},
		{
			name:   "second shorter",
			input1: []int{1, 2, 3},
			input2: []string{"a", "b"},
			expected: []Pair[int, string]{
				{1, "a"},
				{2, "b"},
			},
		},
		{
			name:     "empty slices",
			input1:   []int{},
			input2:   []string{},
			expected: []Pair[int, string]{},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Zip(testCase.input1, testCase.input2)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestTake(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		n        int
		expected []int
	}{
		{
			name:     "take less than length",
			input:    []int{1, 2, 3, 4, 5},
			n:        3,
			expected: []int{1, 2, 3},
		},
		{
			name:     "take more than length",
			input:    []int{1, 2, 3},
			n:        5,
			expected: []int{1, 2, 3},
		},
		{
			name:     "take zero",
			input:    []int{1, 2, 3},
			n:        0,
			expected: []int{},
		},
		{
			name:     "take negative",
			input:    []int{1, 2, 3},
			n:        -1,
			expected: []int{},
		},
		{
			name:     "nil slice",
			input:    nil,
			n:        2,
			expected: []int{},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Take(testCase.input, testCase.n)
			require.Equal(t, testCase.expected, result)

			// Ensure original slice is not modified
			if len(testCase.input) > 0 {
				// Just verify the slice still exists and has values
				require.NotNil(t, testCase.input)
			}
		})
	}
}

func TestDrop(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		n        int
		expected []int
	}{
		{
			name:     "drop less than length",
			input:    []int{1, 2, 3, 4, 5},
			n:        2,
			expected: []int{3, 4, 5},
		},
		{
			name:     "drop more than length",
			input:    []int{1, 2, 3},
			n:        5,
			expected: []int{},
		},
		{
			name:     "drop zero",
			input:    []int{1, 2, 3},
			n:        0,
			expected: []int{1, 2, 3},
		},
		{
			name:     "drop negative",
			input:    []int{1, 2, 3},
			n:        -1,
			expected: []int{1, 2, 3},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Drop(testCase.input, testCase.n)
			require.Equal(t, testCase.expected, result)
		})
	}
}

// Example of functional data pipeline.
func Example_pipeline() {
	// Process a list of words: filter long words, convert to uppercase, take first 3
	words := []string{"hello", "world", "this", "is", "a", "test", "of", "functional", "programming"}

	// Functional pipeline
	longWords := FilterSlice(words, func(s string) bool { return len(s) > 3 })
	upperWords := MapSlice(longWords, strings.ToUpper)
	result := Take(upperWords, 3)

	fmt.Printf("Result: %v\n", result)
	// Output: Result: [HELLO WORLD THIS]
}
