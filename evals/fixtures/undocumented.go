package fixtures

func Process() error {
	return nil
}

type Handler struct{}

const (
	MaxItems    = 100
	DefaultPort = 8080
)

// Documented is fine.
func Documented() {}
