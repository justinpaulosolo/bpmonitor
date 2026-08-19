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
	if r.ReviewStatus != "pending" {
		t.Errorf("GetReading(%d) returned ReviewStatus=%q, want %q", id, r.ReviewStatus, "pending")
	}
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

func TestGetPendingReadings(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}
	defer store.Close()

	userSlot := 2
	deviceTime1 := uint32(1)
	deviceTime2 := uint32(2)
	deviceTime3 := uint32(3)

	morning1 := a6session.Reading{
		Systolic: 120, Diastolic: 80, MeanPressure: 93, Pulse: 70, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime1,
	}
	morning2 := a6session.Reading{
		Systolic: 122, Diastolic: 81, MeanPressure: 94, Pulse: 72, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime2,
	}
	night1 := a6session.Reading{
		Systolic: 130, Diastolic: 90, MeanPressure: 103, Pulse: 75, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime3,
	}

	morningTime := time.Date(2026, 8, 18, 9, 0, 0, 0, time.Local)
	nightTime := time.Date(2026, 8, 18, 1, 0, 0, 0, time.Local)

	if _, err := store.SaveReading(morning1, morningTime); err != nil {
		t.Fatalf("SaveReading(morning1) returned error: %v", err)
	}
	if _, err := store.SaveReading(morning2, morningTime.Add(time.Minute)); err != nil {
		t.Fatalf("SaveReading(morning2) returned error: %v", err)
	}
	if _, err := store.SaveReading(night1, nightTime); err != nil {
		t.Fatalf("SaveReading(night1) returned error: %v", err)
	}

	readings, err := store.GetPendingReadings("morning", "2026-08-18")
	if err != nil {
		t.Fatalf("GetPendingReadings returned error: %v", err)
	}
	if len(readings) != 2 {
		t.Fatalf("got %d readings, want 2", len(readings))
	}
	if readings[0].Systolic != morning1.Systolic || readings[1].Systolic != morning2.Systolic {
		t.Errorf("got readings %+v, want systolic 120 then 122 in order", readings)
	}
	for _, r := range readings {
		if r.SessionType != "morning" {
			t.Errorf("got SessionType=%q, want %q", r.SessionType, "morning")
		}
		if r.ReviewStatus != "pending" {
			t.Errorf("got ReviewStatus=%q, want %q", r.ReviewStatus, "pending")
		}
	}
}

func TestDelete(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}
	defer store.Close()

	userSlot := 2
	deviceTime1 := uint32(1)

	morning1 := a6session.Reading{
		Systolic: 120, Diastolic: 80, MeanPressure: 93, Pulse: 70, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime1,
	}

	morningTime := time.Date(2026, 8, 18, 9, 0, 0, 0, time.Local)

	id, err := store.SaveReading(morning1, morningTime)
	if err != nil {
		t.Fatalf("SaveReading(morning1) returned error: %v", err)
	}

	if err := store.DeleteReading(id); err != nil {
		t.Fatalf("DeleteReading returned error: %v", err)
	}
	r, err := store.GetReading(id)
	if err != nil {
		t.Fatalf("GetReading returned error: %v", err)
	}
	if r != nil {
		t.Fatalf("expected reading to be deleted, but got %+v", r)
	}
}

func TestRejectReading(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}
	defer store.Close()

	userSlot := 2
	deviceTime1 := uint32(1)

	morning1 := a6session.Reading{
		Systolic: 120, Diastolic: 80, MeanPressure: 93, Pulse: 70, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime1,
	}

	morningTime := time.Date(2026, 8, 18, 9, 0, 0, 0, time.Local)

	id, err := store.SaveReading(morning1, morningTime)
	if err != nil {
		t.Fatalf("SaveReading(morning1) returned error: %v", err)
	}

	if err := store.RejectReading(id); err != nil {
		t.Fatalf("RejectReading returned error: %v", err)
	}

	r, err := store.GetReading(id)
	if err != nil {
		t.Fatalf("GetReading returned error: %v", err)
	}
	if r == nil {
		t.Fatal("GetReading returned nil, want the rejected reading to still exist")
	}
	if r.ReviewStatus != "rejected" {
		t.Errorf("ReviewStatus = %q, want %q", r.ReviewStatus, "rejected")
	}

	pending, err := store.GetPendingReadings("morning", "2026-08-18")
	if err != nil {
		t.Fatalf("GetPendingReadings returned error: %v", err)
	}
	for _, p := range pending {
		if p.ID == id {
			t.Errorf("rejected reading %d still appears in GetPendingReadings: %+v", id, pending)
		}
	}
}

