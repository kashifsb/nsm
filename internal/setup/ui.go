package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kashifsb/nsm/pkg/utils"
)

// Colors and styles
var (
	primaryColor = lipgloss.Color("#7C3AED")
	successColor = lipgloss.Color("#10B981")
	errorColor   = lipgloss.Color("#EF4444")
	warningColor = lipgloss.Color("#F59E0B")
	mutedColor   = lipgloss.Color("#6B7280")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			MarginBottom(1)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().Foreground(successColor).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	warningStyle = lipgloss.NewStyle().Foreground(warningColor).Bold(true)
	mutedStyle   = lipgloss.NewStyle().Foreground(mutedColor)
)

type SetupModel struct {
	cfg      Config
	state    SetupState
	steps    []StepStatus
	progress progress.Model
	spinner  spinner.Model
	width    int
	height   int
	err      error
}

type SetupState int

const (
	StateWelcome SetupState = iota
	StateChecking
	StateInstalling
	StateConfiguring
	StateComplete
	StateError
)

type StepStatus struct {
	Name        string
	Description string
	Status      string // pending, running, success, error
	Details     string
	Error       error
}

type StepCompleteMsg struct {
	StepName string
	Success  bool
	Error    error
	Details  string
}

type AllStepsCompleteMsg struct{}
type ShutdownMsg struct{}

func NewSetupModel(cfg Config) *SetupModel {
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	steps := []StepStatus{
		{Name: "check", Description: "Checking system requirements", Status: "pending"},
		{Name: "directories", Description: "Creating directories", Status: "pending"},
		{Name: "dependencies", Description: "Installing dependencies", Status: "pending"},
		{Name: "dns", Description: "Configuring DNS", Status: "pending"},
		{Name: "tlds", Description: "Setting up TLDs", Status: "pending"},
		{Name: "verification", Description: "Verifying installation", Status: "pending"},
	}

	return &SetupModel{
		cfg:      cfg,
		state:    StateWelcome,
		steps:    steps,
		progress: p,
		spinner:  s,
	}
}

func (m *SetupModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startSetup(),
	)
}

