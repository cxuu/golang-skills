package cache

const MaxRetries = 3

type Store struct {
	name string
}

func (s *Store) Name() string {
	return s.name
}
