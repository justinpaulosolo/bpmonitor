package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/NimbleMarkets/ntcharts/v2/linechart/timeserieslinechart"
	"github.com/charmbracelet/x/ansi"
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
type committedReadingsLoadedMsg []storage.StoredReading

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
		screen:      screenQueue,
		trendsTable: newCommittedTable(1, 1),
		trendsChart: buildTrendsChart(nil, 1, chartHeight),
	}
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
		m.list = newReadingList(combinedListItems(m.pending, m.rejectedStack), contentWidth(leftPanelWidth(m.width)), contentHeight(panelHeight(m.height)))
		m.listReady = true
		return m, nil
	case readingRestoredMsg:
		reading := storage.StoredReading(msg)
		m.pending = append(m.pending, reading)
		if len(m.rejectedStack) > 0 {
			m.rejectedStack = m.rejectedStack[:len(m.rejectedStack)-1]
		}
		m.list = newReadingList(combinedListItems(m.pending, m.rejectedStack), contentWidth(leftPanelWidth(m.width)), contentHeight(panelHeight(m.height)))
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
		m.list = newReadingList(combinedListItems(m.pending, m.rejectedStack), contentWidth(leftPanelWidth(m.width)), contentHeight(panelHeight(m.height)))
		m.listReady = true
		return m, nil
	case sessionCommittedMsg:
		return m, tea.Batch(loadPendingSessions(m.store), loadCommittedReadings(m.store, m.trendsFilter))
	case sessionAbandonedMsg:
		return m, loadPendingSessions(m.store)
	case committedReadingsLoadedMsg:
		m.committed = msg
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
				m.trendsFilter = ""
				return m, loadCommittedReadings(m.store, m.trendsFilter)
			case "m":
				m.trendsFilter = "morning"
				return m, loadCommittedReadings(m.store, m.trendsFilter)
			case "n":
				m.trendsFilter = "night"
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

const chromeLines = 4

func panelHeight(termHeight int) int {
	h := termHeight - chromeLines
	if h < 1 {
		return 1
	}
	return h
}

// lipgloss's Width()/Height() on a bordered style are *total* dimensions —
// they include the border rows/columns — and they only ever *pad* content
// that's too small, never clip content that's too big. So the usable content
// area is 2 smaller in each direction, and anything larger than that must be
// clipped by us beforehand.
func contentWidth(panelWidth int) int {
	if panelWidth-2 < 1 {
		return 1
	}
	return panelWidth - 2
}

func contentHeight(panelHeight int) int {
	if panelHeight-2 < 1 {
		return 1
	}
	return panelHeight - 2
}

// clip trims content to exactly fit within width x height *before* it goes into
// a bordered panel. Without this, content too wide for the panel gets *wrapped*,
// and each wrapped line adds a row that Height() won't trim — pushing the panel
// (and its bottom border) past the bottom of the terminal. Clipping the outer
// box instead doesn't work: that just cuts off the border itself.
func clip(s string, width, height int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "")
	}
	return strings.Join(lines, "\n")
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

	left := panelStyle(m.focus == focusQueue, leftWidth, ph).
		Render(clip(queueBody, contentWidth(leftWidth), contentHeight(ph)))
	summary := trendsSummary(m.committed, m.trendsFilter, contentWidth(rightWidth))
	rightBody := m.trendsChart.View() + "\n\n" + summary + "\n\n" + m.trendsTable.View()
	right := panelStyle(m.focus == focusTrends, rightWidth, ph).
		Render(clip(rightBody, contentWidth(rightWidth), contentHeight(ph)))

	footerText := "↑↓ move · x reject · u undo · c commit · D abandon · a/m/n filter · tab panel · q quit"
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

