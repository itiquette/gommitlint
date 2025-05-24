// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package slices_test

import (
	"slices"
	"testing"

	commonslices "github.com/itiquette/gommitlint/internal/common/slices"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name   string
		input  []int
		expect []int
		mapFn  func(int) int
	}{
		{
			name:   "Double numbers",
			input:  []int{1, 2, 3},
			expect: []int{2, 4, 6},
			mapFn:  func(i int) int { return i * 2 },
		},
		{
			name:   "Empty slice",
			input:  []int{},
			expect: []int{},
			mapFn:  func(i int) int { return i * 2 },
		},
		{
			name:   "Nil input",
			input:  nil,
			expect: nil,
			mapFn:  func(i int) int { return i * 2 },
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := commonslices.Map(testCase.input, testCase.mapFn)
			require.Equal(t, testCase.expect, result)
		})
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		expect    []int
		predicate func(int) bool
	}{
		{
			name:      "Filter even numbers",
			input:     []int{1, 2, 3, 4, 5},
			expect:    []int{2, 4},
			predicate: func(i int) bool { return i%2 == 0 },
		},
		{
			name:      "Empty slice",
			input:     []int{},
			expect:    []int{},
			predicate: func(i int) bool { return i%2 == 0 },
		},
		{
			name:      "Nil input",
			input:     nil,
			expect:    nil,
			predicate: func(i int) bool { return i%2 == 0 },
		},
		{
			name:      "None match",
			input:     []int{1, 3, 5},
			expect:    []int{},
			predicate: func(i int) bool { return i%2 == 0 },
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := commonslices.Filter(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expect, result)
		})
	}
}

func TestReduce(t *testing.T) {
	tests := []struct {
		name    string
		input   []int
		initial int
		expect  int
		fn      func(int, int) int
	}{
		{
			name:    "Sum numbers",
			input:   []int{1, 2, 3, 4},
			initial: 0,
			expect:  10,
			fn:      func(acc, v int) int { return acc + v },
		},
		{
			name:    "Product numbers",
			input:   []int{2, 3, 4},
			initial: 1,
			expect:  24,
			fn:      func(acc, v int) int { return acc * v },
		},
		{
			name:    "Empty slice",
			input:   []int{},
			initial: 5,
			expect:  5,
			fn:      func(acc, v int) int { return acc + v },
		},
		{
			name:    "Nil input",
			input:   nil,
			initial: 5,
			expect:  5,
			fn:      func(acc, v int) int { return acc + v },
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := commonslices.Reduce(testCase.input, testCase.initial, testCase.fn)
			require.Equal(t, testCase.expect, result)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		target string
		expect bool
	}{
		{
			name:   "Element exists",
			input:  []string{"a", "b", "c"},
			target: "b",
			expect: true,
		},
		{
			name:   "Element not exists",
			input:  []string{"a", "b", "c"},
			target: "d",
			expect: false,
		},
		{
			name:   "Empty slice",
			input:  []string{},
			target: "a",
			expect: false,
		},
		{
			name:   "Nil input",
			input:  nil,
			target: "a",
			expect: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := slices.Contains(testCase.input, testCase.target)
			require.Equal(t, testCase.expect, result)
		})
	}
}

func TestSome(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expect    bool
	}{
		{
			name:      "Some match",
			input:     []int{1, 2, 3, 4},
			predicate: func(i int) bool { return i%2 == 0 },
			expect:    true,
		},
		{
			name:      "None match",
			input:     []int{1, 3, 5},
			predicate: func(i int) bool { return i%2 == 0 },
			expect:    false,
		},
		{
			name:      "Empty slice",
			input:     []int{},
			predicate: func(i int) bool { return i%2 == 0 },
			expect:    false,
		},
		{
			name:      "Nil input",
			input:     nil,
			predicate: func(i int) bool { return i%2 == 0 },
			expect:    false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := slices.ContainsFunc(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expect, result)
		})
	}
}

func TestMapKeys(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]int
		expect []string
	}{
		{
			name:   "Extract keys from map",
			input:  map[string]int{"a": 1, "b": 2, "c": 3},
			expect: []string{"a", "b", "c"},
		},
		{
			name:   "Empty map",
			input:  map[string]int{},
			expect: []string{},
		},
		{
			name:   "Nil input",
			input:  nil,
			expect: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := commonslices.MapKeys(testCase.input)
			if testCase.expect == nil {
				require.Nil(t, result)
			} else {
				// Since maps are unordered, sort both slices before comparing
				sort := func(s []string) []string {
					sort := make([]string, len(s))
					copy(sort, s)
					slices.Sort(sort)

					return sort
				}

				sortedResult := sort(result)
				sortedExpect := sort(testCase.expect)
				require.Equal(t, sortedExpect, sortedResult)
			}
		})
	}
}
