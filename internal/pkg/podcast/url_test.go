package podcast_test

import (
	"testing"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/podcast"
)

func TestExtractLookupIDs(t *testing.T) {
	tests := []struct {
		title string
		input string
		want  []uint64
	}{
		// Test case 1: Basic input with valid IDs
		{
			title: "Basic input with valid IDs",
			input: "https://example.com/api?&id=123,456,789",
			want:  []uint64{123, 456, 789},
		},
		// Test case 2: Empty input string, should return an empty slice
		{
			title: "Empty input string, should return an empty slice",
			input: "",
			want:  []uint64{},
		},
		// Test case 3: No 'id' parameter, should return an empty slice
		{
			title: "No 'id' parameter, should return an empty slice",
			input: "https://example.com/api?key=value",
			want:  []uint64{},
		},
		// Test case 4: Invalid ID should be ignored and logged, returning 0 for the invalid ID
		{
			title: "Invalid ID should be ignored",
			input: "https://example.com/api?&id=123,abc,789",
			want:  []uint64{123, 789},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			result := podcast.ExtractLookupIDs(test.input)

			t.Logf("result %d, want %d", len(result), len(test.want))

			if len(result) != len(test.want) {
				t.Errorf("Input: %s, got: %v, want: %v", test.input, result, test.want)
			}

			for i := range result {
				if result[i] != test.want[i] {
					t.Errorf("Input: %s, got: %v, want: %v", test.input, result, test.want)
				}
			}
		})
	}
}
