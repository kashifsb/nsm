package setup

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	// Send message to update UI
	// This would be implemented with proper message passing
}

func (m *SetupModel) sendStepComplete(stepName string, success bool, err error, details string) {
	// Send step completion message
	// This would be implemented with proper message passing
}

func (m *SetupModel) sendAllComplete() {
	// Send all steps complete message
	// This would be implemented with proper message passing
}

// Placeholder methods for setup steps
func (m *SetupModel) checkSystem() error {
	time.Sleep(1 * time.Second)
	return nil
}

func (m *SetupModel) createDirectories() error {
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (m *SetupModel) installDependencies() error {
	time.Sleep(2 * time.Second)
	return nil
}

func (m *SetupModel) configureDNS() error {
	time.Sleep(1 * time.Second)
	return nil
}

func (m *SetupModel) setupTLDs() error {
	time.Sleep(1 * time.Second)
	return nil
}

func (m *SetupModel) verifySetup() error {
	time.Sleep(500 * time.Millisecond)
	return nil
}
