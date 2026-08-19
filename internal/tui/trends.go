package tui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/NimbleMarkets/ntcharts/v2/linechart/timeserieslinechart"
	"github.com/justinpaulosolo/bpmonitor/internal/storage"
)

func loadCommittedReadings(store *storage.Store, filter string) tea.Cmd {
	return func() tea.Msg {
		readings, err := store.GetCommittedReadings(filter)
		if err != nil {
			return errMsg(err)
		}
		return committedReadingsLoadedMsg{filter: filter, readings: readings}
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
	if sessionType == filterNight {
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
	case filterMorning:
		label = sessionEmoji(filterMorning) + "  Mornings"
	case filterNight:
		label = sessionEmoji(filterNight) + "  Nights"
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
