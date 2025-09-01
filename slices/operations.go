package slices

import (
	"sync"
)

func MapParallelMany[K comparable, V any, R any](items map[K]V, operation func(K, V) R) map[K]R {
	var categoryWg sync.WaitGroup
	var result = make(map[K]R, len(items))
	for key, value := range items {
		categoryWg.Add(1)
		go func() {
			defer categoryWg.Done()
			result[key] = operation(key, value)
		}()
	}
	categoryWg.Wait()
	return result
}

func ParallelMany[T any, R any](items []T, operation func(T) []R) []R {
	result := make([]R, len(items))
	var queriesWg sync.WaitGroup
	for _, item := range items {
		queriesWg.Add(1)
		go func() {
			defer queriesWg.Done()
			r := operation(item)
			result = append(result, r...)
		}()
	}
	queriesWg.Wait()
	return result
}

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
