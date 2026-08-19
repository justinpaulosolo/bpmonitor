package tui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/justinpaulosolo/bpmonitor/internal/a6session"
	"github.com/justinpaulosolo/bpmonitor/internal/storage"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

func TestReadingItem_Title_NotRejected(t *testing.T) {
	item := readingItem{
		reading: storage.StoredReading{
			ID:      1,
			Reading: a6session.Reading{Systolic: 120, Diastolic: 80, Pulse: 70},
		},
		rejected: false,
	}
	want := "#1  120/80  pulse 70"
	if item.Title() != want {
		t.Errorf("Title() = %q, want %q", item.Title(), want)
	}
}

func TestReadingItem_Title_Rejected(t *testing.T) {
	item := readingItem{
		reading: storage.StoredReading{
			ID:      1,
			Reading: a6session.Reading{Systolic: 120, Diastolic: 80, Pulse: 70},
		},
		rejected: true,
	}
	plain := "#1  120/80  pulse 70"
	if item.Title() == plain {
		t.Errorf("Title() = %q, want styled (strikethrough) output, different from plain text", item.Title())
	}
	if stripANSI(item.Title()) != plain {
		t.Errorf("Title() with ANSI codes stripped = %q, want %q", stripANSI(item.Title()), plain)
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
