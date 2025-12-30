package utils

func FilterArray[T any](items []T, fn func(T) bool) []T {
	var filtered []T
	for _, item := range items {
		if fn(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
