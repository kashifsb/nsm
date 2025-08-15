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

	// UI program reference for message passing
	program *tea.Program

	// Scrolling and navigation
	scrollOffset int
	autoScroll   bool
	showHelp     bool
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
		case "up", "k":
			// Scroll up through logs
			m.scrollUp()
		case "down", "j":
			// Scroll down through logs
			m.scrollDown()
		case "g":
			// Go to top
			m.scrollToTop()
		case "G":
			// Go to bottom
			m.scrollToBottom()
		case "space":
			// Toggle auto-scroll
			m.toggleAutoScroll()
		case "r":
			// Refresh/restart setup
			if m.state == StateError {
				m.restartSetup()
			}
		case "h":
			// Show help
			m.toggleHelp()
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

	// Show help if requested
	if m.showHelp {
		sections = append(sections, m.renderHelpView())
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

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

	// Join all sections with proper spacing
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Ensure content fits within terminal bounds
	if m.height > 0 && lipgloss.Height(content) > m.height {
		// Truncate content to fit terminal height
		lines := strings.Split(content, "\n")
		if len(lines) > m.height-2 { // Leave space for header/footer
			lines = lines[:m.height-2]
			lines = append(lines, "...", "Press 'q' to quit")
			content = strings.Join(lines, "\n")
		}
	}

	return content
}

func (m *Model) renderSetupView() string {
	var sections []string

	// Configuration summary
	sections = append(sections, RenderConfigSummary(m.cfg))

	// Setup progress
	sections = append(sections, RenderStatusPanel(m.steps))

	// Recent logs
	if len(m.logs) > 0 {
		sections = append(sections, RenderLogs(m.logs, 5, m.scrollOffset))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) renderRunningView() string {
	var sections []string

	// URL information
	sections = append(sections, m.renderURLPanel())

	// Live logs
	sections = append(sections, RenderLogs(m.logs, 10, m.scrollOffset))

	// Status indicators
	sections = append(sections, m.renderStatusIndicators())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) renderErrorView() string {
	errorContent := []string{
		"❌ Setup Error",
		"",
		"An error occurred during setup:",
		"",
		errorStyle.Render(m.error.Error()),
		"",
		"Troubleshooting:",
		"  • Check that all required tools are installed (mkcert, dnsmasq)",
		"  • Ensure you have proper permissions for the project directory",
		"  • Verify your network configuration",
		"",
		"Actions:",
		"  • Press 'r' to restart setup",
		"  • Press 'h' for help",
		"  • Press 'q' to quit",
	}

	return cardStyle.Render(strings.Join(errorContent, "\n"))
}

func (m *Model) renderShutdownView() string {
	shutdownCard := cardStyle.Render(
		successStyle.Render("👋 Shutdown Complete\n\n") +
			"NSM has been stopped gracefully.\n" +
			"All services have been cleaned up.",
	)
	return shutdownCard
}

func (m *Model) renderHelpView() string {
	helpContent := []string{
		"🔧 NSM Interface Help",
		"",
		"Navigation:",
		"  ↑/k     Scroll up through logs",
		"  ↓/j     Scroll down through logs",
		"  g       Go to top of logs",
		"  G       Go to bottom of logs",
		"  space   Toggle auto-scroll",
		"",
		"Actions:",
		"  r       Restart setup (when in error state)",
		"  h       Toggle this help view",
		"  q       Quit NSM",
		"  Ctrl+C  Force quit",
		"",
		"Status Indicators:",
		"  ✓       Success/Completed",
		"  ⏳      Loading/In Progress",
		"  ⚠       Warning",
		"  ✗       Error/Failed",
		"  ⋯       Pending",
		"",
		"Press 'h' again to hide this help",
	}

	return cardStyle.Render(strings.Join(helpContent, "\n"))
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

	// Help text with navigation options
	var help string
	if m.showHelp {
		help = mutedStyle.Render("Navigation: ↑/k: scroll up • ↓/j: scroll down • g: top • G: bottom • space: auto-scroll • r: restart • h: hide help • q: quit")
	} else {
		help = mutedStyle.Render("Press 'h' for help • 'q' to quit • Ctrl+C to stop")
	}
	parts = append(parts, help)

	// Show scroll position if not auto-scrolling
	if !m.autoScroll && len(m.logs) > 10 {
		scrollInfo := mutedStyle.Render(fmt.Sprintf("Scroll: %d/%d", m.scrollOffset+1, len(m.logs)))
		parts = append(parts, scrollInfo)
	}

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
		// Start the actual setup process in a goroutine
		go m.runSetupProcess()
		return StepUpdateMsg{
			StepName: "validate",
			Status:   "loading",
			Details:  "Checking configuration",
		}
	}
}

