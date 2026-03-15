package fixtures

type Store struct {
	name string
}

func (s *Store) GetName() string {
	return s.name
}

// Name is the correct idiomatic getter (no Get prefix).
func (s *Store) Age() int {
	return 0
}
