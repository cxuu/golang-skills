package documented

// Cache stores values by key.
type Cache[K comparable, V any] struct{}

// Map applies f to each value.
func Map[T any](in []T, f func(T) T) []T {
	out := make([]T, 0, len(in))
	for _, v := range in {
		out = append(out, f(v))
	}
	return out
}

// Loader loads a value.
type Loader[T any] interface {
	Load(T) error
}
