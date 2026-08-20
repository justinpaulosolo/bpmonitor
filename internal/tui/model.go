package tui

import (
	"fmt"
	"sort"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/NimbleMarkets/ntcharts/v2/linechart/timeserieslinechart"
	"github.com/justinpaulosolo/bpmonitor/internal/storage"
)

type Model struct {
	store           *storage.Store
	currentSession  *storage.PendingSession
	pendingSessions []storage.PendingSession
	pending         []storage.StoredReading
	list            list.Model
	width, height   int
	listReady       bool
	focus           Focus
	rejectedStack   []storage.StoredReading
	err             error
	committed       []storage.StoredReading
	trendsFilter    string
	trendsTable     table.Model
	trendsChart     timeserieslinechart.Model
}

type pendingSessionsLoadedMsg []storage.PendingSession
type pendingReadingsLoadedMsg []storage.StoredReading
type errMsg error
type sessionCommittedMsg struct{}
type sessionAbandonedMsg struct{}
type readingRejectedMsg storage.StoredReading
type readingRestoredMsg storage.StoredReading
type committedReadingsLoadedMsg struct {
	filter   string
	readings []storage.StoredReading
}

type Focus int

const (
	focusQueue Focus = iota
	focusTrends
)

const (
	filterAll     = ""
	filterMorning = "morning"
	filterNight   = "night"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadPendingSessions(m.store),
		loadCommittedReadings(m.store, m.trendsFilter),
	)
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
	return Model{
		store:       store,
		trendsTable: newCommittedTable(1, 1),
		trendsChart: buildTrendsChart(nil, 1, chartHeight),
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pendingSessionsLoadedMsg:
		m.pendingSessions = msg
		m.rejectedStack = nil
		m.pending = nil
		m.listReady = false
		m.err = nil
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
		m.err = nil
		m.rebuildList()
		return m, nil
	case readingRestoredMsg:
		reading := storage.StoredReading(msg)
		m.pending = append(m.pending, reading)
		if len(m.rejectedStack) > 0 {
			m.rejectedStack = m.rejectedStack[:len(m.rejectedStack)-1]
		}
		m.rebuildList()
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
		m.rebuildList()
		return m, nil
	case sessionCommittedMsg:
		return m, tea.Batch(loadPendingSessions(m.store), loadCommittedReadings(m.store, m.trendsFilter))
	case sessionAbandonedMsg:
		return m, loadPendingSessions(m.store)
	case committedReadingsLoadedMsg:
		if msg.filter != m.trendsFilter {
			return m, nil
		}
		m.committed = msg.readings
		m.err = nil
		m.trendsTable.SetRows(committedRows(m.committed))
		m.trendsChart = buildTrendsChart(m.committed, contentWidth(rightPanelWidth(m.width)), chartHeight)
		return m, nil
	case errMsg:
		m.err = msg
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.listReady {
			m.list.SetSize(contentWidth(leftPanelWidth(msg.Width)), contentHeight(panelHeight(msg.Height)))
		}
		m.trendsTable.SetWidth(contentWidth(rightPanelWidth(msg.Width)))
		m.trendsTable.SetColumns(committedColumns(contentWidth(rightPanelWidth(msg.Width))))
		m.trendsTable.SetHeight(trendsTableHeight(panelHeight(msg.Height)))
		m.trendsChart.Resize(contentWidth(rightPanelWidth(msg.Width)), chartHeight)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.focus == focusQueue {
				m.focus = focusTrends
				m.trendsTable.Focus()
			} else {
				m.focus = focusQueue
				m.trendsTable.Blur()
			}
			return m, nil
		}
		if m.focus == focusTrends {
			switch msg.String() {
			case "a":
				m.trendsFilter = filterAll
				return m, loadCommittedReadings(m.store, m.trendsFilter)
			case "m":
				m.trendsFilter = filterMorning
				return m, loadCommittedReadings(m.store, m.trendsFilter)
			case "n":
				m.trendsFilter = filterNight
				return m, loadCommittedReadings(m.store, m.trendsFilter)
			}
			var cmd tea.Cmd
			m.trendsTable, cmd = m.trendsTable.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "x":
			item, ok := m.list.SelectedItem().(readingItem)
			if !ok || item.rejected {
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
				count := len(m.pending)
				return m, func() tea.Msg {
					return errMsg(fmt.Errorf("need exactly 3 to commit, have %d", count))
				}
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
		case "D":
			if len(m.pending) == 0 {
				return m, nil
			}
			ids := make([]int64, len(m.pending))
			for i, r := range m.pending {
				ids[i] = r.ID
			}
			return m, func() tea.Msg {
				for _, id := range ids {
					if err := m.store.DeleteReading(id); err != nil {
						return errMsg(err)
					}
				}
				return sessionAbandonedMsg{}
			}
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
	return m, nil
}

// combinedListItems merges the active (pending) readings with whatever's been
// rejected so far this sitting, so rejected readings stay visible (struck
// through) in their original position instead of disappearing outright.
func combinedListItems(pending, rejectedStack []storage.StoredReading) []list.Item {
	type entry struct {
		reading  storage.StoredReading
		rejected bool
	}
	all := make([]entry, 0, len(pending)+len(rejectedStack))
	for _, r := range pending {
		all = append(all, entry{reading: r})
	}
	for _, r := range rejectedStack {
		all = append(all, entry{reading: r, rejected: true})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].reading.ID < all[j].reading.ID })

	items := make([]list.Item, len(all))
	for i, e := range all {
		items[i] = readingItem{reading: e.reading, rejected: e.rejected}
	}
	return items
}

func (m *Model) rebuildList() {
	m.list = newReadingList(combinedListItems(m.pending, m.rejectedStack), contentWidth(leftPanelWidth(m.width)), contentHeight(panelHeight(m.height)))
	m.listReady = true
}

func newReadingList(items []list.Item, width, height int) list.Model {
	d := list.NewDefaultDelegate()
	d.Styles.NormalTitle = d.Styles.NormalTitle.Foreground(colorText)
	d.Styles.NormalDesc = d.Styles.NormalDesc.Foreground(colorTextDim)
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(colorSelectedFg).BorderForeground(colorBorderFocus)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(colorTextDim).BorderForeground(colorBorderFocus)

	l := list.New(items, d, width, height)
	l.SetShowHelp(false)
	l.Styles.Title = l.Styles.Title.Foreground(colorSelectedFg).Background(colorSelectedBg)
	l.Styles.PaginationStyle = l.Styles.PaginationStyle.Foreground(colorTextDim)
	l.Styles.ActivePaginationDot = l.Styles.ActivePaginationDot.Foreground(colorAccent)
	l.Styles.InactivePaginationDot = l.Styles.InactivePaginationDot.Foreground(colorBorder)
	l.Styles.NoItems = l.Styles.NoItems.Foreground(colorTextDim)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "reject")),
			key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "undo")),
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "commit")),
			key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "abandon sitting")),
		}
	}
	return l
}
