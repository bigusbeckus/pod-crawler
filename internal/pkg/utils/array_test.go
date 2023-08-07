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

func TestJoinNumbers(t *testing.T) {
	t.Run("Returns joined ids on valid input", func(t *testing.T) {
		validInput := []uint64{123, 234, 456}
		separator := ","
		expected := "123,234,456"

		result := utils.JoinNumbers(validInput, separator)
		if result != expected {
			t.Fatalf("JoinNumbers() returned an incorrect result.\nExpected: %s\nResult: %s\n", expected, result)
		}
	})

	t.Run("Returns empty string on empty slice", func(t *testing.T) {
		emptySlice := []int{}
		separator := ","
		expected := ""

		result := utils.JoinNumbers(emptySlice, separator)
		if result != expected {
			t.Fatalf("JoinNumbers() returned an incorrect result.\nExpected: \"\" (empty string)\nResult: %s\n", result)
		}
	})
}
