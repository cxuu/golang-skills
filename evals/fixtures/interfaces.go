package fixtures

// Reader defines a simple reading interface.
type Reader interface {
	Read(p []byte) (n int, err error)
}

type myReader struct{}

func (r *myReader) Read(p []byte) (int, error) {
	return 0, nil
}
