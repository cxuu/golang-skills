package good

// Reader reads bytes.
type Reader interface {
	Read([]byte) (int, error)
}

type reader struct{}

var _ Reader = (*reader)(nil)

func (*reader) Read([]byte) (int, error) {
	return 0, nil
}
