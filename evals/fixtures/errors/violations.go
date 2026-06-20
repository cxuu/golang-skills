package errorsbad

import (
	"errors"
	"log"
)

func LoadConfig(path string) error {
	err := readFile(path)
	if err != nil {
		return err
	}
	return nil
}

func ValidateInput(s string) error {
	err := check(s)
	if err != nil {
		if err.Error() == "not found" {
			return errors.New("missing")
		}
		return err
	}
	return nil
}

func ProcessRecord(id int) error {
	err := fetch(id)
	if err != nil {
		log.Printf("failed to process record %d: %v", id, err)
		return err
	}
	return nil
}

func FetchName() (string, error) {
	err := fetch(1)
	if err != nil {
		return "", err
	}
	return "ok", nil
}

func readFile(_ string) error { return nil }
func check(_ string) error    { return errors.New("not found") }
func fetch(_ int) error       { return nil }