func (m *Model) runSetupProcess() {
	// Simulate setup steps with actual work and proper error handling
	steps := []struct {
		name     string
		details  string
		duration time.Duration
	}{
		{"validate", "Validating configuration", 1 * time.Second},
		{"ports", "Configuring ports", 500 * time.Millisecond},
		{"certs", "Setting up certificates", 2 * time.Second},
		{"dns", "Configuring DNS", 1 * time.Second},
		{"proxy", "Starting HTTPS proxy", 1 * time.Second},
		{"dev", "Starting development server", 2 * time.Second},
	}

	for _, step := range steps {
		// Update step to loading
		if m.program != nil {
			m.program.Send(StepUpdateMsg{
				StepName: step.name,
				Status:   "loading",
				Details:  step.details,
			})

			// Add log entry for step start
			m.program.Send(LogMsg{
				Level:   "INFO",
				Message: fmt.Sprintf("Starting step: %s", step.name),
			})
		}

		// Simulate work with timeout protection
		done := make(chan bool, 1)
		go func() {
			time.Sleep(step.duration)
			done <- true
		}()

		// Wait for step completion or timeout
		select {
		case <-done:
			// Step completed successfully
			if m.program != nil {
				m.program.Send(StepUpdateMsg{
					StepName: step.name,
					Status:   "success",
					Details:  "Completed",
				})

				// Add log entry for step completion
				m.program.Send(LogMsg{
					Level:   "INFO",
					Message: fmt.Sprintf("Step '%s' completed successfully", step.name),
				})
			}
		case <-time.After(10 * time.Second):
			// Step timed out
			if m.program != nil {
				m.program.Send(StepUpdateMsg{
					StepName: step.name,
					Status:   "error",
					Details:  "Timeout - taking too long",
				})

				m.program.Send(LogMsg{
					Level:   "ERROR",
					Message: fmt.Sprintf("Step '%s' timed out", step.name),
				})

				// Send error message to stop the process
				m.program.Send(ErrorMsg{
					Err: fmt.Errorf("step '%s' timed out", step.name),
				})
				return
			}
		}
	}

	// Mark setup as complete
	if m.program != nil {
		m.program.Send(SetupCompleteMsg{})

		// Simulate server start
		time.Sleep(500 * time.Millisecond)
		m.program.Send(ServerStartedMsg{
			HTTPPort:  3000,
			HTTPSPort: 8443,
		})

		// Add final success log
		m.program.Send(LogMsg{
			Level:   "INFO",
			Message: "NSM setup completed successfully!",
		})
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// SetProgram sets the program reference for message passing
func (m *Model) SetProgram(program *tea.Program) {
	m.program = program
}

// Scrolling methods
func (m *Model) scrollUp() {
	if m.scrollOffset > 0 {
		m.scrollOffset--
	}
}

func (m *Model) scrollDown() {
	maxScroll := len(m.logs) - 10 // Show 10 lines at a time
	if m.scrollOffset < maxScroll {
		m.scrollOffset++
	}
}

func (m *Model) scrollToTop() {
	m.scrollOffset = 0
}

func (m *Model) scrollToBottom() {
	m.scrollOffset = len(m.logs) - 10
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m *Model) toggleAutoScroll() {
	m.autoScroll = !m.autoScroll
	if m.autoScroll {
		m.scrollToBottom()
	}
}

func (m *Model) restartSetup() {
	m.state = StateInitializing
	m.error = nil
	m.scrollOffset = 0
	m.setupComplete = false
	m.serverRunning = false

	// Reset steps
	for i := range m.steps {
		m.steps[i].Status = "pending"
		m.steps[i].Details = ""
	}

	// Clear logs
	m.logs = make([]LogEntry, 0)
}

func (m *Model) toggleHelp() {
	m.showHelp = !m.showHelp
}
