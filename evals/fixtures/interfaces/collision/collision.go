package collision

// MyStringer is consumed by this package; no local type is intended to
// implement it.
type MyStringer interface {
	String(prefix string) string
}

type Unrelated struct{}

func (Unrelated) String() string {
	return "x"
}
