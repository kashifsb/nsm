package project

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/shlex"

	"github.com/kashifsb/nsm/internal/config"
	"github.com/kashifsb/nsm/pkg/logger"
)

type Runner struct {
	cfg     *config.Config
	cmd     *exec.Cmd
	program *tea.Program
}

type RunnerConfig struct {
	WorkingDir string
	Command    string
	Env        map[string]string
}

type OutputMsg struct {
	Source string // "stdout" or "stderr"
	Line   string
}

type ProcessExitMsg struct {
	ExitCode int
	Error    error
}

func NewRunner(cfg *config.Config, program *tea.Program) *Runner {
	return &Runner{
		cfg:     cfg,
		program: program,
	}
}

func (r *Runner) Start(ctx context.Context, runnerCfg RunnerConfig) error {
	// Parse command
	args, err := r.parseCommand(runnerCfg.Command)
	if err != nil {
		return fmt.Errorf("parse command: %w", err)
	}

	// Create command
	r.cmd = exec.CommandContext(ctx, args[0], args[1:]...)
	r.cmd.Dir = runnerCfg.WorkingDir
	r.cmd.Env = r.buildEnvironment(runnerCfg.Env)

	// Set process group to handle cleanup properly
	r.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Setup pipes
	stdout, err := r.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}

	stderr, err := r.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("create stderr pipe: %w", err)
	}

	// Start command
	logger.Info("Starting development command",
		"command", runnerCfg.Command,
		"working_dir", runnerCfg.WorkingDir)

	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	// Start output streaming
	go r.streamOutput(stdout, "stdout")
	go r.streamOutput(stderr, "stderr")

	// Wait for completion
	go r.waitForCompletion()

	return nil
}

func (r *Runner) Stop() error {
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}

	logger.Info("Stopping development command")

	// Send SIGTERM to the process group
	pgid, err := syscall.Getpgid(r.cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		// Fallback to killing just the main process
		r.cmd.Process.Signal(os.Interrupt)
	}

	// Wait for graceful shutdown with timeout
	done := make(chan error, 1)
	go func() {
		done <- r.cmd.Wait()
	}()

	select {
	case <-done:
		logger.Info("Development command stopped gracefully")
		return nil
	case <-time.After(10 * time.Second):
		logger.Warn("Development command didn't stop gracefully, forcing kill")

		// Force kill the process group
		if pgid, err := syscall.Getpgid(r.cmd.Process.Pid); err == nil {
			syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			r.cmd.Process.Kill()
		}

		return fmt.Errorf("process killed after timeout")
	}
}

func (r *Runner) IsRunning() bool {
	return r.cmd != nil && r.cmd.Process != nil && r.cmd.ProcessState == nil
}

func (r *Runner) GetPID() int {
	if r.cmd != nil && r.cmd.Process != nil {
		return r.cmd.Process.Pid
	}
	return 0
}

func (r *Runner) parseCommand(command string) ([]string, error) {
	// First try shlex for proper shell parsing
	args, err := shlex.Split(command)
	if err != nil {
		// Fallback to simple space splitting
		logger.Debug("Failed to parse command with shlex, using simple split", "error", err)
		return strings.Fields(command), nil
	}
	return args, nil
}

func (r *Runner) buildEnvironment(extraEnv map[string]string) []string {
	env := os.Environ()

	// Add extra environment variables
	for key, value := range extraEnv {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// Add NSM-specific environment variables
	nsmEnv := map[string]string{
		"NSM_ENABLED":      "true",
		"NSM_VERSION":      "3.0.0",
		"NSM_PROJECT_TYPE": string(r.cfg.ProjectType),
		"NSM_PROJECT_NAME": r.cfg.ProjectName,
		"NSM_DOMAIN":       r.cfg.Domain,
		"NSM_DATA_DIR":     r.cfg.DataDir,
	}

	if r.cfg.EnableHTTPS {
		nsmEnv["NSM_HTTPS_ENABLED"] = "true"
		nsmEnv["NSM_CERT_PATH"] = r.cfg.CertPath
		nsmEnv["NSM_KEY_PATH"] = r.cfg.KeyPath
	}

	if r.cfg.UsePort443 {
		nsmEnv["NSM_CLEAN_URLS"] = "true"
	}

	for key, value := range nsmEnv {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env
}

func (r *Runner) streamOutput(reader io.Reader, source string) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // Handle large lines

	for scanner.Scan() {
		line := scanner.Text()

		// Filter and enhance output
		line = r.processOutputLine(line, source)

		// Send to UI
		if r.program != nil {
			r.program.Send(OutputMsg{
				Source: source,
				Line:   line,
			})
		}

		// Also log for debugging
		if source == "stderr" {
			logger.Debug("Dev command stderr", "line", line)
		} else {
			logger.Debug("Dev command stdout", "line", line)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Error reading command output", "source", source, "error", err)
	}
}

func (r *Runner) processOutputLine(line, source string) string {
	// Remove ANSI color codes if needed (optional)
	// line = stripANSI(line)

	// Add timestamp for important messages
	if r.isImportantLine(line) {
		timestamp := time.Now().Format("15:04:05")
		return fmt.Sprintf("[%s] %s", timestamp, line)
	}

	return line
}

func (r *Runner) isImportantLine(line string) bool {
	importantPatterns := []string{
		"error",
		"Error",
		"ERROR",
		"warning",
		"Warning",
		"WARN",
		"Local:",
		"Network:",
		"ready in",
		"compiled",
		"running at",
		"listening on",
		"server started",
	}

	lineLower := strings.ToLower(line)
	for _, pattern := range importantPatterns {
		if strings.Contains(lineLower, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

func (r *Runner) waitForCompletion() {
	err := r.cmd.Wait()

	exitCode := 0
	if r.cmd.ProcessState != nil {
		exitCode = r.cmd.ProcessState.ExitCode()
	}

	// Send completion message to UI
	if r.program != nil {
		r.program.Send(ProcessExitMsg{
			ExitCode: exitCode,
			Error:    err,
		})
	}

	if err != nil {
		logger.Error("Development command exited with error",
			"exit_code", exitCode,
			"error", err)
	} else {
		logger.Info("Development command completed", "exit_code", exitCode)
	}
}

// Helper function to strip ANSI escape codes (optional)
func stripANSI(str string) string {
	// This is a simple implementation - you might want to use a proper library
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZ-z]))"
	// For simplicity, returning as-is. Implement proper ANSI stripping if needed.
	return str
}
