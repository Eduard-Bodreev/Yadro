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
		expected string
	}{
		{
			name:     "With stopwords",
			input:    "this is a test of the normalization process",
			expected: "test normalization process",
		},
		{
			name:     "Without stopwords",
			input:    "short long trees",
			expected: "short long trees",
		},
		{
			name:     "Empty input",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeInput(tc.input); got != tc.expected {
				t.Errorf("normalizeInput() = %q, want %q", got, tc.expected)
			}
		})
	}
}