func TestRestoreReading(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}
	defer store.Close()

	userSlot := 2
	deviceTime1 := uint32(1)

	morning1 := a6session.Reading{
		Systolic: 120, Diastolic: 80, MeanPressure: 93, Pulse: 70, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime1,
	}

	morningTime := time.Date(2026, 8, 18, 9, 0, 0, 0, time.Local)

	id, err := store.SaveReading(morning1, morningTime)
	if err != nil {
		t.Fatalf("SaveReading(morning1) returned error: %v", err)
	}
	if err := store.RejectReading(id); err != nil {
		t.Fatalf("RejectReading returned error: %v", err)
	}

	if err := store.RestoreReading(id); err != nil {
		t.Fatalf("RestoreReading returned error: %v", err)
	}

	r, err := store.GetReading(id)
	if err != nil {
		t.Fatalf("GetReading returned error: %v", err)
	}
	if r == nil {
		t.Fatal("GetReading returned nil, want the restored reading to still exist")
	}
	if r.ReviewStatus != "pending" {
		t.Errorf("ReviewStatus = %q, want %q", r.ReviewStatus, "pending")
	}

	pending, err := store.GetPendingReadings("morning", "2026-08-18")
	if err != nil {
		t.Fatalf("GetPendingReadings returned error: %v", err)
	}
	found := false
	for _, p := range pending {
		if p.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("restored reading %d does not appear in GetPendingReadings: %+v", id, pending)
	}
}

func TestCommitReadings(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}
	defer store.Close()

	userSlot := 2
	deviceTime1 := uint32(1)
	deviceTime2 := uint32(2)
	deviceTime3 := uint32(3)
	deviceTime4 := uint32(4)

	morning1 := a6session.Reading{
		Systolic: 120, Diastolic: 80, MeanPressure: 93, Pulse: 70, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime1,
	}
	morning2 := a6session.Reading{
		Systolic: 122, Diastolic: 81, MeanPressure: 94, Pulse: 72, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime2,
	}
	morning3 := a6session.Reading{
		Systolic: 130, Diastolic: 90, MeanPressure: 103, Pulse: 75, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime3,
	}

	morning4 := a6session.Reading{
		Systolic: 135, Diastolic: 95, MeanPressure: 108, Pulse: 78, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime4,
	}

	morningTime := time.Date(2026, 8, 18, 9, 0, 0, 0, time.Local)

	r1, err := store.SaveReading(morning1, morningTime)
	if err != nil {
		t.Fatalf("SaveReading(morning1) returned error: %v", err)
	}
	r2, err := store.SaveReading(morning2, morningTime.Add(time.Minute))
	if err != nil {
		t.Fatalf("SaveReading(morning2) returned error: %v", err)
	}
	r3, err := store.SaveReading(morning3, morningTime.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("SaveReading(morning3) returned error: %v", err)
	}
	r4, err := store.SaveReading(morning4, morningTime.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("SaveReading(morning4) returned error: %v", err)
	}

	store.CommitReadings([]int64{r1, r2, r3}, "morning", "2026-08-18")
	r, err := store.GetReading(r1)
	if err != nil {
		t.Fatalf("GetReading(r1) returned error: %v", err)
	}
	if r.ReviewStatus != "committed" {
		t.Errorf("GetReading(r1).ReviewStatus = %q, want %q", r.ReviewStatus, "committed")
	}

	r, err = store.GetReading(r2)
	if err != nil {
		t.Fatalf("GetReading(r2) returned error: %v", err)
	}
	if r.ReviewStatus != "committed" {
		t.Errorf("GetReading(r2).ReviewStatus = %q, want %q", r.ReviewStatus, "committed")
	}

	r, err = store.GetReading(r3)
	if err != nil {
		t.Fatalf("GetReading(r3) returned error: %v", err)
	}
	if r.ReviewStatus != "committed" {
		t.Errorf("GetReading(r3).ReviewStatus = %q, want %q", r.ReviewStatus, "committed")
	}

	r, err = store.GetReading(r4)
	if err != nil {
		t.Fatalf("GetReading(r4) returned error: %v", err)
	}
	if r != nil {
		t.Errorf("GetReading(r4) = %v, want nil", r)
	}

	pending, err := store.GetPendingReadings("morning", "2026-08-18")
	if err != nil {
		t.Fatalf("GetPendingReadings(\"morning\", \"2026-08-18\") returned error: %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("GetPendingReadings(\"morning\", \"2026-08-18\") = %v, want empty slice", pending)
	}
}

