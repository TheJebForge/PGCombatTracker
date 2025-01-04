package utils

func MapCreateUpdate[K comparable, V any](mapObj map[K]V, key K, createFunc func() V, updateFunc func(V) V) {
	if value, has := mapObj[key]; has {
		mapObj[key] = updateFunc(value)
	} else {
		mapObj[key] = createFunc()
	}
}
