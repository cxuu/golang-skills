package errorsclean

import "fmt"

func LoadConfig(path string) error {
	if err := readFile(path); err != nil {
		return fmt.Errorf("load config %q: %w", path, err)
	}
	return nil
}

func readFile(_ string) error { return nil }
