package utils

func Contains[T comparable](slice []T, num T) bool {
	for _, v := range slice {
		if v == num {
			return true
		}
	}
	return false
}
