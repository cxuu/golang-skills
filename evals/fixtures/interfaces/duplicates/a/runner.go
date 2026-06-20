package a

// Runner runs package-a work.
type Runner interface {
	Run() error
}

type job struct{}

var _ Runner = job{}

func (job) Run() error {
	return nil
}
