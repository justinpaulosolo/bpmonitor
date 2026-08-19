package tui

import (
	"strings"
	"testing"

	"github.com/justinpaulosolo/bpmonitor/internal/a6session"
	"github.com/justinpaulosolo/bpmonitor/internal/storage"
)

func TestReadingItem_Title_Unmarked(t *testing.T) {
	item := readingItem{
		reading: storage.StoredReading{
			ID:      1,
			Reading: a6session.Reading{Systolic: 120, Diastolic: 80, Pulse: 70},
		},
		marked: false,
	}
	if !strings.HasPrefix(item.Title(), "[ ]") {
		t.Errorf("Title() = %q, want prefix %q", item.Title(), "[ ]")
	}
}

func TestReadingItem_Title_Marked(t *testing.T) {
	item := readingItem{
		reading: storage.StoredReading{
			ID:      1,
			Reading: a6session.Reading{Systolic: 120, Diastolic: 80, Pulse: 70},
		},
		marked: true,
	}
	if !strings.HasPrefix(item.Title(), "[x]") {
		t.Errorf("Title() = %q, want prefix %q", item.Title(), "[x]")
	}
}

func TestReadingItem_Description(t *testing.T) {
	item := readingItem{
		reading: storage.StoredReading{
			SessionType: "morning",
			SessionDate: "2026-08-18",
		},
	}
	desc := item.Description()
	if !strings.Contains(desc, "morning") {
		t.Errorf("Description() = %q, want it to contain %q", desc, "morning")
	}
	if !strings.Contains(desc, "2026-08-18") {
		t.Errorf("Description() = %q, want it to contain %q", desc, "2026-08-18")
	}
}
