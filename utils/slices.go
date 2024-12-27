package utils

func CreateUpdate[T any](slice []T, foundFunc func(T) bool, createFunc func() T, updateFunc func(T) T) []T {
	for i, t := range slice {
		if foundFunc(t) {
			slice[i] = updateFunc(t)
			return slice
		}
	}

	return append(slice, createFunc())
}
