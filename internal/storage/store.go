package storage

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const createReadingsTable = `
CREATE TABLE IF NOT EXISTS readings (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    recorded_at    DATETIME NOT NULL,
    systolic       INTEGER NOT NULL,
    diastolic      INTEGER NOT NULL,
    mean_pressure  INTEGER NOT NULL,
    pulse          INTEGER NOT NULL,
    status         INTEGER NOT NULL,
    user_slot      INTEGER,
    device_time    INTEGER
);`

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("storage: open %s: %w", path, err)
	}

	if _, err := db.Exec(createReadingsTable); err != nil {
		return nil, fmt.Errorf("storage: create readings table: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