func TestCommitReadings_WrongCount(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}
	defer store.Close()

	userSlot := 2
	deviceTime1 := uint32(1)
	deviceTime2 := uint32(2)

	morning1 := a6session.Reading{
		Systolic: 120, Diastolic: 80, MeanPressure: 93, Pulse: 70, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime1,
	}
	morning2 := a6session.Reading{
		Systolic: 122, Diastolic: 81, MeanPressure: 94, Pulse: 72, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime2,
	}
	morningTime := time.Date(2026, 8, 18, 9, 0, 0, 0, time.Local)

	r1, err := store.SaveReading(morning1, morningTime)
	if err != nil {
		t.Fatalf("SaveReading(morning1) returned error: %v", err)
	}
	r2, err := store.SaveReading(morning2, morningTime.Add(time.Minute))
	if err != nil {
		t.Fatalf("SaveReading(morning2) returned error: %v", err)
	}

	err = store.CommitReadings([]int64{r1, r2}, "morning", "2026-08-18")
	if err == nil {
		t.Fatal("CommitReadings with 2 ids returned no error, want an error")
	}

	readings, err := store.GetPendingReadings("morning", "2026-08-18")
	if err != nil {
		t.Fatalf("GetPendingReadings(\"morning\", \"2026-08-18\") returned error: %v", err)
	}
	if len(readings) != 2 {
		t.Errorf("got %d pending readings after rejected commit, want 2 (nothing should have changed)", len(readings))
	}
}

func TestSessionTypeFor(t *testing.T) {
	// 1am
	t1 := time.Date(2026, 8, 18, 1, 0, 0, 0, time.Local)
	// 10 am
	t2 := time.Date(2026, 8, 18, 10, 0, 0, 0, time.Local)

	sessionType := SessionTypeFor(t1)
	if sessionType != "night" {
		t.Errorf("SessionTypeFor(%v) = %q, want %q", t1, sessionType, "night")
	}
	sessionType = SessionTypeFor(t2)
	if sessionType != "morning" {
		t.Errorf("SessionTypeFor(%v) = %q, want %q", t2, sessionType, "morning")
	}
}

func TestSessionDateFor(t *testing.T) {
	// August 18, 2026, 1am
	t1 := time.Date(2026, 8, 18, 1, 0, 0, 0, time.Local)
	expected := "2026-08-18"
	sessionDate := SessionDateFor(t1)
	if sessionDate != expected {
		t.Errorf("SessionDateFor(%v) = %q, want %q", t1, sessionDate, expected)
	}

	// Same day, 9am - night + morning readings should share the same date
	t2 := time.Date(2026, 8, 18, 9, 0, 0, 0, time.Local)
	sessionDate = SessionDateFor(t2)
	if sessionDate != expected {
		t.Errorf("SessionDateFor(%v) = %q, want %q", t2, sessionDate, expected)
	}

	// Single-digit month/day - confirms zero-padding
	t3 := time.Date(2026, 9, 5, 10, 0, 0, 0, time.Local)
	expected3 := "2026-09-05"
	sessionDate = SessionDateFor(t3)
	if sessionDate != expected3 {
		t.Errorf("SessionDateFor(%v) = %q, want %q", t3, sessionDate, expected3)
	}
}

