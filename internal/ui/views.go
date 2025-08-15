package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kashifsb/nsm/internal/config"
)

type Model struct {
	cfg    *config.Config
	state  AppState
	width  int
	height int

	// Components
	spinner spinner.Model
	logs    []LogEntry
	steps   []StatusStep
	urls    URLInfo

	// State
	setupComplete bool
	serverRunning bool
	error         error
}

type AppState int

const (
	StateInitializing AppState = iota
	StateSetup
	StateRunning
	StateShutdown
	StateError
)

type URLInfo struct {
	Primary string
	Local   string
	DevURL  string
}

// Messages
type (
	SetupCompleteMsg struct{}
	ServerStartedMsg struct {
		HTTPPort  int
		HTTPSPort int
	}
	LogMsg struct {
		Level   string
		Message string
	}
	StepUpdateMsg struct {
		StepName string
		Status   string
		Details  string
	}
	ShutdownMsg struct{}
	ErrorMsg    struct {
		Err error
	}
	TickMsg time.Time
)

func NewModel(cfg *config.Config) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	return &Model{
		cfg:     cfg,
		state:   StateInitializing,
		spinner: s,
		logs:    make([]LogEntry, 0),
		steps: []StatusStep{
			{Name: "validate", Description: "Validating configuration", Status: "pending"},
			{Name: "ports", Description: "Configuring ports", Status: "pending"},
			{Name: "certs", Description: "Setting up certificates", Status: "pending"},
			{Name: "dns", Description: "Configuring DNS", Status: "pending"},
			{Name: "proxy", Description: "Starting HTTPS proxy", Status: "pending"},
			{Name: "dev", Description: "Starting development server", Status: "pending"},
		},
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startSetup(),
		tickCmd(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case TickMsg:
		cmds = append(cmds, tickCmd())

	case StepUpdateMsg:
		m.updateStep(msg.StepName, msg.Status, msg.Details)

	case LogMsg:
		m.addLog(msg.Level, msg.Message)

	case ServerStartedMsg:
		m.state = StateRunning
		m.serverRunning = true
		m.updateURLs(msg.HTTPPort, msg.HTTPSPort)

	case SetupCompleteMsg:
		m.setupComplete = true

	case ErrorMsg:
		m.state = StateError
		m.error = msg.Err
		m.addLog("ERROR", msg.Err.Error())

	case ShutdownMsg:
		m.state = StateShutdown
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var sections []string

	// Header
	sections = append(sections, RenderHeader())

	// Main content based on state
	switch m.state {
	case StateInitializing, StateSetup:
		sections = append(sections, m.renderSetupView())
	case StateRunning:
		sections = append(sections, m.renderRunningView())
	case StateError:
		sections = append(sections, m.renderErrorView())
	case StateShutdown:
		sections = append(sections, m.renderShutdownView())
	}

	// Footer
	sections = append(sections, m.renderFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) renderSetupView() string {
	var sections []string

	// Configuration summary
	sections = append(sections, RenderConfigSummary(m.cfg))

	// Setup progress
	sections = append(sections, RenderStatusPanel(m.steps))

	// Recent logs
	if len(m.logs) > 0 {
		sections = append(sections, RenderLogs(m.logs, 5))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) renderRunningView() string {
	var sections []string

	// URL information
	sections = append(sections, m.renderURLPanel())

	// Live logs
	sections = append(sections, RenderLogs(m.logs, 10))

	// Status indicators
	sections = append(sections, m.renderStatusIndicators())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) renderErrorView() string {
	errorCard := cardStyle.Render(
		errorStyle.Render("❌ Error\n\n") +
			m.error.Error() + "\n\n" +
			mutedStyle.Render("Press 'q' to quit"),
	)
	return errorCard
}

func (m *Model) renderShutdownView() string {
	shutdownCard := cardStyle.Render(
		successStyle.Render("👋 Shutdown Complete\n\n") +
			"NSM has been stopped gracefully.\n" +
			"All services have been cleaned up.",
	)
	return shutdownCard
}

func (m *Model) renderURLPanel() string {
	if m.urls.Primary == "" {
		return ""
	}

	var urls []string

	urls = append(urls, fmt.Sprintf("🌐 %s %s",
		successStyle.Render("Primary:"),
		highlightStyle.Render(m.urls.Primary)))

	if m.urls.Local != "" {
		urls = append(urls, fmt.Sprintf("🏠 %s %s",
			infoStyle.Render("Local:"),
			mutedStyle.Render(m.urls.Local)))
	}

	urls = append(urls, fmt.Sprintf("⚙️  %s %s",
		mutedStyle.Render("Dev Server:"),
		mutedStyle.Render(m.urls.DevURL)))

	content := strings.Join(urls, "\n")
	return cardStyle.Render("🔗 Access URLs\n\n" + content)
}

func (m *Model) renderStatusIndicators() string {
	var indicators []string

	// Server status
	if m.serverRunning {
		indicators = append(indicators, fmt.Sprintf("%s Server: %s",
			StatusIndicator("success"),
			successStyle.Render("Running")))
	} else {
		indicators = append(indicators, fmt.Sprintf("%s Server: %s",
			StatusIndicator("error"),
			errorStyle.Render("Stopped")))
	}

	// Feature status
	if m.cfg.EnableHTTPS {
		indicators = append(indicators, fmt.Sprintf("%s HTTPS: %s",
			StatusIndicator("success"),
			successStyle.Render("Active")))
	}

	if m.cfg.EnableDNS && m.cfg.Domain != "" {
		indicators = append(indicators, fmt.Sprintf("%s DNS: %s",
			StatusIndicator("success"),
			successStyle.Render("Configured")))
	}

	if m.cfg.UsePort443 {
		indicators = append(indicators, fmt.Sprintf("%s Clean URLs: %s",
			StatusIndicator("success"),
			successStyle.Render("Enabled")))
	}

	content := strings.Join(indicators, "\n")
	return cardStyle.Render("📊 Status\n\n" + content)
}

func (m *Model) renderFooter() string {
	var parts []string

	// Show spinner during setup
	if m.state == StateSetup {
		parts = append(parts, m.spinner.View())
	}

	// Help text
	help := mutedStyle.Render("Press 'q' to quit • Ctrl+C to stop")
	parts = append(parts, help)

	return strings.Join(parts, " ")
}

// Helper methods
func (m *Model) updateStep(name, status, details string) {
	for i, step := range m.steps {
		if step.Name == name {
			m.steps[i].Status = status
			m.steps[i].Details = details
			break
		}
	}
}

func (m *Model) addLog(level, message string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	}

	m.logs = append(m.logs, entry)

	// Keep only last 100 logs
	if len(m.logs) > 100 {
		m.logs = m.logs[len(m.logs)-100:]
	}
}

func (m *Model) updateURLs(httpPort, httpsPort int) {
	domain := m.cfg.Domain
	if domain == "" {
		domain = "localhost"
	}

	// Primary URL
	if m.cfg.UsePort443 && m.cfg.EnableHTTPS {
		m.urls.Primary = fmt.Sprintf("https://%s", domain)
	} else if m.cfg.EnableHTTPS {
		m.urls.Primary = fmt.Sprintf("https://%s:%d", domain, httpsPort)
	} else {
		m.urls.Primary = fmt.Sprintf("http://%s:%d", domain, httpPort)
	}

	// Local fallback
	if domain != "localhost" {
		m.urls.Local = fmt.Sprintf("https://localhost:%d", httpsPort)
	}

	// Dev server URL
	m.urls.DevURL = fmt.Sprintf("http://127.0.0.1:%d", httpPort)
}

func (m *Model) startSetup() tea.Cmd {
	return func() tea.Msg {
		// This would trigger the actual setup process
		return StepUpdateMsg{
			StepName: "validate",
			Status:   "loading",
			Details:  "Checking configuration",
		}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
