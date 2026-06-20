package multiline

func LoadTriple() (int, int, error) {
	err := fetch()
	if err != nil {
		return 0,
			0,
			err
	}
	return 1, 2, nil
}

func fetch() error { return nil }
