package consumer

// Runner is consumed by this package; implementations live elsewhere.
type Runner interface {
	Run() error
}

func Use(r Runner) error {
	return r.Run()
}
