package main

import (
	"testing"
)

func TestNormalizeInput(t *testing.T) {
	err := loadStopWords("/home/edik/Yadro/stopwords.txt")
	if err != nil {
		t.Fatalf("Failed to load stop words: %v", err)
	}

	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "With stopwords",
			input:    "this is a test of the normalization process",
			expected: []string{"test", "normalization", "process"},
		},
		{
			name:     "Without stopwords",
			input:    "short long trees",
			expected: []string{"short", "long", "trees"},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeInput(tc.input)
			if !compareSlices(got, tc.expected) {
				t.Errorf("normalizeInput() = %s, want %s", got, tc.expected)
			}
		})
	}
}

func compareSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
