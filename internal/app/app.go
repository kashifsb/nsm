package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kashifsb/nsm/internal/cert"
	"github.com/kashifsb/nsm/internal/config"
	"github.com/kashifsb/nsm/internal/dns"
	"github.com/kashifsb/nsm/internal/platform"
	"github.com/kashifsb/nsm/internal/project"
	"github.com/kashifsb/nsm/internal/server"
	"github.com/kashifsb/nsm/internal/ui"
	"github.com/kashifsb/nsm/pkg/logger"
	"github.com/kashifsb/nsm/pkg/utils"
)

type App struct {
	cfg *config.Config

	// Managers
	portManager *platform.PortManager
	certManager *cert.Manager
	dnsResolver *dns.Resolver
	proxyServer *server.ProxyServer
	runner      *project.Runner

	// UI
	program *tea.Program

	// State
	httpPort  int
	httpsPort int
	running   bool
}

type SetupStep struct {
	Name        string
	Description string
	Execute     func(ctx context.Context) error
}

func NewApp(cfg *config.Config) (*App, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize managers
	portManager := platform.NewPortManager()

	certManager, err := cert.NewManager(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("create cert manager: %w", err)
	}

	dnsResolver := dns.NewResolver(dns.ResolverConfig{
		Domain:    cfg.Domain,
		EnableDNS: cfg.EnableDNS,
	})

	return &App{
		cfg:         cfg,
		portManager: portManager,
		certManager: certManager,
		dnsResolver: dnsResolver,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	logger.Info("Starting NSM application", "project", a.cfg.ProjectName)

	// Setup UI
	model := ui.NewModel(a.cfg)
	a.program = tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Set program reference in model for message passing
	model.SetProgram(a.program)

	// Initialize runner with UI program
	a.runner = project.NewRunner(a.cfg, a.program)

	// Start UI in background
	uiDone := make(chan error, 1)
	go func() {
		_, err := a.program.Run()
		uiDone <- err
	}()

	// Run setup steps
	if err := a.runSetup(ctx); err != nil {
		a.program.Send(ui.ErrorMsg{Err: err})
		return fmt.Errorf("setup failed: %w", err)
	}

	// Start services
	if err := a.startServices(ctx); err != nil {
		a.program.Send(ui.ErrorMsg{Err: err})
		return fmt.Errorf("failed to start services: %w", err)
	}

	a.running = true
	a.program.Send(ui.ServerStartedMsg{
		HTTPPort:  a.httpPort,
		HTTPSPort: a.httpsPort,
	})

	// Wait for shutdown
	select {
	case <-ctx.Done():
		logger.Info("Received shutdown signal")
	case err := <-uiDone:
		if err != nil {
			logger.Error("UI error", "error", err)
		}
	}

	// Cleanup
	return a.shutdown()
}

func (a *App) RunHeadless(ctx context.Context) error {
	logger.Info("Starting NSM application in headless mode", "project", a.cfg.ProjectName)

	// Run setup steps
	if err := a.runSetup(ctx); err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// Start services
	if err := a.startServices(ctx); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	a.running = true
	logger.Info("NSM running in headless mode",
		"project", a.cfg.ProjectName,
		"http_port", a.httpPort,
		"https_port", a.httpsPort)

	// Wait for shutdown signal
	<-ctx.Done()

	// Cleanup
	return a.shutdown()
}

func (a *App) runSetup(ctx context.Context) error {
	steps := []SetupStep{
		{
			Name:        "validate",
			Description: "Validating configuration",
			Execute:     a.setupValidation,
		},
		{
			Name:        "ports",
			Description: "Configuring ports",
			Execute:     a.setupPorts,
		},
		{
			Name:        "certs",
			Description: "Setting up certificates",
			Execute:     a.setupCertificates,
		},
		{
			Name:        "dns",
			Description: "Configuring DNS",
			Execute:     a.setupDNS,
		},
	}

	for _, step := range steps {
		logger.Info("Executing setup step", "step", step.Name)

		// Send UI update if program is available
		if a.program != nil {
			a.program.Send(ui.StepUpdateMsg{
				StepName: step.Name,
				Status:   "loading",
				Details:  "In progress...",
			})
		}

		// Execute step with timeout
		stepCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		stepDone := make(chan error, 1)

		go func(step SetupStep) {
			stepDone <- step.Execute(stepCtx)
		}(step)

		select {
		case err := <-stepDone:
			cancel()
			if err != nil {
				// Send UI error if program is available
				if a.program != nil {
					a.program.Send(ui.StepUpdateMsg{
						StepName: step.Name,
						Status:   "error",
						Details:  err.Error(),
					})
				}
				return fmt.Errorf("step %s failed: %w", step.Name, err)
			}
		case <-stepCtx.Done():
			cancel()
			err := fmt.Errorf("step %s timed out after 30 seconds", step.Name)
			// Send UI error if program is available
			if a.program != nil {
				a.program.Send(ui.StepUpdateMsg{
					StepName: step.Name,
					Status:   "error",
					Details:  err.Error(),
				})
			}
			return err
		}

		// Send UI success if program is available
		if a.program != nil {
			a.program.Send(ui.StepUpdateMsg{
				StepName: step.Name,
				Status:   "success",
				Details:  "Completed",
			})
		}
	}

	// Send UI completion if program is available
	if a.program != nil {
		a.program.Send(ui.SetupCompleteMsg{})
	}

	return nil
}

func (a *App) setupValidation(ctx context.Context) error {
	// Check required tools
	if !utils.IsCommandAvailable("mkcert") {
		return fmt.Errorf("mkcert not found - install with: brew install mkcert")
	}

	// Validate project directory
	if !utils.DirExists(a.cfg.ProjectDir) {
		return fmt.Errorf("project directory does not exist: %s", a.cfg.ProjectDir)
	}

	// Create data directories
	dirs := []string{
		filepath.Join(a.cfg.DataDir, "certs"),
		filepath.Join(a.cfg.DataDir, "logs"),
		filepath.Join(a.cfg.DataDir, "config"),
	}

	for _, dir := range dirs {
		if err := utils.EnsureDir(dir); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	return nil
}

func (a *App) setupPorts(ctx context.Context) error {
	var err error

	// Setup HTTP port
	if a.cfg.HTTPPort == 0 {
		a.httpPort, err = a.portManager.FindFreePort()
		if err != nil {
			return fmt.Errorf("find free HTTP port: %w", err)
		}
	} else {
		if !a.portManager.IsPortAvailable(a.cfg.HTTPPort) {
			return fmt.Errorf("HTTP port %d is not available", a.cfg.HTTPPort)
		}
		a.httpPort = a.cfg.HTTPPort
	}

	// Setup HTTPS port
	if a.cfg.HTTPSPort == 0 {
		if a.cfg.UsePort443 && a.portManager.CanUsePort443() {
			a.httpsPort = 443
			logger.Info("Using port 443 for clean URLs")
		} else {
			a.httpsPort, err = a.portManager.FindFreePortNear(8443)
			if err != nil {
				return fmt.Errorf("find free HTTPS port: %w", err)
			}
			a.cfg.UsePort443 = false
		}
	} else {
		if !a.portManager.IsPortAvailable(a.cfg.HTTPSPort) {
			return fmt.Errorf("HTTPS port %d is not available", a.cfg.HTTPSPort)
		}
		a.httpsPort = a.cfg.HTTPSPort
		a.cfg.UsePort443 = (a.httpsPort == 443)
	}

	logger.Info("Ports configured",
		"http", a.httpPort,
		"https", a.httpsPort,
		"clean_urls", a.cfg.UsePort443)

	return nil
}

func (a *App) setupCertificates(ctx context.Context) error {
	if !a.cfg.EnableHTTPS {
		logger.Info("HTTPS disabled, skipping certificate setup")
		return nil
	}

	domain := a.cfg.Domain
	if domain == "" {
		domain = "localhost"
	}

	certInfo, err := a.certManager.EnsureCertificate(domain, false)
	if err != nil {
		return fmt.Errorf("certificate setup: %w", err)
	}

	a.cfg.CertPath = certInfo.CertPath
	a.cfg.KeyPath = certInfo.KeyPath

	if certInfo.Created {
		logger.Info("Created new certificate", "domain", domain)
	} else {
		logger.Info("Using existing certificate", "domain", domain)
	}

	return nil
}

func (a *App) setupDNS(ctx context.Context) error {
	if !a.cfg.EnableDNS || a.cfg.Domain == "" || a.cfg.Domain == "localhost" {
		logger.Info("DNS setup skipped")
		return nil
	}

	if err := a.dnsResolver.Setup(); err != nil {
		logger.Warn("DNS setup failed, continuing without custom DNS", "error", err)
		// Don't fail the entire setup for DNS issues
		return nil
	}

	// Test DNS resolution
	if err := a.dnsResolver.Test(); err != nil {
		logger.Warn("DNS test failed", "error", err)
		// Don't fail for DNS test issues
	} else {
		logger.Info("DNS resolution configured successfully", "domain", a.cfg.Domain)
	}

	return nil
}

func (a *App) startServices(ctx context.Context) error {
	// Start proxy server
	if a.cfg.EnableProxy {
		a.program.Send(ui.StepUpdateMsg{
			StepName: "proxy",
			Status:   "loading",
			Details:  "Starting HTTPS proxy",
		})

		proxyConfig := server.ProxyConfig{
			TargetHost:  "127.0.0.1",
			TargetPort:  a.httpPort,
			ProxyPort:   a.httpsPort,
			Domain:      a.cfg.Domain,
			CertPath:    a.cfg.CertPath,
			KeyPath:     a.cfg.KeyPath,
			EnableHTTPS: a.cfg.EnableHTTPS,
		}

		a.proxyServer = server.NewProxyServer(a.cfg, proxyConfig)
		if err := a.proxyServer.Start(ctx, a.httpsPort); err != nil {
			return fmt.Errorf("start proxy server: %w", err)
		}

		a.program.Send(ui.StepUpdateMsg{
			StepName: "proxy",
			Status:   "success",
			Details:  fmt.Sprintf("Running on port %d", a.httpsPort),
		})
	}

	// Start development server
	if a.program != nil {
		a.program.Send(ui.StepUpdateMsg{
			StepName: "dev",
			Status:   "loading",
			Details:  "Starting development server",
		})
	}

	runnerConfig := project.RunnerConfig{
		WorkingDir: a.cfg.ProjectDir,
		Command:    a.cfg.Command,
		Env: map[string]string{
			"NSM_HTTP_PORT":  fmt.Sprintf("%d", a.httpPort),
			"NSM_HTTPS_PORT": fmt.Sprintf("%d", a.httpsPort),
			"PORT":           fmt.Sprintf("%d", a.httpPort),
			"HOST":           "127.0.0.1",
		},
	}

	if err := a.runner.Start(ctx, runnerConfig); err != nil {
		if a.program != nil {
			a.program.Send(ui.StepUpdateMsg{
				StepName: "dev",
				Status:   "error",
				Details:  fmt.Sprintf("Failed to start: %v", err),
			})
		}
		return fmt.Errorf("start development server: %w", err)
	}

	// Wait for development server to be ready with better error handling
	logger.Info("Waiting for development server to be ready", "port", a.httpPort)

	serverReady := make(chan bool, 1)
	go func() {
		if err := a.portManager.WaitForPort(a.httpPort, 30*time.Second); err != nil {
			logger.Warn("Development server may not be ready", "error", err)
			serverReady <- false
		} else {
			serverReady <- true
		}
	}()

	// Wait for server readiness or timeout
	select {
	case ready := <-serverReady:
		if ready {
			logger.Info("Development server is ready", "port", a.httpPort)
		} else {
			logger.Warn("Development server may not be ready, continuing anyway")
		}
	case <-time.After(35 * time.Second):
		logger.Warn("Timeout waiting for development server, continuing anyway")
	}

	if a.program != nil {
		a.program.Send(ui.StepUpdateMsg{
			StepName: "dev",
			Status:   "success",
			Details:  fmt.Sprintf("Running on port %d", a.httpPort),
		})
	}

	return nil
}

func (a *App) shutdown() error {
	logger.Info("Shutting down NSM application")

	var errs []error

	// Stop development server with better error handling
	if a.runner != nil {
		logger.Info("Stopping development server")
		if err := a.runner.Stop(); err != nil {
			logger.Warn("Failed to stop development server gracefully", "error", err)
			errs = append(errs, fmt.Errorf("stop development server: %w", err))
		} else {
			logger.Info("Development server stopped successfully")
		}
	}

	// Stop proxy server with better error handling
	if a.proxyServer != nil {
		logger.Info("Stopping proxy server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.proxyServer.Stop(ctx); err != nil {
			logger.Warn("Failed to stop proxy server gracefully", "error", err)
			errs = append(errs, fmt.Errorf("stop proxy server: %w", err))
		} else {
			logger.Info("Proxy server stopped successfully")
		}
	}

	// Cleanup DNS with better error handling
	if a.dnsResolver != nil {
		logger.Info("Cleaning up DNS configuration")
		if err := a.dnsResolver.Cleanup(); err != nil {
			logger.Warn("Failed to cleanup DNS configuration", "error", err)
			errs = append(errs, fmt.Errorf("cleanup DNS: %w", err))
		} else {
			logger.Info("DNS configuration cleaned up successfully")
		}
	}

	// Release ports with better error handling
	if a.portManager != nil {
		logger.Info("Releasing ports")
		if a.httpPort > 0 {
			a.portManager.ReleasePort(a.httpPort)
			logger.Debug("Released HTTP port", "port", a.httpPort)
		}
		if a.httpsPort > 0 {
			a.portManager.ReleasePort(a.httpsPort)
			logger.Debug("Released HTTPS port", "port", a.httpsPort)
		}
	}

	// Clean up port info file
	portInfoPath := filepath.Join(a.cfg.ProjectDir, ".nsm-ports.json")
	if utils.FileExists(portInfoPath) {
		if err := os.Remove(portInfoPath); err != nil {
			logger.Warn("Failed to remove port info file", "path", portInfoPath, "error", err)
		} else {
			logger.Debug("Removed port info file", "path", portInfoPath)
		}
	}

	// Log shutdown results
	if len(errs) > 0 {
		logger.Warn("Shutdown completed with some errors", "error_count", len(errs))
		for i, err := range errs {
			logger.Warn("Shutdown error", "index", i, "error", err.Error())
		}
		return fmt.Errorf("shutdown completed with %d errors: %v", len(errs), errs)
	}

	logger.Info("Shutdown completed successfully")
	return nil
}

func (a *App) IsRunning() bool {
	return a.running
}

func (a *App) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"running":       a.running,
		"project_name":  a.cfg.ProjectName,
		"project_type":  string(a.cfg.ProjectType),
		"domain":        a.cfg.Domain,
		"http_port":     a.httpPort,
		"https_port":    a.httpsPort,
		"clean_urls":    a.cfg.UsePort443,
		"https_enabled": a.cfg.EnableHTTPS,
		"dns_enabled":   a.cfg.EnableDNS,
	}

	if a.runner != nil {
		status["dev_server_pid"] = a.runner.GetPID()
		status["dev_server_running"] = a.runner.IsRunning()
	}

	return status
}
