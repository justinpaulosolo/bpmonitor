package storage

import "github.com/justinpaulosolo/bpmonitor/internal/a6session"

type StoredReading struct {
	ID           int64
	RecordedAt   int64
	SessionType  string
	ReviewStatus string
	a6session.Reading
}
