package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kashifsb/nsm/pkg/logger"
	"github.com/kashifsb/nsm/pkg/utils"
)

func initializeConfig(cfg *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	cfg.HomeDir = homeDir
	cfg.ConfigDir = filepath.Join(homeDir, ".nsm")
	cfg.DataDir = filepath.Join(cfg.ConfigDir, "data")
	cfg.LogDir = filepath.Join(cfg.ConfigDir, "logs")
	cfg.BinDir = filepath.Join(homeDir, "bin")

	// Check system capabilities
	cfg.HasHomebrew = utils.IsCommandAvailable("brew")
	cfg.HasSystemd = utils.IsCommandAvailable("systemctl")
	cfg.HasSudo = utils.IsCommandAvailable("sudo")

	return nil
}

func createDirectories(ctx context.Context, cfg *Config) error {
	dirs := []string{
		cfg.ConfigDir,
		cfg.DataDir,
		cfg.LogDir,
		filepath.Join(cfg.DataDir, "certs"),
		filepath.Join(cfg.DataDir, "dns"),
	}

	for _, dir := range dirs {
		if err := utils.EnsureDir(dir); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
		logger.Debug("Created directory", "path", dir)
	}

	return nil
}

func installDependencies(ctx context.Context, cfg *Config) error {
	if cfg.SkipDeps {
		logger.Info("Skipping dependency installation")
		return nil
	}

	deps := getDependencies(cfg.Platform)

	for _, dep := range deps {
		if dep.Checker() {
			logger.Info("Dependency already installed", "name", dep.Name)
			continue
		}

		if !dep.Required {
			logger.Info("Optional dependency not found", "name", dep.Name)
			continue
		}

		logger.Info("Installing dependency", "name", dep.Name, "description", dep.Description)
		if err := dep.Installer(); err != nil {
			return fmt.Errorf("install %s: %w", dep.Name, err)
		}
	}

	return nil
}

func configureDNS(ctx context.Context, cfg *Config) error {
	logger.Info("Configuring DNS for development domains")

	switch cfg.Platform {
	case "darwin":
		return configureDNSMacOS(cfg)
	case "linux":
		return configureDNSLinux(cfg)
	default:
		logger.Warn("DNS auto-configuration not supported on this platform")
		return nil
	}
}

func setupTLDs(ctx context.Context, cfg *Config) error {
	logger.Info("Setting up TLDs", "tlds", cfg.TLDs)

	for _, tld := range cfg.TLDs {
		if err := addTLDConfiguration(tld); err != nil {
			return fmt.Errorf("setup TLD %s: %w", tld, err)
		}
		logger.Info("Configured TLD", "tld", tld)
	}

	return nil
}

func verifySetup(ctx context.Context, cfg *Config) error {
	logger.Info("Verifying installation")

	// Check dependencies
	deps := getDependencies(cfg.Platform)
	for _, dep := range deps {
		if dep.Required && !dep.Checker() {
			return fmt.Errorf("required dependency %s not found", dep.Name)
		}
	}

	// Test DNS resolution
	for _, tld := range cfg.TLDs {
		if err := testTLDResolution(tld); err != nil {
			logger.Warn("TLD resolution test failed", "tld", tld, "error", err)
		}
	}

	// Save setup completion
	if err := saveSetupCompletion(cfg); err != nil {
		logger.Warn("Failed to save setup completion", "error", err)
	}

	return nil
}

func getDependencies(platform string) []Dependency {
	deps := []Dependency{
		{
			Name:        "mkcert",
			Description: "Local certificate authority for HTTPS",
			Required:    true,
			Checker:     func() bool { return utils.IsCommandAvailable("mkcert") },
			Installer:   installMkcert,
		},
		{
			Name:        "dnsmasq",
			Description: "Lightweight DNS server for local development",
			Required:    true,
			Checker:     func() bool { return utils.IsCommandAvailable("dnsmasq") },
			Installer:   installDnsmasq,
		},
	}

	if platform == "darwin" {
		deps = append(deps, Dependency{
			Name:        "homebrew",
			Description: "Package manager for macOS",
			Required:    false,
			Checker:     func() bool { return utils.IsCommandAvailable("brew") },
			Installer:   installHomebrew,
		})
	}

	return deps
}

func installMkcert() error {
	if runtime.GOOS == "darwin" && utils.IsCommandAvailable("brew") {
		return runCommand("brew", "install", "mkcert")
	}

	// Platform-specific installation
	switch runtime.GOOS {
	case "linux":
		// Try different package managers
		if utils.IsCommandAvailable("apt") {
			return runCommand("sudo", "apt", "install", "-y", "mkcert")
		}
		if utils.IsCommandAvailable("yum") {
			return runCommand("sudo", "yum", "install", "-y", "mkcert")
		}
		if utils.IsCommandAvailable("pacman") {
			return runCommand("sudo", "pacman", "-S", "--noconfirm", "mkcert")
		}
		return fmt.Errorf("no supported package manager found")
	default:
		return fmt.Errorf("automatic installation not supported on %s", runtime.GOOS)
	}
}

func installDnsmasq() error {
	if runtime.GOOS == "darwin" && utils.IsCommandAvailable("brew") {
		if err := runCommand("brew", "install", "dnsmasq"); err != nil {
			return err
		}
		// Start dnsmasq service
		return runCommand("brew", "services", "start", "dnsmasq")
	}

	switch runtime.GOOS {
	case "linux":
		if utils.IsCommandAvailable("apt") {
			return runCommand("sudo", "apt", "install", "-y", "dnsmasq")
		}
		if utils.IsCommandAvailable("yum") {
			return runCommand("sudo", "yum", "install", "-y", "dnsmasq")
		}
		if utils.IsCommandAvailable("pacman") {
			return runCommand("sudo", "pacman", "-S", "--noconfirm", "dnsmasq")
		}
		return fmt.Errorf("no supported package manager found")
	default:
		return fmt.Errorf("automatic installation not supported on %s", runtime.GOOS)
	}
}

