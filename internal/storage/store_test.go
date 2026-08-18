package storage

import (
	"testing"
	"time"

	"github.com/justinpaulosolo/bpmonitor/internal/a6session"
)

func TestOpen(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}

	defer store.Close()

	var name string
	err = store.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='readings'`).Scan(&name)
	if err != nil {
		t.Fatalf("readings table not found: %v", err)
	}
	if name != "readings" {
		t.Errorf("got table name %q, want %q", name, "readings")
	}
}

func TestSaveReading(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}
	defer store.Close()

	userSlot := 2
	deviceTime := uint32(500000000)

	r1 := a6session.Reading{
		Systolic:     120,
		Diastolic:    80,
		MeanPressure: 93,
		Pulse:        70,
		Status:       0,
		UserSlot:     &userSlot,
		DeviceTime:   &deviceTime,
	}

	r2 := a6session.Reading{
		Systolic:     130,
		Diastolic:    90,
		MeanPressure: 103,
		Pulse:        75,
		Status:       0,
		UserSlot:     nil,
		DeviceTime:   nil,
	}
	id1, err := store.SaveReading(r1, time.Now())
	if err != nil {
		t.Fatalf("SaveReading(r1) returned error: %v", err)
	}
	if id1 != 1 {
		t.Errorf("SaveReading(r1) returned id=%d, want 1", id1)
	}

	id2, err := store.SaveReading(r2, time.Now())
	if err != nil {
		t.Fatalf("SaveReading(r2) returned error: %v", err)
	}
	if id2 != 2 {
		t.Errorf("SaveReading(r2) returned id=%d, want 2", id2)
	}

	var count int
	err = store.db.QueryRow(`SELECT COUNT(*) FROM readings`).Scan(&count)
	if err != nil {
		t.Fatalf("querying row count: %v", err)
	}
	if count != 2 {
		t.Errorf("got %d rows, want 2", count)
	}
}

func TestSaveReading_DuplicateDeviceTime(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}
	defer store.Close()

	userSlot := 2
	deviceTime := uint32(500000000)

	r1 := a6session.Reading{
		Systolic:     120,
		Diastolic:    80,
		MeanPressure: 93,
		Pulse:        70,
		Status:       0,
		UserSlot:     &userSlot,
		DeviceTime:   &deviceTime,
	}

	r2 := a6session.Reading{
		Systolic:     130,
		Diastolic:    90,
		MeanPressure: 103,
		Pulse:        75,
		Status:       0,
		UserSlot:     &userSlot,
		DeviceTime:   &deviceTime,
	}

	id1, err := store.SaveReading(r1, time.Now())
	if err != nil {
		t.Fatalf("SaveReading(r1) returned error: %v", err)
	}

	id2, err := store.SaveReading(r2, time.Now())
	if err != nil {
		t.Fatalf("SaveReading(r2) returned error: %v", err)
	}

	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM readings`).Scan(&count); err != nil {
		t.Fatalf("querying row count: %v", err)
	}

	if count != 1 {
		t.Errorf("got %d rows, want 1", count)
	}

	if id1 != id2 {
		t.Errorf("expected duplicate device time to result in same ID, got id1=%d, id2=%d", id1, id2)
	}
}

func TestGetReading(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}
	defer store.Close()

	userSlot := 2
	deviceTime := uint32(500000000)

	r1 := a6session.Reading{
		Systolic:     120,
		Diastolic:    80,
		MeanPressure: 93,
		Pulse:        70,
		Status:       0,
		UserSlot:     &userSlot,
		DeviceTime:   &deviceTime,
	}

	id, err := store.SaveReading(r1, time.Now())

	if err != nil {
		t.Fatalf("SaveReading(r1) returned error: %v", err)
	}
	if id != 1 {
		t.Errorf("SaveReading(r1) returned id=%d, want 1", id)
	}

	r, err := store.GetReading(id)
	if err != nil {
		t.Fatalf("GetReading(%d) returned error: %v", id, err)
	}
	if r == nil {
		t.Fatalf("GetReading(%d) returned nil", id)
	}
	if r.ID != id {
		t.Errorf("GetReading(%d) returned reading with ID=%d, want %d", id, r.ID, id)
	}
	if r.Systolic != r1.Systolic || r.Diastolic != r1.Diastolic || r.MeanPressure != r1.MeanPressure || r.Pulse != r1.Pulse || r.Status != r1.Status {
		t.Errorf("GetReading(%d) returned reading %+v, want %+v", id, r, r1)
	}
}

func TestGetReadings(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}
	defer store.Close()

	userSlot := 2
	deviceTime := uint32(500000000)

	r1 := a6session.Reading{
		Systolic:     120,
		Diastolic:    80,
		MeanPressure: 93,
		Pulse:        70,
		Status:       0,
		UserSlot:     &userSlot,
		DeviceTime:   &deviceTime,
	}

	r2 := a6session.Reading{
		Systolic:     130,
		Diastolic:    90,
		MeanPressure: 103,
		Pulse:        75,
		Status:       0,
		UserSlot:     nil,
		DeviceTime:   nil,
	}

	_, err = store.SaveReading(r1, time.Now())
	if err != nil {
		t.Fatalf("SaveReading(r1) returned error: %v", err)
	}
	_, err = store.SaveReading(r2, time.Now())
	if err != nil {
		t.Fatalf("SaveReading(r2) returned error: %v", err)
	}

	readings, err := store.GetReadings()
	if err != nil {
		t.Fatalf("GetReadings returned error: %v", err)
	}
	if len(readings) != 2 {
		t.Errorf("GetReadings returned %d readings, want 2", len(readings))
	}
	if readings[0].Systolic != r1.Systolic || readings[1].Systolic != r2.Systolic {
		t.Errorf("GetReadings returned readings %+v, want %+v and %+v", readings, r1, r2)
	}
}
