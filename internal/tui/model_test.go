package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
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
