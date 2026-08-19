package tui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/justinpaulosolo/bpmonitor/internal/storage"
)

type Model struct {
	store           *storage.Store
	screen          Screen
	currentSession  *storage.PendingSession
	pendingSessions []storage.PendingSession
	pending         []storage.StoredReading
	list            list.Model
	width, height   int
	listReady       bool
}

type pendingSessionsLoadedMsg []storage.PendingSession
type pendingReadingsLoadedMsg []storage.StoredReading
type errMsg error
type readingsDeletedMsg []int64

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

func NewModel(store *storage.Store) Model {
	return Model{store: store, screen: screenQueue}
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
		items := toListItems(msg)
		m.list = list.New(items, list.NewDefaultDelegate(), m.width, m.height)
		m.listReady = true
		return m, nil
	case readingsDeletedMsg:
		deleted := make(map[int64]bool, len(msg))
		for _, id := range msg {
			deleted[id] = true
		}
		var remaining []storage.StoredReading
		for _, r := range m.pending {
			if !deleted[r.ID] {
				remaining = append(remaining, r)
			}
		}
		m.pending = remaining
		m.list = list.New(toListItems(remaining), list.NewDefaultDelegate(), m.width, m.height)
		m.listReady = true
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.listReady {
			m.list.SetSize(msg.Width, msg.Height)
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "space":
			if item, ok := m.list.SelectedItem().(readingItem); ok {
				item.marked = !item.marked
				m.list.SetItem(m.list.Index(), item)
			}
			return m, nil
		case "x":
			var ids []int64
			for _, it := range m.list.Items() {
				if ri, ok := it.(readingItem); ok && ri.marked {
					ids = append(ids, ri.reading.ID)
				}
			}
			if len(ids) == 0 {
				return m, nil
			}
			return m, func() tea.Msg {
				for _, id := range ids {
					if err := m.store.DeleteReading(id); err != nil {
						return errMsg(err)
					}
				}
				return readingsDeletedMsg(ids)
			}
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) View() tea.View {
	var v tea.View
	if m.currentSession == nil {
		v = tea.NewView("No pending readings.\n\nq to quit")
	} else {
		v = tea.NewView(m.list.View() + "\n\nspace: mark/unmark   x: delete marked   c: commit (when 3 remain)   q: quit")
	}
	v.AltScreen = true
	return v
	// if m.currentSession == nil {
	// 	return tea.NewView("No pending readings.\n\nq to quit")
	// }
	// return tea.NewView(m.list.View() + "\n\nd: mark/unmark   x: delete marked   c: commit (when 3 remain)   q: quit")
}

func toListItems(readings []storage.StoredReading) []list.Item {
	items := make([]list.Item, len(readings))
	for i, r := range readings {
		items[i] = readingItem{reading: r}
	}
	return items
}