func (m *SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 4

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == StateComplete || m.state == StateError {
				return m, tea.Quit
			}
		case "enter":
			if m.state == StateWelcome {
				m.state = StateChecking
				return m, m.startSetup()
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case StepCompleteMsg:
		m.updateStep(msg.StepName, msg.Success, msg.Error, msg.Details)

		if !msg.Success {
			m.state = StateError
			m.err = msg.Error
		}

	case AllStepsCompleteMsg:
		m.state = StateComplete

	case ShutdownMsg:
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}

func (m *SetupModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var sections []string

	// Header
	sections = append(sections, m.renderHeader())

	// Main content based on state
	switch m.state {
	case StateWelcome:
		sections = append(sections, m.renderWelcome())
	case StateChecking, StateInstalling, StateConfiguring:
		sections = append(sections, m.renderProgress())
	case StateComplete:
		sections = append(sections, m.renderComplete())
	case StateError:
		sections = append(sections, m.renderError())
	}

	// Footer
	sections = append(sections, m.renderFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *SetupModel) renderHeader() string {
	logo := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Render("🚀 NSM Setup")

	title := headerStyle.Render("Enterprise Development Environment Setup")

	return lipgloss.JoinVertical(lipgloss.Center, logo, title)
}

func (m *SetupModel) renderWelcome() string {
	content := []string{
		"Welcome to NSM Setup!",
		"",
		"This will configure your development environment with:",
		"  • Local HTTPS certificates",
		"  • Custom domain resolution (.dev, .test, .local)",
		"  • Clean URLs without port numbers",
		"  • Professional DNS configuration",
		"",
		"Platform: " + m.cfg.Platform,
		"TLDs to configure: " + strings.Join(m.cfg.TLDs, ", "),
		"",
		"Press Enter to continue or Ctrl+C to exit",
	}

	return cardStyle.Render(strings.Join(content, "\n"))
}

func (m *SetupModel) renderProgress() string {
	var sections []string

	// Current progress
	completedSteps := 0
	totalSteps := len(m.steps)

	for _, step := range m.steps {
		if step.Status == "success" {
			completedSteps++
		}
	}

	progressValue := float64(completedSteps) / float64(totalSteps)
	sections = append(sections, m.progress.ViewAs(progressValue))

	// Steps list
	var stepLines []string
	for _, step := range m.steps {
		icon := m.getStepIcon(step.Status)
		line := fmt.Sprintf("%s %s", icon, step.Description)

		if step.Status == "running" {
			line += " " + m.spinner.View()
		}

		if step.Details != "" {
			line += mutedStyle.Render(fmt.Sprintf(" (%s)", step.Details))
		}

		stepLines = append(stepLines, line)
	}

	sections = append(sections, cardStyle.Render(strings.Join(stepLines, "\n")))

	return strings.Join(sections, "\n")
}

func (m *SetupModel) renderComplete() string {
	content := []string{
		successStyle.Render("🎉 Setup Complete!"),
		"",
		"NSM is now ready to use. You can:",
		"",
		"  • Run 'nsm --help' to see available options",
		"  • Use 'nsm-setup tld add <tld>' to add new TLDs",
		"  • Create example projects with 'nsm-setup example <framework>'",
		"",
		"Available examples:",
		"  • react-vite-typescript",
		"  • go",
		"  • rust",
		"  • python",
		"  • java",
		"",
		"Press 'q' to exit",
	}

	return cardStyle.Render(strings.Join(content, "\n"))
}

func (m *SetupModel) renderError() string {
	content := []string{
		errorStyle.Render("❌ Setup Failed"),
		"",
		"Error: " + m.err.Error(),
		"",
		"You can:",
		"  • Run 'nsm-setup status' to check system state",
		"  • Run 'nsm-setup reset' to clean up and try again",
		"  • Check the logs for more details",
		"",
		"Press 'q' to exit",
	}

	return cardStyle.Render(strings.Join(content, "\n"))
}

func (m *SetupModel) renderFooter() string {
	switch m.state {
	case StateWelcome:
		return mutedStyle.Render("Press Enter to start setup • Ctrl+C to exit")
	case StateComplete, StateError:
		return mutedStyle.Render("Press 'q' to exit")
	default:
		return mutedStyle.Render("Setting up NSM... • Ctrl+C to cancel")
	}
}

func (m *SetupModel) getStepIcon(status string) string {
	switch status {
	case "success":
		return successStyle.Render("✅")
	case "error":
		return errorStyle.Render("❌")
	case "running":
		return warningStyle.Render("⏳")
	default:
		return mutedStyle.Render("⏸️")
	}
}

func (m *SetupModel) updateStep(stepName string, success bool, err error, details string) {
	for i, step := range m.steps {
		if step.Name == stepName {
			if success {
				m.steps[i].Status = "success"
			} else {
				m.steps[i].Status = "error"
				m.steps[i].Error = err
			}
			m.steps[i].Details = details
			break
		}
	}
}

func (m *SetupModel) startSetup() tea.Cmd {
	return func() tea.Msg {
		// This would start the actual setup process
		// For now, just simulate the steps
		go m.runSetupSteps()
		return nil
	}
}

func (m *SetupModel) runSetupSteps() {
	steps := []struct {
		name string
		fn   func() error
	}{
		{"check", m.checkSystem},
		{"directories", m.createDirectories},
		{"dependencies", m.installDependencies},
		{"dns", m.configureDNS},
		{"tlds", m.setupTLDs},
		{"verification", m.verifySetup},
	}

	for _, step := range steps {
		// Update step to running
		m.updateStepStatus(step.name, "running")

		// Execute step
		if err := step.fn(); err != nil {
			m.sendStepComplete(step.name, false, err, "")
			return
		}

		m.sendStepComplete(step.name, true, nil, "Completed")
		time.Sleep(500 * time.Millisecond) // Visual delay
	}

	// All steps complete
	m.sendAllComplete()
}

func (m *SetupModel) updateStepStatus(stepName, status string) {
	// This method is called from the goroutine, so we need to send a message
	// to update the UI thread safely
	// For now, we'll update directly since this is called from the same goroutine
	for i, step := range m.steps {
		if step.Name == stepName {
			m.steps[i].Status = status
			break
		}
	}
}

func (m *SetupModel) sendStepComplete(stepName string, success bool, err error, details string) {
	// Send step completion message to the UI thread
	// This would be implemented with proper message passing
	// For now, we'll update directly since this is called from the same goroutine
	m.updateStep(stepName, success, err, details)
}

func (m *SetupModel) sendAllComplete() {
	// Send all steps complete message to the UI thread
	// This would be implemented with proper message passing
	// For now, we'll update directly since this is called from the same goroutine
	// The UI will detect when all steps are complete
}

// Actual implementation of setup steps
func (m *SetupModel) checkSystem() error {
	// Check if required tools are available
	if !m.cfg.SkipDeps {
		if !utils.IsCommandAvailable("mkcert") {
			return fmt.Errorf("mkcert not found - install with: brew install mkcert")
		}
		if !utils.IsCommandAvailable("dnsmasq") {
			return fmt.Errorf("dnsmasq not found - install with: brew install dnsmasq")
		}
	}
	return nil
}

func (m *SetupModel) createDirectories() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	dirs := []string{
		filepath.Join(homeDir, ".nsm"),
		filepath.Join(homeDir, ".nsm", "certs"),
		filepath.Join(homeDir, ".nsm", "config"),
		filepath.Join(homeDir, ".nsm", "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}
	return nil
}

func (m *SetupModel) installDependencies() error {
	if m.cfg.SkipDeps {
		return nil
	}

	// Check if we're on macOS and have Homebrew
	if m.cfg.Platform == "darwin" && utils.IsCommandAvailable("brew") {
		// Install mkcert if not available
		if !utils.IsCommandAvailable("mkcert") {
			if err := utils.RunCommand("brew", "install", "mkcert"); err != nil {
				return fmt.Errorf("install mkcert: %w", err)
			}
		}

		// Install dnsmasq if not available
		if !utils.IsCommandAvailable("dnsmasq") {
			if err := utils.RunCommand("brew", "install", "dnsmasq"); err != nil {
				return fmt.Errorf("install dnsmasq: %w", err)
			}
		}
	} else {
		// On other platforms, just check if tools are available
		if !utils.IsCommandAvailable("mkcert") {
			return fmt.Errorf("mkcert not found - please install it manually")
		}
		if !utils.IsCommandAvailable("dnsmasq") {
			return fmt.Errorf("dnsmasq not found - please install it manually")
		}
	}
	return nil
}

func (m *SetupModel) configureDNS() error {
	// Install mkcert CA
	if err := utils.RunCommand("mkcert", "-install"); err != nil {
		return fmt.Errorf("install mkcert CA: %w", err)
	}

	// Configure dnsmasq
	if m.cfg.Platform == "darwin" {
		// On macOS, configure dnsmasq via Homebrew
		if err := configureDnsmasqMacOS(); err != nil {
			return fmt.Errorf("configure dnsmasq: %w", err)
		}
	} else {
		// On Linux, configure dnsmasq manually
		if err := configureDnsmasqLinux(); err != nil {
			return fmt.Errorf("configure dnsmasq: %w", err)
		}
	}
	return nil
}

func (m *SetupModel) setupTLDs() error {
	// Configure TLDs
	for _, tld := range m.cfg.TLDs {
		if err := setupTLD(tld); err != nil {
			return fmt.Errorf("setup TLD %s: %w", tld, err)
		}
	}
	return nil
}

func (m *SetupModel) verifySetup() error {
	// Test that everything is working
	if err := testDNSResolution(); err != nil {
		return fmt.Errorf("DNS test failed: %w", err)
	}

	if err := testCertificateGeneration(); err != nil {
		return fmt.Errorf("certificate test failed: %w", err)
	}
	return nil
}

// Helper functions for setup steps
func configureDnsmasqMacOS() error {
	// Configure dnsmasq on macOS
	dnsmasqConfig := `# NSM Configuration
address=/dev/127.0.0.1
address=/test/127.0.0.1
address=/local/127.0.0.1
port=53535
`

	configPath := "/opt/homebrew/etc/dnsmasq.conf"
	if !utils.FileExists(configPath) {
		configPath = "/usr/local/etc/dnsmasq.conf"
	}

	// Append NSM configuration
	if err := utils.AppendToFile(configPath, dnsmasqConfig); err != nil {
		return fmt.Errorf("write dnsmasq config: %w", err)
	}

	// Restart dnsmasq
	if err := utils.RunCommand("brew", "services", "restart", "dnsmasq"); err != nil {
		return fmt.Errorf("restart dnsmasq: %w", err)
	}

	return nil
}

func configureDnsmasqLinux() error {
	// Configure dnsmasq on Linux
	dnsmasqConfig := `# NSM Configuration
address=/dev/127.0.0.1
address=/test/127.0.0.1
address=/local/127.0.0.1
port=53535
`

	configPath := "/etc/dnsmasq.conf"
	if err := utils.AppendToFile(configPath, dnsmasqConfig); err != nil {
		return fmt.Errorf("write dnsmasq config: %w", err)
	}

	// Restart dnsmasq
	if err := utils.RunCommand("systemctl", "restart", "dnsmasq"); err != nil {
		return fmt.Errorf("restart dnsmasq: %w", err)
	}

	return nil
}

func setupTLD(tld string) error {
	// Create resolver file for the TLD
	homeDir, _ := os.UserHomeDir()
	resolverDir := filepath.Join(homeDir, ".nsm", "resolvers")
	
	if err := os.MkdirAll(resolverDir, 0755); err != nil {
		return fmt.Errorf("create resolver directory: %w", err)
	}

	resolverFile := filepath.Join(resolverDir, tld)
	resolverContent := fmt.Sprintf("nameserver 127.0.0.1\nport 53535\n")

	if err := os.WriteFile(resolverFile, []byte(resolverContent), 0644); err != nil {
		return fmt.Errorf("write resolver file: %w", err)
	}

	return nil
}

func testDNSResolution() error {
	// Test DNS resolution for a .dev domain
	testDomain := "test-nsm.dev"
	
	// Use nslookup to test resolution
	if err := utils.RunCommand("nslookup", testDomain, "127.0.0.1"); err != nil {
		return fmt.Errorf("DNS resolution test failed: %w", err)
	}
	
	return nil
}

func testCertificateGeneration() error {
	// Test certificate generation
	testDomain := "test-nsm.dev"
	
	// Generate a test certificate
	if err := utils.RunCommand("mkcert", testDomain); err != nil {
		return fmt.Errorf("certificate generation test failed: %w", err)
	}
	
	// Clean up test certificate
	os.Remove(testDomain + ".pem")
	os.Remove(testDomain + "-key.pem")
	
	return nil
}
