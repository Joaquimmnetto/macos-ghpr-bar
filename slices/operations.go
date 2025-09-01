package slices

func Filter[T any](items []T, predicate func(T) bool) []T {
	var filtered []T
	for _, item := range items {
		if predicate(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func Split[T any](items []T, predicate func(T) bool) ([]T, []T) {
	var trueGroup []T
	var falseGroup []T
	for _, item := range items {
		if predicate(item) {
			trueGroup = append(trueGroup, item)
		} else {
			falseGroup = append(falseGroup, item)
		}
	}
	return trueGroup, falseGroup
}

func Any[T any](items []T, predicate func(T) bool) bool {
	for _, item := range items {
		if predicate(item) {
			return true
		}
	}
	return false
}
