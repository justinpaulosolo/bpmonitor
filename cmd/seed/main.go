package main

import (
	"fmt"
	"time"

	"github.com/justinpaulosolo/bpmonitor/internal/a6session"
	"github.com/justinpaulosolo/bpmonitor/internal/storage"
)

func main() {
	/*
			- Opens storage.Open("bpmonitor.db")
		- Builds a handful of a6session.Reading values (varied systolic/diastolic/pulse, each with a distinct DeviceTime so dedup doesn't collapse them) across a few different scenarios, to exercise both TUI screens:
		  - A "night" batch of 5 pending readings, left uncommitted — good for testing "delete the bad ones, then commit 3" on the queue screen
		  - A "morning" batch of exactly 3 pending readings, left uncommitted — the simpler "just commit these" queue case
		  - A second "morning" batch of 4, where you also call CommitReadings on 3 of them right in the seed script — simulates an already-reviewed sitting, giving the trends screen something real to show
		- Uses time.Date(...) (not time.Now()) for each reading's timestamp, so runs are reproducible and you control which hour lands in which session type via SessionTypeFor's boundary
	*/
	store, err := storage.Open("bpmonitor.db")
	if err != nil {
		panic(err)
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

	deviceTime9 := uint32(9)
	deviceTime10 := uint32(10)
	deviceTime11 := uint32(11)
	deviceTime12 := uint32(12)

	nightbatch1 := a6session.Reading{
		Systolic:   120,
		Diastolic:  80,
		Pulse:      70,
		UserSlot:   &userSlot,
		DeviceTime: &deviceTime1,
	}
	nightbatch2 := a6session.Reading{
		Systolic:   125,
		Diastolic:  82,
		Pulse:      72,
		UserSlot:   &userSlot,
		DeviceTime: &deviceTime2,
	}
	nightbatch3 := a6session.Reading{
		Systolic:   130,
		Diastolic:  85,
		Pulse:      75,
		UserSlot:   &userSlot,
		DeviceTime: &deviceTime3,
	}
	nightbatch4 := a6session.Reading{
		Systolic:   135,
		Diastolic:  88,
		Pulse:      78,
		UserSlot:   &userSlot,
		DeviceTime: &deviceTime4,
	}
	nightbatch5 := a6session.Reading{
		Systolic:   140,
		Diastolic:  90,
		Pulse:      80,
		UserSlot:   &userSlot,
		DeviceTime: &deviceTime5,
	}

	morningBatch1 := a6session.Reading{
		Systolic:   115,
		Diastolic:  75,
		Pulse:      68,
		UserSlot:   &userSlot,
		DeviceTime: &deviceTime6,
	}
	morningBatch2 := a6session.Reading{
		Systolic:   118,
		Diastolic:  77,
		Pulse:      70,
		UserSlot:   &userSlot,
		DeviceTime: &deviceTime7,
	}
	morningBatch3 := a6session.Reading{
		Systolic:   120,
		Diastolic:  78,
		Pulse:      72,
		UserSlot:   &userSlot,
		DeviceTime: &deviceTime8,
	}

	secondMorningBatch1 := a6session.Reading{
		Systolic:   122,
		Diastolic:  80,
		Pulse:      74,
		UserSlot:   &userSlot,
		DeviceTime: &deviceTime9,
	}
	secondMorningBatch2 := a6session.Reading{
		Systolic:   124,
		Diastolic:  82,
		Pulse:      76,
		UserSlot:   &userSlot,
		DeviceTime: &deviceTime10,
	}
	secondMorningBatch3 := a6session.Reading{
		Systolic:   126,
		Diastolic:  84,
		Pulse:      78,
		UserSlot:   &userSlot,
		DeviceTime: &deviceTime11,
	}
	secondMorningBatch4 := a6session.Reading{
		Systolic:   128,
		Diastolic:  86,
		Pulse:      80,
		UserSlot:   &userSlot,
		DeviceTime: &deviceTime12,
	}

	nightTime := time.Date(2026, 8, 18, 1, 0, 0, 0, time.Local)

	if _, err := store.SaveReading(nightbatch1, nightTime); err != nil {
		fmt.Printf("SaveReading(nightbatch1, nightTime) failed: %v", err)
	}
	if _, err := store.SaveReading(nightbatch2, nightTime.Add(time.Minute)); err != nil {
		fmt.Printf("SaveReading(nightbatch2, nightTime) failed: %v", err)
	}
	if _, err := store.SaveReading(nightbatch3, nightTime.Add(2*time.Minute)); err != nil {
		fmt.Printf("SaveReading(nightbatch3, nightTime) failed: %v", err)
	}
	if _, err := store.SaveReading(nightbatch4, nightTime.Add(3*time.Minute)); err != nil {
		fmt.Printf("SaveReading(nightbatch4, nightTime) failed: %v", err)
	}
	if _, err := store.SaveReading(nightbatch5, nightTime.Add(4*time.Minute)); err != nil {
		fmt.Printf("SaveReading(nightbatch5, nightTime) failed: %v", err)
	}

	morningTime := time.Date(2026, 8, 18, 9, 0, 0, 0, time.Local)

	if _, err := store.SaveReading(morningBatch1, morningTime); err != nil {
		fmt.Printf("SaveReading(morningBatch1, morningTime) failed: %v", err)
	}
	if _, err := store.SaveReading(morningBatch2, morningTime.Add(time.Minute)); err != nil {
		fmt.Printf("SaveReading(morningBatch2, morningTime) failed: %v", err)
	}
	if _, err := store.SaveReading(morningBatch3, morningTime.Add(2*time.Minute)); err != nil {
		fmt.Printf("SaveReading(morningBatch3, morningTime) failed: %v\n", err)
	}

	morningTime2 := time.Date(2026, 8, 19, 10, 0, 0, 0, time.Local)

	morningTime2_1, err := store.SaveReading(secondMorningBatch1, morningTime2)
	if err != nil {
		fmt.Printf("SaveReading(secondMorningBatch1, morningTime2) failed: %v", err)
	}
	morningTime2_2, err := store.SaveReading(secondMorningBatch2, morningTime2.Add(time.Minute))
	if err != nil {
		fmt.Printf("SaveReading(secondMorningBatch2, morningTime2) failed: %v", err)
	}
	morningTime2_3, err := store.SaveReading(secondMorningBatch3, morningTime2.Add(2*time.Minute))
	if err != nil {
		fmt.Printf("SaveReading(secondMorningBatch3, morningTime2) failed: %v", err)
	}

	if err := store.CommitReadings([]int64{morningTime2_1, morningTime2_2, morningTime2_3}, "morning", "2026-08-19"); err != nil {
		fmt.Printf("CommitReadings failed: %v", err)
	}
	if _, err := store.SaveReading(secondMorningBatch4, morningTime2.Add(3*time.Minute)); err != nil {
		fmt.Printf("SaveReading(secondMorningBatch4, morningTime2) failed: %v", err)
	}
}
