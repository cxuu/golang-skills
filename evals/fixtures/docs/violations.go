package docsbad

type Cache[K comparable, V any] struct{}

func Map[T any](in []T) []T {
	return in
}

type Loader[T any] interface {
	Load(T) error
}
