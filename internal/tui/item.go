package tui

import (
	"fmt"

	"github.com/justinpaulosolo/bpmonitor/internal/storage"
)

type readingItem struct {
	reading storage.StoredReading
	marked  bool
}

func (i readingItem) FilterValue() string { return "" }

func (i readingItem) Title() string {
	mark := "[ ]"
	if i.marked {
		mark = "[x]"
	}
	return fmt.Sprintf("%s #%d  %d/%d  pulse %d", mark, i.reading.ID, i.reading.Systolic, i.reading.Diastolic, i.reading.Pulse)
}

func (i readingItem) Description() string {
	return fmt.Sprintf("%s %s", i.reading.SessionType, i.reading.SessionDate)
}
