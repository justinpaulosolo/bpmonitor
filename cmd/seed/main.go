package main

import (
	"fmt"
	"time"

	"github.com/justinpaulosolo/bpmonitor/internal/a6session"
	"github.com/justinpaulosolo/bpmonitor/internal/storage"
)

// Seeds bpmonitor.db with fixture data for building/testing the TUI without
// the real cuff nearby:
//   - 10 days of history (2026-08-09 .. 2026-08-18), each with a committed
//     morning sitting and a committed night sitting (3 readings each),
//     gently trending values so a chart has something real to show.
//   - Two pending sittings left over on the most recent day, to exercise the
//     queue screen: a morning sitting with 4 readings (reject down to 3),
//     and a night sitting with exactly 1 reading (the "D" abandon-sitting case).
func main() {
	store, err := storage.Open("bpmonitor.db")
	if err != nil {
		panic(err)
	}
	defer store.Close()

	userSlot := 2
	var deviceTime uint32 = 1
	nextDeviceTime := func() *uint32 {
		dt := deviceTime
		deviceTime++
		return &dt
	}

	baseDate := time.Date(2026, 8, 9, 0, 0, 0, 0, time.Local)

	// commitSitting saves n readings at the given base time (one minute apart)
	// with the given trend/noise, then commits them as sessionType/sessionDate.
	commitSitting := func(day int, base time.Time, n int, baseSys, baseDia, basePulse int) {
		ids := make([]int64, 0, n)
		for i := 0; i < n; i++ {
			sys := baseSys - day/2 + i
			dia := baseDia - day/3 + i
			pulse := basePulse + i
			r := a6session.Reading{
				Systolic:     sys,
				Diastolic:    dia,
				MeanPressure: (sys+dia)/2 + 5,
				Pulse:        pulse,
				Status:       0,
				UserSlot:     &userSlot,
				DeviceTime:   nextDeviceTime(),
			}
			id, err := store.SaveReading(r, base.Add(time.Duration(i)*time.Minute))
			if err != nil {
				fmt.Printf("SaveReading(day %d, %v) failed: %v\n", day, base, err)
				continue
			}
			ids = append(ids, id)
		}
		if len(ids) != 3 {
			return // only exactly-3 sittings are committable
		}
		sessionType := storage.SessionTypeFor(base)
		sessionDate := storage.SessionDateFor(base)
		if err := store.CommitReadings(ids, sessionType, sessionDate); err != nil {
			fmt.Printf("CommitReadings(day %d, %s %s) failed: %v\n", day, sessionType, sessionDate, err)
		}
	}

	const historyDays = 10
	for day := 0; day < historyDays; day++ {
		date := baseDate.AddDate(0, 0, day)
		night := time.Date(date.Year(), date.Month(), date.Day(), 1, 0, 0, 0, time.Local)
		morning := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, time.Local)

		commitSitting(day, night, 3, 138, 90, 76)
		commitSitting(day, morning, 3, 128, 84, 70)
	}

	// Leftover pending sittings on the most recent day, for testing the queue.
	lastDate := baseDate.AddDate(0, 0, historyDays-1)
	pendingMorning := time.Date(lastDate.Year(), lastDate.Month(), lastDate.Day(), 10, 0, 0, 0, time.Local)
	pendingNight := time.Date(lastDate.Year(), lastDate.Month(), lastDate.Day(), 2, 0, 0, 0, time.Local)

	for i := 0; i < 4; i++ {
		r := a6session.Reading{
			Systolic: 122 + i, Diastolic: 80 + i, MeanPressure: 100 + i, Pulse: 68 + i, Status: 0,
			UserSlot: &userSlot, DeviceTime: nextDeviceTime(),
		}
		if _, err := store.SaveReading(r, pendingMorning.Add(time.Duration(i)*time.Minute)); err != nil {
			fmt.Printf("SaveReading(pending morning #%d) failed: %v\n", i, err)
		}
	}

	lonely := a6session.Reading{
		Systolic: 131, Diastolic: 87, MeanPressure: 105, Pulse: 74, Status: 0,
		UserSlot: &userSlot, DeviceTime: nextDeviceTime(),
	}
	if _, err := store.SaveReading(lonely, pendingNight); err != nil {
		fmt.Printf("SaveReading(pending night) failed: %v\n", err)
	}
}
