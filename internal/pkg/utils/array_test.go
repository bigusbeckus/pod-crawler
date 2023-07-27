package utils_test

import (
	"sort"
	"testing"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/utils"
)

func TestLeftDiff(t *testing.T) {
	t.Run("Returns elements unique to left array (int)", func(t *testing.T) {
		a := []int{1, 2, 2, 3, 4, 4, 5, 6, 7, 8}
		b := []int{2, 3, 5, 6, 7}
		expected := []int{1, 4, 4, 8}

		result := utils.LeftDiff(a, b)

		if len(expected) != len(result) {
			t.Fatalf(
				"LeftDiff() returned a slice of incorrect length. Expected %d but received %d\n",
				len(expected),
				len(result),
			)
		}

		sort.Ints(expected)
		sort.Ints(result)
		var mismatch bool
		for i, value := range result {
			if value != expected[i] {
				mismatch = true
				break
			}
		}

		if mismatch {
			t.Fatalf(
				"LeftDiff() returned an incorrect result.\nExpected: %v\nResult: %v\n",
				expected,
				result,
			)
		}
	})
}
