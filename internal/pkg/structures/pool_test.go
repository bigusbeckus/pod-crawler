package structures_test

import (
	"reflect"
	"testing"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/structures"
)

func TestPool(t *testing.T) {
	pool := structures.CreatePool([]uint64{10, 200, 125})

	t.Run("Get pool length", func(t *testing.T) {
		expectedLength := 3
		actualLength := pool.Length()
		if actualLength != expectedLength {
			t.Errorf("Expected pool length %d, but got %d", expectedLength, actualLength)
		}
	})

	t.Run("Add to pool", func(t *testing.T) {
		pool.Put(uint64(50))
		expectedLength := 4
		actualLength := pool.Length()
		if actualLength != expectedLength {
			t.Errorf("Expected pool length %d after adding an item, but got %d", expectedLength, actualLength)
		}
	})

	t.Run("Remove from pool", func(t *testing.T) {
		items := pool.Take(2)
		expectedLength := 2
		actualLength := pool.Length()
		if actualLength != expectedLength {
			t.Errorf("Expected pool length %d after taking 2 items, but got %d", expectedLength, actualLength)
		}

		// Now, check if the taken items are as expected
		expectedItems := []uint64{10, 200}
		if !reflect.DeepEqual(items, expectedItems) {
			t.Errorf("Expected items %v to be taken from the pool, but got %v", expectedItems, items)
		}
	})

	t.Run("Shuffle pool", func(t *testing.T) {
		// Create a new pool with some ordered elements
		orderedPool := structures.CreatePool([]int{1, 2, 3, 4, 5})

		// Shuffle the pool
		orderedPool.Shuffle()

		// Ensure the length remains the same
		expectedLength := 5
		actualLength := orderedPool.Length()
		if actualLength != expectedLength {
			t.Errorf("Expected pool length %d after shuffling, but got %d", expectedLength, actualLength)
		}

		// Ensure that the shuffled pool contains all the same elements, but in a different order
		originalElements := []int{1, 2, 3, 4, 5}
		shuffledElements := orderedPool.Take(5)
		if len(shuffledElements) != len(originalElements) {
			t.Errorf("Expected %d elements in the shuffled pool, but got %d", len(originalElements), len(shuffledElements))
		}
		// Check that all original elements are still present
		for _, elem := range originalElements {
			if !contains(shuffledElements, elem) {
				t.Errorf("Expected shuffled pool to contain element %d, but it's missing", elem)
			}
		}
	})
}

// Helper function to check if a slice contains a specific element
func contains(slice []int, elem int) bool {
	for _, item := range slice {
		if item == elem {
			return true
		}
	}
	return false
}
