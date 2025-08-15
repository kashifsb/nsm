package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kashifsb/nsm/pkg/logger"
	"github.com/kashifsb/nsm/pkg/utils"
)

func RunInteractive(ctx context.Context, cfg Config) error {
	model := NewSetupModel(cfg)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	go func() {
		<-ctx.Done()
		p.Send(ShutdownMsg{})
	}()

	_, err := p.Run()
	return err
}

func RunHeadless(ctx context.Context, cfg Config) error {
	logger.Info("Running headless setup")

	// Initialize configuration
	if err := initializeConfig(&cfg); err != nil {
		return fmt.Errorf("initialize config: %w", err)
	}

	// Run setup steps
	steps := []SetupStep{
		{Name: "directories", Fn: createDirectories},
		{Name: "dependencies", Fn: installDependencies},
		{Name: "dns", Fn: configureDNS},
		{Name: "tlds", Fn: setupTLDs},
		{Name: "verification", Fn: verifySetup},
	}

	for _, step := range steps {
		logger.Info("Executing step", "step", step.Name)
		if err := step.Fn(ctx, &cfg); err != nil {
			return fmt.Errorf("step %s failed: %w", step.Name, err)
		}
	}

	// Save configuration
	if err := saveConfig(cfg); err != nil {
		logger.Warn("Failed to save configuration", "error", err)
	}

	logger.Info("🎉 NSM setup completed successfully!")
	return nil
}

func AddTLD(ctx context.Context, tld string) error {
	logger.Info("Adding TLD", "tld", tld)

	// Validate TLD
	if err := validateTLD(tld); err != nil {
		return fmt.Errorf("invalid TLD: %w", err)
	}

	// Load current config
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Check if already exists
	for _, existingTLD := range cfg.TLDs {
		if existingTLD == tld {
			logger.Info("TLD already configured", "tld", tld)
			return nil
		}
	}

	// Add TLD
	if err := addTLDConfiguration(tld); err != nil {
		return fmt.Errorf("configure TLD: %w", err)
	}

	// Update config
	cfg.TLDs = append(cfg.TLDs, tld)
	if err := saveConfig(*cfg); err != nil {
		logger.Warn("Failed to save config", "error", err)
	}

	logger.Info("✅ TLD added successfully", "tld", tld)
	return nil
}

func RemoveTLD(ctx context.Context, tld string) error {
	logger.Info("Removing TLD", "tld", tld)

	// Load current config
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Remove TLD configuration
	if err := removeTLDConfiguration(tld); err != nil {
		return fmt.Errorf("remove TLD config: %w", err)
	}

	// Update config
	var newTLDs []string
	for _, existingTLD := range cfg.TLDs {
		if existingTLD != tld {
			newTLDs = append(newTLDs, existingTLD)
		}
	}
	cfg.TLDs = newTLDs

	if err := saveConfig(*cfg); err != nil {
		logger.Warn("Failed to save config", "error", err)
	}

	logger.Info("✅ TLD removed successfully", "tld", tld)
	return nil
}

func ListTLDs(ctx context.Context) error {
	status, err := getSystemStatus()
	if err != nil {
		return fmt.Errorf("get system status: %w", err)
	}

	fmt.Println("📋 Configured TLDs:")
	fmt.Println()

	if len(status.TLDs) == 0 {
		fmt.Println("  No TLDs configured")
		return nil
	}

	for _, tld := range status.TLDs {
		status := "❌"
		if tld.Configured {
			status = "✅"
		}
		fmt.Printf("  %s %s\n", status, tld.Name)
		if tld.ResolverFile != "" {
			fmt.Printf("    Resolver: %s\n", tld.ResolverFile)
		}
	}

	return nil
}

func ShowStatus(ctx context.Context) error {
	status, err := getSystemStatus()
	if err != nil {
		return fmt.Errorf("get system status: %w", err)
	}

	// Pretty print status
	fmt.Println("🔍 NSM Environment Status")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("Platform: %s\n", status.Platform)
	fmt.Printf("NSM Installed: %s\n", boolToStatus(status.NSMInstalled))
	fmt.Println()

	fmt.Println("Dependencies:")
	fmt.Printf("  mkcert:   %s\n", boolToStatus(status.Dependencies.Mkcert))
	fmt.Printf("  dnsmasq:  %s\n", boolToStatus(status.Dependencies.Dnsmasq))
	if status.Platform == "darwin" {
		fmt.Printf("  homebrew: %s\n", boolToStatus(status.Dependencies.Homebrew))
	}
	fmt.Println()

	fmt.Printf("Configuration Directory: %s\n", status.ConfigDir)
	if status.LastSetup != "" {
		fmt.Printf("Last Setup: %s\n", status.LastSetup)
	}

	return nil
}

func Reset(ctx context.Context) error {
	logger.Info("Resetting NSM configuration")

	// Get current config
	cfg, err := loadConfig()
	if err != nil {
		logger.Warn("Could not load config", "error", err)
		// Continue with reset anyway
	}

	// Remove TLD configurations
	if cfg != nil {
		for _, tld := range cfg.TLDs {
			if err := removeTLDConfiguration(tld); err != nil {
				logger.Warn("Failed to remove TLD", "tld", tld, "error", err)
			}
		}
	}

	// Remove configuration directories
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".nsm")

	if utils.DirExists(configDir) {
		if err := os.RemoveAll(configDir); err != nil {
			logger.Warn("Failed to remove config directory", "error", err)
		} else {
			logger.Info("Removed configuration directory", "dir", configDir)
		}
	}

	logger.Info("🧹 NSM configuration reset completed")
	return nil
}

func CreateExample(ctx context.Context, framework string) error {
	logger.Info("Creating example project", "framework", framework)

	exampleManager := NewExampleManager()
	return exampleManager.Create(framework)
}

// Helper functions
func boolToStatus(b bool) string {
	if b {
		return "✅ Available"
	}
	return "❌ Missing"
}

func getSystemStatus() (*SystemStatus, error) {
	status := &SystemStatus{
		Platform: runtime.GOOS,
	}

	// Check NSM installation
	status.NSMInstalled = utils.IsCommandAvailable("nsm")

	// Check dependencies
	status.Dependencies.Mkcert = utils.IsCommandAvailable("mkcert")
	status.Dependencies.Dnsmasq = utils.IsCommandAvailable("dnsmasq")

	if status.Platform == "darwin" {
		status.Dependencies.Homebrew = utils.IsCommandAvailable("brew")
	}

	// Get config directory
	homeDir, _ := os.UserHomeDir()
	status.ConfigDir = filepath.Join(homeDir, ".nsm")

	// Load TLDs
	if cfg, err := loadConfig(); err == nil {
		for _, tld := range cfg.TLDs {
			tldConfig := TLDConfig{
				Name:       tld,
				Configured: isTLDConfigured(tld),
			}
			status.TLDs = append(status.TLDs, tldConfig)
		}

		// Get last setup time
		if setupTime, err := getLastSetupTime(); err == nil {
			status.LastSetup = setupTime.Format("2006-01-02 15:04:05")
		}
	}

	return status, nil
}

type SetupStep struct {
	Name string
	Fn   func(context.Context, *Config) error
}
