package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/justinpaulosolo/bpmonitor/internal/storage"
)

func TestUpdate_PendingSessionsLoaded_Empty(t *testing.T) {
	m := Model{}
	newModel, cmd := m.Update(pendingSessionsLoadedMsg{})

	updated := newModel.(Model)
	if updated.currentSession != nil {
		t.Errorf("currentSession = %+v, want nil", updated.currentSession)
	}
	if cmd != nil {
		t.Error("cmd = non-nil, want nil")
	}
}

func TestUpdate_PendingSessionsLoaded_NonEmpty(t *testing.T) {
	m := Model{}
	sessions := pendingSessionsLoadedMsg{
		{SessionType: "morning", SessionDate: "2026-08-18"},
		{SessionType: "night", SessionDate: "2026-08-19"},
	}
	newModel, cmd := m.Update(sessions)

	updated := newModel.(Model)
	if updated.currentSession == nil {
		t.Fatal("currentSession = nil, want non-nil")
	}
	if updated.currentSession.SessionType != "morning" || updated.currentSession.SessionDate != "2026-08-18" {
		t.Errorf("currentSession = %+v, want {morning 2026-08-18}", updated.currentSession)
	}
	if cmd == nil {
		t.Error("cmd = nil, want non-nil")
	}
	if len(updated.pendingSessions) != 2 {
		t.Errorf("pendingSessions has %d entries, want 2", len(updated.pendingSessions))
	}
}

func TestUpdate_PendingReadingsLoaded(t *testing.T) {
	m := Model{}
	readings := pendingReadingsLoadedMsg{
		{ID: 1, SessionType: "morning", SessionDate: "2026-08-18"},
		{ID: 2, SessionType: "morning", SessionDate: "2026-08-18"},
	}
	newModel, cmd := m.Update(readings)

	updated := newModel.(Model)
	if len(updated.pending) != 2 {
		t.Fatalf("pending = %+v, want 2 readings", updated.pending)
	}
	if updated.pending[0].ID != 1 || updated.pending[1].ID != 2 {
		t.Errorf("pending = %+v, want IDs 1 then 2", updated.pending)
	}
	if cmd != nil {
		t.Error("cmd = non-nil, want nil")
	}
	if len(updated.list.Items()) != 2 {
		t.Fatalf("list has %d items, want 2", len(updated.list.Items()))
	}
	item, ok := updated.list.Items()[0].(readingItem)
	if !ok {
		t.Fatalf("list.Items()[0] is %T, want readingItem", updated.list.Items()[0])
	}
	if item.reading.ID != 1 {
		t.Errorf("list.Items()[0].reading.ID = %d, want 1", item.reading.ID)
	}
}

func TestUpdate_PendingSessionsLoaded_EmptyClearsQueue(t *testing.T) {
	m := Model{currentSession: &storage.PendingSession{SessionType: "morning", SessionDate: "2026-08-18"}, pending: []storage.StoredReading{{ID: 1}}, listReady: true}
	newModel, cmd := m.Update(pendingSessionsLoadedMsg{})
	updated := newModel.(Model)
	if updated.currentSession != nil || len(updated.pending) != 0 || updated.listReady {
		t.Fatalf("stale queue state remains: current=%v pending=%v listReady=%v", updated.currentSession, updated.pending, updated.listReady)
	}
	if cmd != nil {
		t.Fatal("cmd = non-nil, want nil")
	}
}

func TestUpdate_CommittedReadingsLoaded_IgnoresStaleFilter(t *testing.T) {
	m := Model{trendsFilter: filterMorning}
	newModel, cmd := m.Update(committedReadingsLoadedMsg{
		filter:   filterNight,
		readings: []storage.StoredReading{{ID: 1}},
	})
	updated := newModel.(Model)
	if len(updated.committed) != 0 {
		t.Fatalf("committed = %+v, want empty after stale response", updated.committed)
	}
	if cmd != nil {
		t.Fatal("cmd = non-nil, want nil")
	}
}

func TestUpdate_QuitKey(t *testing.T) {
	m := Model{}
	_, cmd := m.Update(tea.KeyPressMsg{Text: "q"})

	if cmd == nil {
		t.Fatal("cmd = nil, want tea.Quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("cmd() = %T, want tea.QuitMsg", cmd())
	}
}
