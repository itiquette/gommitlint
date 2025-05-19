// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package slices_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/common/slices"
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
			result := slices.Map(testCase.input, testCase.mapFn)
			require.Equal(t, testCase.expect, result)
		})
	}
}

func TestMapMap(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]int
		expect []string
		mapFn  func(string, int) string
	}{
		{
			name:   "Map to strings",
			input:  map[string]int{"a": 1, "b": 2},
			expect: []string{"a:1", "b:2"},
			mapFn:  func(k string, v int) string { return k + ":" + string(rune(v+'0')) },
		},
		{
			name:   "Empty map",
			input:  map[string]int{},
			expect: []string{},
			mapFn:  func(k string, v int) string { return k + ":" + string(rune(v+'0')) },
		},
		{
			name:   "Nil input",
			input:  nil,
			expect: nil,
			mapFn:  func(k string, v int) string { return k + ":" + string(rune(v+'0')) },
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := slices.MapMap(testCase.input, testCase.mapFn)
			if testCase.expect == nil {
				require.Nil(t, result)
			} else {
				// Since maps are unordered, sort the results to compare
				require.ElementsMatch(t, testCase.expect, result)
			}
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
			result := slices.Filter(testCase.input, testCase.predicate)
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
			result := slices.Reduce(testCase.input, testCase.initial, testCase.fn)
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
			result := slices.Some(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expect, result)
		})
	}
}

func TestEvery(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expect    bool
	}{
		{
			name:      "All match",
			input:     []int{2, 4, 6},
			predicate: func(i int) bool { return i%2 == 0 },
			expect:    true,
		},
		{
			name:      "Some not match",
			input:     []int{2, 3, 4},
			predicate: func(i int) bool { return i%2 == 0 },
			expect:    false,
		},
		{
			name:      "Empty slice",
			input:     []int{},
			predicate: func(i int) bool { return i%2 == 0 },
			expect:    true, // By definition, every element of an empty set satisfies any predicate
		},
		{
			name:      "Nil input",
			input:     nil,
			predicate: func(i int) bool { return i%2 == 0 },
			expect:    true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := slices.Every(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expect, result)
		})
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		predicate   func(string) bool
		expectValue string
		expectFound bool
	}{
		{
			name:        "Element found",
			input:       []string{"apple", "banana", "cherry"},
			predicate:   func(s string) bool { return strings.HasPrefix(s, "b") },
			expectValue: "banana",
			expectFound: true,
		},
		{
			name:        "Element not found",
			input:       []string{"apple", "banana", "cherry"},
			predicate:   func(s string) bool { return strings.HasPrefix(s, "d") },
			expectValue: "",
			expectFound: false,
		},
		{
			name:        "Empty slice",
			input:       []string{},
			predicate:   func(_ string) bool { return true },
			expectValue: "",
			expectFound: false,
		},
		{
			name:        "Nil input",
			input:       nil,
			predicate:   func(_ string) bool { return true },
			expectValue: "",
			expectFound: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			value, found := slices.Find(testCase.input, testCase.predicate)
			require.Equal(t, testCase.expectValue, value)
			require.Equal(t, testCase.expectFound, found)
		})
	}
}

func TestFilterMapKeys(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]int
		excludeKeys []string
		expect      map[string]int
	}{
		{
			name:        "Exclude single key",
			input:       map[string]int{"a": 1, "b": 2, "c": 3},
			excludeKeys: []string{"b"},
			expect:      map[string]int{"a": 1, "c": 3},
		},
		{
			name:        "Exclude multiple keys",
			input:       map[string]int{"a": 1, "b": 2, "c": 3},
			excludeKeys: []string{"b", "c"},
			expect:      map[string]int{"a": 1},
		},
		{
			name:        "Exclude non-existent key",
			input:       map[string]int{"a": 1, "b": 2},
			excludeKeys: []string{"z"},
			expect:      map[string]int{"a": 1, "b": 2},
		},
		{
			name:        "Empty exclude list",
			input:       map[string]int{"a": 1, "b": 2},
			excludeKeys: []string{},
			expect:      map[string]int{"a": 1, "b": 2},
		},
		{
			name:        "Nil input",
			input:       nil,
			excludeKeys: []string{"a"},
			expect:      nil,
		},
		{
			name:        "Empty map",
			input:       map[string]int{},
			excludeKeys: []string{"a"},
			expect:      map[string]int{},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := slices.FilterMapKeys(testCase.input, testCase.excludeKeys)
			require.Equal(t, testCase.expect, result)
		})
	}
}
