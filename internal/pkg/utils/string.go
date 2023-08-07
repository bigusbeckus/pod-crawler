package utils

func StringIncludes(arr string, element string) bool {
	for _, value := range arr {
		if string(value) == element {
			return true
		}
	}
	return false
}
