package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/kashifsb/nsm/internal/config"
)

// Header component
func RenderHeader() string {
	logo := logoStyle.Render("🚀 NSM")
	title := headerStyle.Render("Enterprise Development Environment Manager")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		logo,
		title,
	)
}

// Configuration summary component
func RenderConfigSummary(cfg *config.Config) string {
	var rows []string

	// Project information
	rows = append(rows, fmt.Sprintf("%-15s %s",
		tableHeaderStyle.Render("Project:"),
		highlightStyle.Render(string(cfg.ProjectType))))

	rows = append(rows, fmt.Sprintf("%-15s %s",
		tableHeaderStyle.Render("Directory:"),
		mutedStyle.Render(cfg.ProjectName)))

	rows = append(rows, fmt.Sprintf("%-15s %s",
		tableHeaderStyle.Render("Command:"),
		infoStyle.Render(cfg.Command)))

	// Network configuration
	if cfg.Domain != "" {
		rows = append(rows, fmt.Sprintf("%-15s %s",
			tableHeaderStyle.Render("Domain:"),
			successStyle.Render(cfg.Domain)))
	}

	rows = append(rows, fmt.Sprintf("%-15s %s",
		tableHeaderStyle.Render("HTTP Port:"),
		renderPortInfo(cfg.HTTPPort)))

	rows = append(rows, fmt.Sprintf("%-15s %s",
		tableHeaderStyle.Render("HTTPS Port:"),
		renderPortInfo(cfg.HTTPSPort)))

	// Features
	rows = append(rows, fmt.Sprintf("%-15s %s",
		tableHeaderStyle.Render("Clean URLs:"),
		renderFeatureStatus(cfg.UsePort443)))

	rows = append(rows, fmt.Sprintf("%-15s %s",
		tableHeaderStyle.Render("HTTPS:"),
		renderFeatureStatus(cfg.EnableHTTPS)))

	rows = append(rows, fmt.Sprintf("%-15s %s",
		tableHeaderStyle.Render("DNS:"),
		renderFeatureStatus(cfg.EnableDNS)))

	content := strings.Join(rows, "\n")
	return cardStyle.Render(content)
}

// Status panel component
func RenderStatusPanel(steps []StatusStep) string {
	var rows []string

	for _, step := range steps {
		indicator := StatusIndicator(step.Status)
		text := step.Description

		if step.Status == "loading" {
			text = infoStyle.Render(text)
		} else if step.Status == "success" {
			text = mutedStyle.Render(text)
		}

		row := fmt.Sprintf("%s %s", indicator, text)
		if step.Details != "" {
			row += mutedStyle.Render(fmt.Sprintf(" (%s)", step.Details))
		}

		rows = append(rows, row)
	}

	content := strings.Join(rows, "\n")
	return cardStyle.Render("Setup Progress\n\n" + content)
}

// URL display component
func RenderURLs(cfg *config.Config, httpPort, httpsPort int) string {
	var urls []string

	// Primary URL
	protocol := "https"
	if !cfg.EnableHTTPS {
		protocol = "http"
	}

	domain := cfg.Domain
	if domain == "" {
		domain = "localhost"
	}

	var primaryURL string
	if cfg.UsePort443 && cfg.EnableHTTPS {
		primaryURL = fmt.Sprintf("%s://%s", protocol, domain)
	} else {
		port := httpsPort
		if !cfg.EnableHTTPS {
			port = httpPort
		}
		primaryURL = fmt.Sprintf("%s://%s:%d", protocol, domain, port)
	}

	urls = append(urls, fmt.Sprintf("🌐 %s %s",
		successStyle.Render("Primary:"),
		highlightStyle.Render(primaryURL)))

	// Local fallback
	if domain != "localhost" {
		localURL := fmt.Sprintf("https://localhost:%d", httpsPort)
		urls = append(urls, fmt.Sprintf("🏠 %s %s",
			infoStyle.Render("Local:"),
			mutedStyle.Render(localURL)))
	}

	// Development server
	devURL := fmt.Sprintf("http://127.0.0.1:%d", httpPort)
	urls = append(urls, fmt.Sprintf("⚙️  %s %s",
		mutedStyle.Render("Dev Server:"),
		mutedStyle.Render(devURL)))

	content := strings.Join(urls, "\n")
	return cardStyle.Render("🔗 Access URLs\n\n" + content)
}

// Logs component
func RenderLogs(logs []LogEntry, maxLines int) string {
	if len(logs) == 0 {
		return cardStyle.Render("📋 Logs\n\n" + mutedStyle.Render("No logs yet..."))
	}

	var lines []string
	start := 0
	if len(logs) > maxLines {
		start = len(logs) - maxLines
	}

	for i := start; i < len(logs); i++ {
		log := logs[i]
		timestamp := mutedStyle.Render(log.Timestamp.Format("15:04:05"))
		level := renderLogLevel(log.Level)
		message := log.Message

		line := fmt.Sprintf("%s %s %s", timestamp, level, message)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	return cardStyle.Render("📋 Logs\n\n" + content)
}

// Helper functions
func renderPortInfo(port int) string {
	if port == 0 {
		return mutedStyle.Render("auto")
	}
	return infoStyle.Render(fmt.Sprintf("%d", port))
}

func renderFeatureStatus(enabled bool) string {
	if enabled {
		return successStyle.Render("enabled")
	}
	return mutedStyle.Render("disabled")
}

func renderLogLevel(level string) string {
	styles := map[string]lipgloss.Style{
		"DEBUG": mutedStyle,
		"INFO":  infoStyle,
		"WARN":  warningStyle,
		"ERROR": errorStyle,
	}

	if style, ok := styles[level]; ok {
		return style.Render(fmt.Sprintf("[%s]", level))
	}
	return mutedStyle.Render(fmt.Sprintf("[%s]", level))
}

// Data structures
type StatusStep struct {
	Name        string
	Description string
	Status      string // success, warning, error, info, loading, pending
	Details     string
}

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
}
