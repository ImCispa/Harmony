package utils

func Remove[T comparable](slice []T, num T) []T {
	result := []T{}
	for _, v := range slice {
		if v != num {
			result = append(result, v)
		}
	}
	return result
}
