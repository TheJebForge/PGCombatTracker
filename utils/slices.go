package utils

// Map why the fuck this is not included in Go's standard library
func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func CreateUpdate[T any](slice []T, foundFunc func(T) bool, createFunc func() T, updateFunc func(T) T) []T {
	for i, t := range slice {
		if foundFunc(t) {
			slice[i] = updateFunc(t)
			return slice
		}
	}

	return append(slice, createFunc())
}