func newReadingList(items []list.Item, width, height int) list.Model {
	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.SetShowHelp(false)
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

func leftPanelWidth(totalWidth int) int {
	return totalWidth / 3
}

func rightPanelWidth(totalWidth int) int {
	return totalWidth - leftPanelWidth(totalWidth) - 4
}

func loadCommittedReadings(store *storage.Store, filter string) tea.Cmd {
	return func() tea.Msg {
		readings, err := store.GetCommittedReadings(filter)
		if err != nil {
			return errMsg(err)
		}
		return committedReadingsLoadedMsg(readings)
	}
}

const chartHeight = 12

// trendsSummaryLines is the summary banner's total rendered height, borders
// included — lipgloss Width/Height are total dimensions, so Height(3) yields a
// 3-line box with 1 line of content between its two border rows.
const trendsSummaryLines = 3

// trendsTableHeight returns how much of the panel's *content* area is left for
// the readings table once the fixed-height chart, the summary block, and the
// two blank separator lines between them are accounted for.
func trendsTableHeight(ph int) int {
	h := contentHeight(ph) - chartHeight - trendsSummaryLines - 2
	if h < 1 {
		return 1
	}
	return h
}

// committedColumns spreads the table across the full available width,
// distributing it proportionally rather than dumping all the slack into one
// column. Each cell carries 1 column of padding on each side (bubbles' default
// table styles), hence the cellPadding accounting.
func committedColumns(width int) []table.Column {
	const (
		cellPadding = 2
		numCols     = 4
	)
	usable := width - (numCols * cellPadding)
	if usable < 24 {
		usable = 24
	}
	// ~40% for the date/session column, the rest split evenly across the three
	// numeric columns.
	whenW := usable * 40 / 100
	rest := (usable - whenW) / 3
	if rest < 5 {
		rest = 5
	}
	return []table.Column{
		{Title: "When", Width: whenW},
		{Title: "Sys", Width: rest},
		{Title: "Dia", Width: rest},
		{Title: "Pulse", Width: usable - whenW - 2*rest},
	}
}

func newCommittedTable(width, height int) table.Model {
	t := table.New(
		table.WithColumns(committedColumns(width)),
		table.WithWidth(width),
		table.WithHeight(height),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return t
}

func sessionEmoji(sessionType string) string {
	if sessionType == "night" {
		return "🌙"
	}
	return "☀️"
}

func committedRows(committed []storage.StoredReading) []table.Row {
	rows := make([]table.Row, len(committed))
	for i, r := range committed {
		rows[i] = table.Row{
			fmt.Sprintf("%s  %s", sessionEmoji(r.SessionType), r.SessionDate),
			fmt.Sprintf("%d", r.Systolic),
			fmt.Sprintf("%d", r.Diastolic),
			fmt.Sprintf("%d", r.Pulse),
		}
	}
	return rows
}

func buildTrendsChart(committed []storage.StoredReading, width, height int) timeserieslinechart.Model {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	// Without an explicit time/Y range, the chart has no idea what range of
	// dates/values to expect, and points collapse into far too few columns —
	// multiple points landing in the same column get averaged together, which
	// is why the line looked "stuck" at one flat value instead of trending.
	minT, maxT := time.Now(), time.Now()
	minY, maxY := 0.0, 200.0
	first := true
	for _, r := range committed {
		t, err := time.Parse("2006-01-02", r.SessionDate)
		if err != nil {
			continue
		}
		if first {
			minT, maxT = t, t
			minY, maxY = float64(r.Diastolic), float64(r.Systolic)
			first = false
		}
		if t.Before(minT) {
			minT = t
		}
		if t.After(maxT) {
			maxT = t
		}
		if float64(r.Diastolic) < minY {
			minY = float64(r.Diastolic)
		}
		if float64(r.Systolic) > maxY {
			maxY = float64(r.Systolic)
		}
	}
	if minT.Equal(maxT) {
		maxT = maxT.AddDate(0, 0, 1) // avoid a zero-width time range
	}

	chart := timeserieslinechart.New(width, height,
		timeserieslinechart.WithTimeRange(minT, maxT),
		timeserieslinechart.WithYRange(minY-5, maxY+5),
	)
	chart.SetDataSetStyle("systolic", lipgloss.NewStyle().Foreground(lipgloss.Color("205")))
	chart.SetDataSetStyle("diastolic", lipgloss.NewStyle().Foreground(lipgloss.Color("39")))
	for _, r := range committed {
		t, err := time.Parse("2006-01-02", r.SessionDate)
		if err != nil {
			continue
		}
		chart.PushDataSet("systolic", timeserieslinechart.TimePoint{Time: t, Value: float64(r.Systolic)})
		chart.PushDataSet("diastolic", timeserieslinechart.TimePoint{Time: t, Value: float64(r.Diastolic)})
	}
	chart.DrawDataSets([]string{"systolic", "diastolic"})
	return chart
}

// trendsSummary renders the filter label + averages as a banner. It is always
// exactly trendsSummaryLines lines tall regardless of data — the panel height
// budgeting depends on that.
func trendsSummary(committed []storage.StoredReading, filter string, width int) string {
	label := "All readings"
	switch filter {
	case "morning":
		label = sessionEmoji("morning") + "  Mornings"
	case "night":
		label = sessionEmoji("night") + "  Nights"
	}

	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	statStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	var body string
	if len(committed) == 0 {
		body = labelStyle.Render(label) + "   " + dimStyle.Render("no committed readings yet")
	} else {
		var sysSum, diaSum, pulseSum int
		for _, r := range committed {
			sysSum += r.Systolic
			diaSum += r.Diastolic
			pulseSum += r.Pulse
		}
		n := len(committed)
		body = labelStyle.Render(label) + "   " +
			dimStyle.Render("AVG ") + statStyle.Render(fmt.Sprintf("%d/%d", sysSum/n, diaSum/n)) +
			dimStyle.Render("   PULSE ") + statStyle.Render(fmt.Sprintf("%d", pulseSum/n)) +
			dimStyle.Render(fmt.Sprintf("   n=%d", n))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(width).
		Height(trendsSummaryLines).
		Padding(0, 1).
		Render(body)
}
