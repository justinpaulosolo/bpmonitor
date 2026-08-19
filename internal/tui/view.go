package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

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
	if m.currentSession != nil && m.listReady {
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

func leftPanelWidth(totalWidth int) int {
	return totalWidth / 3
}

func rightPanelWidth(totalWidth int) int {
	return totalWidth - leftPanelWidth(totalWidth) - 4
}
