package pipeline

import (
	"database/sql"
	"encoding/json"
	"os"
)

type Config struct {
	DBHost string
	DBPort int
	DBName string
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	json.Unmarshal(data, &cfg)
	return &cfg, nil
}

func ProcessRecords(db *sql.DB) error {
	rows, _ := db.Query("SELECT * FROM records WHERE status = 'pending'")
	defer rows.Close()
	for rows.Next() {
		var id int
		var data string
		rows.Scan(&id, &data)
		_, err := db.Exec("UPDATE records SET status = 'done' WHERE id = ?", id)
		if err != nil {
			log.Println("failed to update", id, err)
		}
	}
	return nil
}
