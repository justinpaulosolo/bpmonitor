package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	focus           Focus
	rejectedStack   []storage.StoredReading
	err             error
}

type pendingSessionsLoadedMsg []storage.PendingSession
type pendingReadingsLoadedMsg []storage.StoredReading
type errMsg error
type readingsDeletedMsg []int64
type sessionCommittedMsg struct{}
type readingRejectedMsg storage.StoredReading
type readingRestoredMsg storage.StoredReading

// A screen type (int, or a small named type) with two values: screenQueue, screenTrends
type Screen int
type Focus int

const (
	focusQueue Focus = iota
	focusTrends
)

const (
	screenQueue Screen = iota
	screenTrends
)

func (m Model) Init() tea.Cmd {
	return loadPendingSessions(m.store)
}

func loadPendingSessions(store *storage.Store) tea.Cmd {
	return func() tea.Msg {
		sessions, err := store.GetPendingSessions()
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
		m.pendingSessions = msg
		m.rejectedStack = nil
		if len(msg) == 0 {
			m.currentSession = nil
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
		m.list = newReadingList(items, leftPanelWidth(m.width), panelHeight(m.height))
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
		m.list = newReadingList(toListItems(remaining), leftPanelWidth(m.width), panelHeight(m.height))
		m.listReady = true
		return m, nil
	case readingRestoredMsg:
		reading := storage.StoredReading(msg)
		m.pending = append(m.pending, reading)
		if len(m.rejectedStack) > 0 {
			m.rejectedStack = m.rejectedStack[:len(m.rejectedStack)-1]
		}
		m.list = newReadingList(toListItems(m.pending), leftPanelWidth(m.width), panelHeight(m.height))
		m.listReady = true
		return m, nil
	case readingRejectedMsg:
		reading := storage.StoredReading(msg)
		var remaining []storage.StoredReading
		for _, r := range m.pending {
			if r.ID != reading.ID {
				remaining = append(remaining, r)
			}
		}
		m.pending = remaining
		m.rejectedStack = append(m.rejectedStack, reading)
		m.list = newReadingList(toListItems(remaining), leftPanelWidth(m.width), panelHeight(m.height))
		m.listReady = true
		return m, nil
	case sessionCommittedMsg:
		return m, loadPendingSessions(m.store)
	case errMsg:
		m.err = msg
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.listReady {
			m.list.SetSize(leftPanelWidth(msg.Width), panelHeight(msg.Height))
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.focus == focusQueue {
				m.focus = focusTrends
			} else {
				m.focus = focusQueue
			}
			return m, nil
		}
		if m.focus != focusQueue {
			return m, nil
		}
		switch msg.String() {
		case "space":
			if item, ok := m.list.SelectedItem().(readingItem); ok {
				item.marked = !item.marked
				m.list.SetItem(m.list.Index(), item)
			}
			return m, nil
		case "x":
			item, ok := m.list.SelectedItem().(readingItem)
			if !ok {
				return m, nil
			}
			reading := item.reading
			return m, func() tea.Msg {
				if err := m.store.RejectReading(reading.ID); err != nil {
					return errMsg(err)
				}
				return readingRejectedMsg(reading)
			}
		case "c":
			if len(m.pending) != 3 {
				return m, nil
			}
			ids := make([]int64, len(m.pending))
			for i, r := range m.pending {
				ids[i] = r.ID
			}
			sessionType := m.currentSession.SessionType
			sessionDate := m.currentSession.SessionDate
			return m, func() tea.Msg {
				if err := m.store.CommitReadings(ids, sessionType, sessionDate); err != nil {
					return errMsg(err)
				}
				return sessionCommittedMsg{}
			}
		case "u":
			if len(m.rejectedStack) == 0 {
				return m, nil
			}
			reading := m.rejectedStack[len(m.rejectedStack)-1]
			return m, func() tea.Msg {
				if err := m.store.RestoreReading(reading.ID); err != nil {
					return errMsg(err)
				}
				return readingRestoredMsg(reading)
			}
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
	return m, nil
}

const chromeLines = 4

func panelHeight(termHeight int) int {
	h := termHeight - chromeLines
	if h < 1 {
		return 1
	}
	return h
}

func panelStyle(focused bool, width, height int) lipgloss.Style {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(width).
		Height(height)
	if focused {
		style = style.BorderForeground(lipgloss.Color("205"))
	}
	return style
}

func (m Model) View() tea.View {
	leftWidth := leftPanelWidth(m.width)
	rightWidth := rightPanelWidth(m.width)
	ph := panelHeight(m.height)

	queueBody := "No pending readings."
	if m.currentSession != nil {
		queueBody = m.list.View()
	}

	left := panelStyle(m.focus == focusQueue, leftWidth, ph).Render(queueBody)
	rightContentHeight := lipgloss.Height(left)
	right := panelStyle(m.focus == focusTrends, rightWidth, rightContentHeight).Render("Trends\n\n(coming soon)")

	footerText := "↑/k up  ↓/j down  space: keep  x: reject  u: undo  c: commit   |   tab: switch panel   q: quit"
	if m.err != nil {
		footerText = fmt.Sprintf("error: %v", m.err) + "\n" + footerText
	}
	if len(m.pendingSessions) > 1 {
		footerText += fmt.Sprintf("   |   %d more sitting(s) pending review", len(m.pendingSessions)-1)
	}
	footer := "\n\n" + lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(footerText)

	v := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, left, right) + footer)
	v.AltScreen = true
	return v
}

func toListItems(readings []storage.StoredReading) []list.Item {
	items := make([]list.Item, len(readings))
	for i, r := range readings {
		items[i] = readingItem{reading: r}
	}
	return items
}

func newReadingList(items []list.Item, width, height int) list.Model {
	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.SetShowHelp(false)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "keep")),
			key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "reject")),
			key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "undo")),
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "commit")),
		}
	}
	return l
}

func leftPanelWidth(totalWidth int) int {
	return totalWidth / 3
}

func rightPanelWidth(totalWidth int) int {
	return totalWidth - leftPanelWidth(totalWidth) - 4
}
