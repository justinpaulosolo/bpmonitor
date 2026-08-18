package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/justinpaulosolo/bpmonitor/internal/a6session"
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
    device_time    INTEGER,
    session_type   TEXT NOT NULL,
    review_status  TEXT NOT NULL DEFAULT 'pending'
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
		db.Close()
		return nil, fmt.Errorf("storage: create readings table: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) SaveReading(r a6session.Reading, t time.Time) (int64, error) {
	sessionType := SessionTypeFor(t)
	t = t.UTC()
	if r.DeviceTime != nil {
		query := `SELECT id FROM readings WHERE device_time = ?`
		var id int64
		err := s.db.QueryRow(query, *r.DeviceTime).Scan(&id)
		if err == nil {
			return id, nil
		}
		if err != sql.ErrNoRows {
			return 0, err
		}
	}
	result, err := s.db.Exec(
		`INSERT INTO readings (recorded_at, systolic, diastolic, mean_pressure, pulse, status, user_slot, device_time, session_type)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t,
		r.Systolic,
		r.Diastolic,
		r.MeanPressure,
		r.Pulse,
		r.Status,
		r.UserSlot,
		r.DeviceTime,
		sessionType,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) GetReading(id int64) (*StoredReading, error) {
	row := s.db.QueryRow(`SELECT id, recorded_at, systolic, diastolic, mean_pressure, pulse, status, user_slot, device_time, session_type, review_status FROM readings WHERE id = ?`, id)
	var r StoredReading
	var recordedAt time.Time
	err := row.Scan(
		&r.ID,
		&recordedAt,
		&r.Systolic,
		&r.Diastolic,
		&r.MeanPressure,
		&r.Pulse,
		&r.Status,
		&r.UserSlot,
		&r.DeviceTime,
		&r.SessionType,
		&r.ReviewStatus,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	r.RecordedAt = recordedAt.Unix()
	return &r, nil
}

func (s *Store) GetReadings() ([]StoredReading, error) {
	rows, err := s.db.Query(`SELECT id, recorded_at, systolic, diastolic, mean_pressure, pulse, status, user_slot, device_time, session_type, review_status FROM readings ORDER BY recorded_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []StoredReading
	for rows.Next() {
		var r StoredReading
		var recordedAt time.Time
		err := rows.Scan(
			&r.ID,
			&recordedAt,
			&r.Systolic,
			&r.Diastolic,
			&r.MeanPressure,
			&r.Pulse,
			&r.Status,
			&r.UserSlot,
			&r.DeviceTime,
			&r.SessionType,
			&r.ReviewStatus,
		)
		if err != nil {
			return nil, err
		}
		r.RecordedAt = recordedAt.Unix()
		readings = append(readings, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return readings, nil
}

func (s *Store) GetPendingReadings(sessionType string) ([]StoredReading, error) {
	rows, err := s.db.Query(`SELECT id, recorded_at, systolic, diastolic, mean_pressure, pulse, status, user_slot, device_time, session_type, review_status FROM readings WHERE session_type = ? AND review_status = 'pending' ORDER BY recorded_at`, sessionType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []StoredReading
	for rows.Next() {
		var r StoredReading
		var recordedAt time.Time
		err := rows.Scan(
			&r.ID,
			&recordedAt,
			&r.Systolic,
			&r.Diastolic,
			&r.MeanPressure,
			&r.Pulse,
			&r.Status,
			&r.UserSlot,
			&r.DeviceTime,
			&r.SessionType,
			&r.ReviewStatus,
		)
		if err != nil {
			return nil, err
		}
		r.RecordedAt = recordedAt.Unix()
		readings = append(readings, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return readings, nil
}

func (s *Store) DeleteReading(id int64) error {
	_, err := s.db.Exec(`DELETE FROM readings WHERE id = ?`, id)
	return err
}

func (s *Store) CommitReadings(ids []int64, sessionType string) (err error) {
	if len(ids) != 3 {
		return fmt.Errorf("expected 3 ids to commit, got %d", len(ids))
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Commit the specified readings
	for _, id := range ids {
		_, err = tx.Exec(`UPDATE readings SET review_status = 'committed' WHERE id = ?`, id)
		if err != nil {
			return err
		}
	}

	// Delete the rest of the readings for the session type
	_, err = tx.Exec(`DELETE FROM readings WHERE session_type = ? AND review_status = 'pending'`, sessionType)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return
}

func SessionTypeFor(t time.Time) string {
	t = t.Local()
	if t.Hour() < 6 {
		return "night"
	}
	return "morning"
}