func TestGetPendingSessions(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}
	defer store.Close()

	userSlot := 2
	deviceTime1 := uint32(1)
	deviceTime2 := uint32(2)
	deviceTime3 := uint32(3)
	deviceTime4 := uint32(4)
	deviceTime5 := uint32(5)
	deviceTime6 := uint32(6)
	deviceTime7 := uint32(7)
	deviceTime8 := uint32(8)

	morning1 := a6session.Reading{
		Systolic: 120, Diastolic: 80, MeanPressure: 93, Pulse: 70, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime1,
	}
	morning2 := a6session.Reading{
		Systolic: 122, Diastolic: 81, MeanPressure: 94, Pulse: 72, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime2,
	}
	morning3 := a6session.Reading{
		Systolic: 130, Diastolic: 90, MeanPressure: 103, Pulse: 75, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime3,
	}

	morning4 := a6session.Reading{
		Systolic: 135, Diastolic: 95, MeanPressure: 108, Pulse: 78, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime4,
	}
	morning5 := a6session.Reading{
		Systolic: 140, Diastolic: 100, MeanPressure: 113, Pulse: 80, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime5,
	}
	morning6 := a6session.Reading{
		Systolic: 145, Diastolic: 105, MeanPressure: 118, Pulse: 82, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime6,
	}
	morning7 := a6session.Reading{
		Systolic: 150, Diastolic: 110, MeanPressure: 123, Pulse: 85, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime7,
	}
	morning8 := a6session.Reading{
		Systolic: 155, Diastolic: 115, MeanPressure: 128, Pulse: 88, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime8,
	}

	morningTime := time.Date(2026, 8, 18, 9, 0, 0, 0, time.Local)
	morningTime2 := time.Date(2026, 8, 19, 10, 0, 0, 0, time.Local)

	if _, err := store.SaveReading(morning1, morningTime); err != nil {
		t.Fatalf("SaveReading(morning1) returned error: %v", err)
	}
	if _, err := store.SaveReading(morning2, morningTime.Add(time.Minute)); err != nil {
		t.Fatalf("SaveReading(morning2) returned error: %v", err)
	}
	if _, err := store.SaveReading(morning3, morningTime.Add(2*time.Minute)); err != nil {
		t.Fatalf("SaveReading(morning3) returned error: %v", err)
	}
	if _, err := store.SaveReading(morning4, morningTime.Add(3*time.Minute)); err != nil {
		t.Fatalf("SaveReading(morning4) returned error: %v", err)
	}

	if _, err := store.SaveReading(morning5, morningTime2.Add(time.Minute)); err != nil {
		t.Fatalf("SaveReading(morning5) returned error: %v", err)
	}
	if _, err := store.SaveReading(morning6, morningTime2.Add(2*time.Minute)); err != nil {
		t.Fatalf("SaveReading(morning6) returned error: %v", err)
	}
	if _, err := store.SaveReading(morning7, morningTime2.Add(3*time.Minute)); err != nil {
		t.Fatalf("SaveReading(morning7) returned error: %v", err)
	}
	if _, err := store.SaveReading(morning8, morningTime2.Add(4*time.Minute)); err != nil {
		t.Fatalf("SaveReading(morning8) returned error: %v", err)
	}

	pending, err := store.GetPendingSessions()
	if err != nil {
		t.Fatalf("GetPendingSessions() returned error: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("Expected 2 pending sessions, got %d", len(pending))
	}
	if pending[0].SessionType != "morning" || pending[0].SessionDate != "2026-08-18" {
		t.Errorf("pending[0] = %+v, want {morning 2026-08-18}", pending[0])
	}
	if pending[1].SessionType != "morning" || pending[1].SessionDate != "2026-08-19" {
		t.Errorf("pending[1] = %+v, want {morning 2026-08-19}", pending[1])
	}
}

func TestGetPendingSessions_OrdersByActualTime(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}
	defer store.Close()

	userSlot := 2
	deviceTime1 := uint32(1)
	deviceTime2 := uint32(2)

	night := a6session.Reading{
		Systolic: 130, Diastolic: 85, MeanPressure: 100, Pulse: 75, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime1,
	}
	morning := a6session.Reading{
		Systolic: 120, Diastolic: 80, MeanPressure: 93, Pulse: 70, Status: 0,
		UserSlot: &userSlot, DeviceTime: &deviceTime2,
	}

	nightTime := time.Date(2026, 8, 18, 1, 0, 0, 0, time.Local)
	morningTime := time.Date(2026, 8, 18, 9, 0, 0, 0, time.Local)

	if _, err := store.SaveReading(night, nightTime); err != nil {
		t.Fatalf("SaveReading(night) returned error: %v", err)
	}
	if _, err := store.SaveReading(morning, morningTime); err != nil {
		t.Fatalf("SaveReading(morning) returned error: %v", err)
	}

	pending, err := store.GetPendingSessions()
	if err != nil {
		t.Fatalf("GetPendingSessions() returned error: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("got %d pending sessions, want 2", len(pending))
	}
	if pending[0].SessionType != "night" {
		t.Errorf("pending[0].SessionType = %q, want %q, at 1am)", pending[0].SessionType, "night")
	}
	if pending[1].SessionType != "morning" {
		t.Errorf("pending[1].SessionType = %q, want %q", pending[1].SessionType, "morning")
	}
}
