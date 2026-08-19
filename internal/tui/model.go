package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/justinpaulosolo/bpmonitor/internal/storage"
)

type Model struct {
	store           *storage.Store
	screen          Screen
	currentSession  *storage.PendingSession
	pendingSessions []storage.PendingSession
	pending         []storage.StoredReading
}

type pendingSessionsLoadedMsg []storage.PendingSession
type pendingReadingsLoadedMsg []storage.StoredReading
type errMsg error

// A screen type (int, or a small named type) with two values: screenQueue, screenTrends
type Screen int

const (
	screenQueue Screen = iota
	screenTrends
)

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		sessions, err := m.store.GetPendingSessions()
		if err != nil {
			return errMsg(err)
		}
		return pendingSessionsLoadedMsg(sessions)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pendingSessionsLoadedMsg:
		if len(msg) == 0 {
			return m, nil
		}
		m.currentSession = &msg[0]
		return m, func() tea.Msg {
			readings, err := m.store.GetPendingReadings(msg[0].SessionType, msg[0].SessionDate)
			if err != nil {
				return errMsg(err)
			}
			return pendingReadingsLoadedMsg(readings)
		}
	case pendingReadingsLoadedMsg:
		m.pending = msg
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	return tea.NewView("")
}
