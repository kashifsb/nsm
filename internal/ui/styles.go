package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Enhanced color palette for better UX
	primaryColor    = lipgloss.Color("#6366F1") // Indigo
	successColor    = lipgloss.Color("#10B981") // Emerald
	warningColor    = lipgloss.Color("#F59E0B") // Amber
	errorColor      = lipgloss.Color("#EF4444") // Red
	accentColor     = lipgloss.Color("#06B6D4") // Cyan
	mutedColor      = lipgloss.Color("#6B7280") // Gray
	backgroundColor = lipgloss.Color("#0F172A") // Slate 900
	cardBgColor     = lipgloss.Color("#1E293B") // Slate 800
	borderColor     = lipgloss.Color("#334155") // Slate 700

	// Base styles with improved spacing
	baseStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Margin(1, 0)

	// Enhanced header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			BorderBackground(cardBgColor).
			Background(cardBgColor).
			Padding(1, 3).
			Margin(1, 0)

	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 2).
			Margin(0, 1)

	// Enhanced status styles
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

	// Enhanced component styles
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Background(cardBgColor).
			Padding(1, 2).
			Margin(0, 1)

	highlightStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

	// Enhanced progress styles
	progressBarStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#374151")).
				Height(1)

	progressFillStyle = lipgloss.NewStyle().
				Background(primaryColor).
				Height(1)

	// Enhanced table styles
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(borderColor)

	tableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Enhanced button styles
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

	// New styles for better UX
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	urlStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Underline(true)

	statusCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Background(cardBgColor).
			Padding(1, 2).
			Margin(0, 1)
)

// Adaptive styles that respond to terminal width
func adaptiveHeaderStyle(width int) lipgloss.Style {
	return headerStyle.Copy().Width(width - 4)
}

func adaptiveCardStyle(width int) lipgloss.Style {
	return cardStyle.Copy().Width(width - 6)
}

// Enhanced status indicators with better visual feedback
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

// Enhanced progress bar component
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

// New utility functions for better UX
func RenderSection(title string, content string) string {
	titleSection := titleStyle.Render(title)
	return lipgloss.JoinVertical(lipgloss.Left, titleSection, content)
}

func RenderInfoBox(title string, content string) string {
	header := infoStyle.Render("ℹ " + title)
	body := cardStyle.Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func RenderSuccessBox(title string, content string) string {
	header := successStyle.Render("✓ " + title)
	body := cardStyle.Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func RenderErrorBox(title string, content string) string {
	header := errorStyle.Render("✗ " + title)
	body := cardStyle.Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}
