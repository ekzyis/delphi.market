package lib

func Merge[T comparable](target *map[T]any, src *map[T]any) {
	for k, v := range *src {
		(*target)[k] = v
	}
}
