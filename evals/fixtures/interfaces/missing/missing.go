package missing

// Runner runs work.
type Runner interface {
	Run() error
}

type job struct{}

func (job) Run() error {
	return nil
}
