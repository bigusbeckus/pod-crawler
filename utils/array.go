package utils

func ArrayIncludes[T comparable](arr []T, element T) bool {
	for _, value := range arr {
		if value == element {
			return true
		}
	}
	return false
}
