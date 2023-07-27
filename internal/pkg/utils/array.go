package utils

func ArrayIncludes[T comparable](arr []T, element T) bool {
	for _, value := range arr {
		if value == element {
			return true
		}
	}
	return false
}

func Deduplicate[T comparable](arr []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}

	for _, item := range arr {
		if _, ok := allKeys[item]; !ok {
			allKeys[item] = true
			list = append(list, item)
		}
	}

	return list
}

// Given two slices, a and b, returns the values in a that are not present in b
func LeftDiff[T comparable](a []T, b []T) []T {
	// Put second slice (b) into a map for fast "exists" checks
	bMap := make(map[T]int)
	for _, item := range b {
		bMap[item]++
	}

	// Pre-allocate memory for unique elements slice
	uniqueToA := make([]T, 0, len(a))

	// Append elements of "a" not in "bMap" to the unique elements slice
	for _, item := range a {
		_, exists := bMap[item]
		if !exists {
			uniqueToA = append(uniqueToA, item)
		}
	}

	return uniqueToA
}
