package storage

import "github.com/justinpaulosolo/bpmonitor/internal/a6session"

type StoredReading struct {
	ID         int64
	RecordedAt int64
	a6session.Reading
}
