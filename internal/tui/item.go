package tui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/justinpaulosolo/bpmonitor/internal/storage"
)

type readingItem struct {
	reading  storage.StoredReading
	rejected bool
}

func (i readingItem) FilterValue() string { return "" }

func (i readingItem) Title() string {
	text := fmt.Sprintf("#%d  %d/%d  pulse %d", i.reading.ID, i.reading.Systolic, i.reading.Diastolic, i.reading.Pulse)
	if i.rejected {
		return lipgloss.NewStyle().Strikethrough(true).Render(text)
	}
	return text
}

func (i readingItem) Description() string {
	return fmt.Sprintf("%s %s", i.reading.SessionType, i.reading.SessionDate)
}
