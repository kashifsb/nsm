package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette
	primaryColor    = lipgloss.Color("#7C3AED") // Purple
	successColor    = lipgloss.Color("#10B981") // Green
	warningColor    = lipgloss.Color("#F59E0B") // Yellow
	errorColor      = lipgloss.Color("#EF4444") // Red
	accentColor     = lipgloss.Color("#06B6D4") // Cyan
	mutedColor      = lipgloss.Color("#6B7280") // Gray
	backgroundColor = lipgloss.Color("#0F172A") // Dark blue

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Margin(1, 0)

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 3).
			Margin(1, 0)

	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Background(lipgloss.Color("#1E1B4B")).
			Padding(0, 2).
			Margin(0, 1)

	// Status styles
	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Component styles
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2).
			Margin(0, 1)

	highlightStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	// Progress styles
	progressBarStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#374151")).
				Height(1)

	progressFillStyle = lipgloss.NewStyle().
				Background(primaryColor).
				Height(1)

	// Table styles
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(mutedColor)

	tableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Button styles
	buttonStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 3).
			Margin(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Foreground(primaryColor)

	buttonActiveStyle = buttonStyle.Copy().
				Background(primaryColor).
				Foreground(lipgloss.Color("#FFFFFF"))
)

// Adaptive styles that respond to terminal width
func adaptiveHeaderStyle(width int) lipgloss.Style {
	return headerStyle.Copy().Width(width - 4)
}

func adaptiveCardStyle(width int) lipgloss.Style {
	return cardStyle.Copy().Width(width - 6)
}

// Status indicators
func StatusIndicator(status string) string {
	indicators := map[string]string{
		"success": successStyle.Render("✓"),
		"warning": warningStyle.Render("⚠"),
		"error":   errorStyle.Render("✗"),
		"info":    infoStyle.Render("ℹ"),
		"loading": infoStyle.Render("⏳"),
		"pending": mutedStyle.Render("⋯"),
	}

	if indicator, ok := indicators[status]; ok {
		return indicator
	}
	return mutedStyle.Render("•")
}

// Progress bar component
func ProgressBar(progress float64, width int) string {
	if width < 10 {
		width = 10
	}

	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}

	bar := progressFillStyle.Render(strings.Repeat("█", filled))
	empty := progressBarStyle.Render(strings.Repeat("░", width-filled))

	return bar + empty
}
