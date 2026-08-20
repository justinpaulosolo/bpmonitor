package tui

import "charm.land/lipgloss/v2"

// Tokyo Night ("night" variant) raw palette, from folke/tokyonight.nvim.
// These are the *source* colors — nothing outside this file should reference
// them directly. Everything else uses the semantic roles below, so a future
// theme swap only has to re-point the roles.
const (
	tnBg          = "#1a1b26"
	tnBgDark      = "#16161e"
	tnBgHighlight = "#292e42"
	tnFg          = "#c0caf5"
	tnFgDark      = "#a9b1d6"
	tnFgGutter    = "#3b4261"
	tnComment     = "#565f89"
	tnDark3       = "#545c7e"
	tnDark5       = "#737aa2"
	tnBlue        = "#7aa2f7"
	tnBlue0       = "#3d59a1"
	tnBlue7       = "#394b70"
	tnCyan        = "#7dcfff"
	tnMagenta     = "#bb9af7"
	tnPurple      = "#9d7cd8"
	tnOrange      = "#ff9e64"
	tnYellow      = "#e0af68"
	tnGreen       = "#9ece6a"
	tnTeal        = "#73daca"
	tnRed         = "#f7768e"
	tnRed1        = "#db4b4b"
)

// Semantic roles. Every style in the TUI pulls from here.
var (
	// colorBorder is the border of an *unfocused* panel.
	colorBorder = lipgloss.Color(tnFgGutter)
	// colorBorderFocus is the border of the panel that currently has focus.
	colorBorderFocus = lipgloss.Color(tnBlue)
	// colorText is normal body text (list titles, table cells).
	colorText = lipgloss.Color(tnFg)
	// colorTextDim is de-emphasized text: labels, hints, the footer.
	colorTextDim = lipgloss.Color(tnComment)
	// colorAccent is the highlight color for headings and the filter label.
	colorAccent = lipgloss.Color(tnMagenta)
	// colorStat is the emphasized numbers in the averages banner.
	colorStat = lipgloss.Color(tnFg)
	// colorSelectedFg / colorSelectedBg are the selected row in the table and
	// the selected item in the queue list.
	colorSelectedFg = lipgloss.Color(tnFg)
	colorSelectedBg = lipgloss.Color(tnBlue7)
	// colorSystolic / colorDiastolic are the two chart series. Red/blue splits
	// on both hue and lightness, so the two lines stay distinguishable.
	colorSystolic  = lipgloss.Color(tnRed)
	colorDiastolic = lipgloss.Color(tnBlue)
	// colorError is the error line in the footer.
	colorError = lipgloss.Color(tnRed1)
	// colorRejected is a struck-through, rejected reading in the queue.
	colorRejected = lipgloss.Color(tnComment)
)