func installHomebrew() error {
	logger.Info("Installing Homebrew...")
	cmd := exec.Command("bash", "-c",
		`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)
	return cmd.Run()
}

func configureDNSMacOS(cfg *Config) error {
	// Configure dnsmasq
	dnsmasqConf := getDnsmasqConfig()
	confPath := "/opt/homebrew/etc/dnsmasq.conf"

	if err := os.WriteFile(confPath, []byte(dnsmasqConf), 0644); err != nil {
		return fmt.Errorf("write dnsmasq config: %w", err)
	}

	// Create hosts file
	hostsPath := "/opt/homebrew/etc/dnsmasq.hosts"
	hostsContent := "127.0.0.1 localhost\n"

	if err := os.WriteFile(hostsPath, []byte(hostsContent), 0644); err != nil {
		return fmt.Errorf("write dnsmasq hosts: %w", err)
	}

	// Restart dnsmasq
	return runCommand("brew", "services", "restart", "dnsmasq")
}

func configureDNSLinux(cfg *Config) error {
	// Similar to macOS but for Linux
	// Implementation would depend on the specific Linux distribution
	return fmt.Errorf("Linux DNS configuration not yet implemented")
}

func addTLDConfiguration(tld string) error {
	switch runtime.GOOS {
	case "darwin":
		return addTLDMacOS(tld)
	case "linux":
		return addTLDLinux(tld)
	default:
		return fmt.Errorf("TLD configuration not supported on %s", runtime.GOOS)
	}
}

func addTLDMacOS(tld string) error {
	// Create resolver file
	resolverDir := "/etc/resolver"
	resolverFile := filepath.Join(resolverDir, tld)

	resolverContent := "nameserver 127.0.0.1\nport 5353\n"

	if err := os.WriteFile(resolverFile, []byte(resolverContent), 0644); err != nil {
		// Try with sudo if permission denied
		cmd := exec.Command("sudo", "tee", resolverFile)
		cmd.Stdin = strings.NewReader(resolverContent)
		return cmd.Run()
	}

	return nil
}

func addTLDLinux(tld string) error {
	// Add to dnsmasq configuration
	// Implementation would depend on the specific setup
	return fmt.Errorf("Linux TLD configuration not yet implemented")
}

func removeTLDConfiguration(tld string) error {
	switch runtime.GOOS {
	case "darwin":
		resolverFile := filepath.Join("/etc/resolver", tld)
		return os.Remove(resolverFile)
	case "linux":
		// Remove from dnsmasq configuration
		return fmt.Errorf("Linux TLD removal not yet implemented")
	default:
		return nil
	}
}

func testTLDResolution(tld string) error {
	testDomain := fmt.Sprintf("test.%s", tld)
	cmd := exec.Command("nslookup", testDomain)
	return cmd.Run()
}

func isTLDConfigured(tld string) bool {
	switch runtime.GOOS {
	case "darwin":
		resolverFile := filepath.Join("/etc/resolver", tld)
		return utils.FileExists(resolverFile)
	default:
		return false
	}
}

func validateTLD(tld string) error {
	if tld == "" {
		return fmt.Errorf("TLD cannot be empty")
	}

	if strings.Contains(tld, ".") {
		return fmt.Errorf("TLD should not contain dots")
	}

	if len(tld) > 63 {
		return fmt.Errorf("TLD too long (max 63 characters)")
	}

	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getDnsmasqConfig() string {
	return `# NSM dnsmasq configuration
port=5353
listen-address=127.0.0.1
bind-interfaces

# Handle all local development TLDs
local=/dev/
local=/test/
local=/local/
local=/app/

# Additional hosts file
addn-hosts=/opt/homebrew/etc/dnsmasq.hosts

# Upstream DNS servers
server=1.1.1.1
server=1.0.0.1
server=8.8.8.8

# Cache settings
cache-size=1000
neg-ttl=60

# Don't read /etc/hosts
no-hosts

# Don't poll /etc/resolv.conf
no-poll

# Development domains
address=/dev/127.0.0.1
address=/test/127.0.0.1
address=/local/127.0.0.1
address=/app/127.0.0.1
`
}

func saveConfig(cfg Config) error {
	configFile := filepath.Join(cfg.ConfigDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	return os.WriteFile(configFile, data, 0644)
}

func loadConfig() (*Config, error) {
	homeDir, _ := os.UserHomeDir()
	configFile := filepath.Join(homeDir, ".nsm", "config.json")

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

func saveSetupCompletion(cfg *Config) error {
	setupInfo := map[string]interface{}{
		"completed_at": time.Now().Format(time.RFC3339),
		"version":      "3.0.0",
		"platform":     cfg.Platform,
		"tlds":         cfg.TLDs,
	}

	setupFile := filepath.Join(cfg.ConfigDir, "setup.json")
	data, err := json.MarshalIndent(setupInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal setup info: %w", err)
	}

	return os.WriteFile(setupFile, data, 0644)
}

func getLastSetupTime() (time.Time, error) {
	homeDir, _ := os.UserHomeDir()
	setupFile := filepath.Join(homeDir, ".nsm", "setup.json")

	data, err := os.ReadFile(setupFile)
	if err != nil {
		return time.Time{}, err
	}

	var setupInfo map[string]interface{}
	if err := json.Unmarshal(data, &setupInfo); err != nil {
		return time.Time{}, err
	}

	timeStr, ok := setupInfo["completed_at"].(string)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid setup time format")
	}

	return time.Parse(time.RFC3339, timeStr)
}
