package dns

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kashifsb/nsm/pkg/logger"
)

type Resolver struct {
	domain     string
	tld        string
	configured bool
}

type ResolverConfig struct {
	Domain    string
	EnableDNS bool
}

func NewResolver(cfg ResolverConfig) *Resolver {
	domain := cfg.Domain
	if domain == "" {
		domain = "localhost"
	}

	var tld string
	if parts := strings.Split(domain, "."); len(parts) > 1 {
		tld = parts[len(parts)-1]
	}

	return &Resolver{
		domain: domain,
		tld:    tld,
	}
}

func (r *Resolver) Setup() error {
	if r.domain == "localhost" || r.domain == "" {
		logger.Debug("Skipping DNS setup for localhost")
		return nil
	}

	logger.Info("Setting up DNS resolution", "domain", r.domain)

	switch runtime.GOOS {
	case "darwin":
		return r.setupMacOS()
	case "linux":
		return r.setupLinux()
	default:
		logger.Warn("DNS auto-configuration not supported on this platform")
		return r.setupManual()
	}
}

func (r *Resolver) Cleanup() error {
	if !r.configured {
		return nil
	}

	logger.Info("Cleaning up DNS configuration", "domain", r.domain)

	switch runtime.GOOS {
	case "darwin":
		return r.cleanupMacOS()
	case "linux":
		return r.cleanupLinux()
	default:
		return nil
	}
}

func (r *Resolver) Test() error {
	if r.domain == "localhost" {
		return nil
	}

	logger.Debug("Testing DNS resolution", "domain", r.domain)

	// Test with timeout
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * 2,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addrs, err := resolver.LookupHost(ctx, r.domain)
	if err != nil {
		return fmt.Errorf("DNS lookup failed: %w", err)
	}

	// Check if resolves to localhost
	for _, addr := range addrs {
		if addr == "127.0.0.1" || addr == "::1" {
			logger.Debug("DNS resolution working", "domain", r.domain, "ip", addr)
			return nil
		}
	}

	return fmt.Errorf("domain resolves to %v instead of localhost", addrs)
}

// macOS implementation
func (r *Resolver) setupMacOS() error {
	// First ensure dnsmasq is available and configured
	if err := r.ensureDnsmasq(); err != nil {
		return fmt.Errorf("dnsmasq setup failed: %w", err)
	}

	// Create resolver file for the TLD
	if err := r.createMacOSResolver(); err != nil {
		logger.Warn("Failed to create resolver file", "error", err)
		// Continue without resolver file - dnsmasq might still work
	}

	// Add domain to dnsmasq
	if err := r.addDnsmasqEntry(); err != nil {
		return fmt.Errorf("failed to add dnsmasq entry: %w", err)
	}

	// Restart dnsmasq
	if err := r.restartDnsmasq(); err != nil {
		logger.Warn("Failed to restart dnsmasq", "error", err)
	}

	r.configured = true
	return nil
}

func (r *Resolver) setupLinux() error {
	// Try systemd-resolved first
	if r.hasSystemdResolved() {
		return r.setupSystemdResolved()
	}

	// Fall back to dnsmasq
	return r.ensureDnsmasq()
}

func (r *Resolver) setupManual() error {
	logger.Info("Manual DNS setup required")
	logger.Info("Add this line to your /etc/hosts file:")
	logger.Info(fmt.Sprintf("127.0.0.1 %s", r.domain))
	return nil
}

func (r *Resolver) ensureDnsmasq() error {
	// Check if dnsmasq is installed
	if _, err := exec.LookPath("dnsmasq"); err != nil {
		return fmt.Errorf("dnsmasq not installed: %w", err)
	}

	// Get dnsmasq configuration path
	configPath := r.getDnsmasqConfigPath()
	hostsPath := r.getDnsmasqHostsPath()

	// Ensure configuration exists
	if err := r.ensureDnsmasqConfig(configPath, hostsPath); err != nil {
		return fmt.Errorf("dnsmasq config: %w", err)
	}

	// Ensure hosts file exists
	if err := r.ensureDnsmasqHosts(hostsPath); err != nil {
		return fmt.Errorf("dnsmasq hosts: %w", err)
	}

	return nil
}

func (r *Resolver) createMacOSResolver() error {
	resolverDir := "/etc/resolver"
	resolverFile := filepath.Join(resolverDir, r.tld)

	// Check if resolver directory exists
	if _, err := os.Stat(resolverDir); os.IsNotExist(err) {
		logger.Debug("Resolver directory doesn't exist, skipping resolver file creation")
		return nil
	}

	// Check if resolver file already exists and is correct
	if content, err := os.ReadFile(resolverFile); err == nil {
		if strings.Contains(string(content), "127.0.0.1") && strings.Contains(string(content), "5353") {
			logger.Debug("Resolver file already configured", "file", resolverFile)
			return nil
		}
	}

	// Try to create resolver file
	resolverContent := "nameserver 127.0.0.1\nport 5353\n"
	if err := os.WriteFile(resolverFile, []byte(resolverContent), 0o644); err != nil {
		return fmt.Errorf("create resolver file: %w", err)
	}

	logger.Info("Created DNS resolver file", "file", resolverFile)
	return nil
}

func (r *Resolver) addDnsmasqEntry() error {
	hostsPath := r.getDnsmasqHostsPath()
	entry := fmt.Sprintf("127.0.0.1 %s", r.domain)

	// Check if entry already exists
	if r.entryExists(hostsPath, entry) {
		logger.Debug("DNS entry already exists", "entry", entry)
		return nil
	}

	// Add entry
	file, err := os.OpenFile(hostsPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("open hosts file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(entry + "\n"); err != nil {
		return fmt.Errorf("write hosts entry: %w", err)
	}

	logger.Debug("Added DNS entry", "entry", entry)
	return nil
}

func (r *Resolver) restartDnsmasq() error {
	// Try brew services first (macOS)
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("brew", "services", "restart", "dnsmasq")
		if err := cmd.Run(); err == nil {
			logger.Debug("Restarted dnsmasq via brew services")
			return nil
		}
	}

	// Try systemctl (Linux)
	cmd := exec.Command("sudo", "systemctl", "restart", "dnsmasq")
	if err := cmd.Run(); err == nil {
		logger.Debug("Restarted dnsmasq via systemctl")
		return nil
	}

	// Try service command
	cmd = exec.Command("sudo", "service", "dnsmasq", "restart")
	if err := cmd.Run(); err == nil {
		logger.Debug("Restarted dnsmasq via service")
		return nil
	}

	return fmt.Errorf("failed to restart dnsmasq")
}

func (r *Resolver) getDnsmasqConfigPath() string {
	if runtime.GOOS == "darwin" {
		if brewPrefix := r.getBrewPrefix(); brewPrefix != "" {
			return filepath.Join(brewPrefix, "etc", "dnsmasq.conf")
		}
	}
	return "/etc/dnsmasq.conf"
}

func (r *Resolver) getDnsmasqHostsPath() string {
	if runtime.GOOS == "darwin" {
		if brewPrefix := r.getBrewPrefix(); brewPrefix != "" {
			return filepath.Join(brewPrefix, "etc", "dnsmasq.hosts")
		}
	}
	return "/etc/dnsmasq.hosts"
}

func (r *Resolver) getBrewPrefix() string {
	cmd := exec.Command("brew", "--prefix")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (r *Resolver) ensureDnsmasqConfig(configPath, hostsPath string) error {
	config := fmt.Sprintf(`# NSM dnsmasq configuration
port=5353
listen-address=127.0.0.1
bind-interfaces

# Handle local development domains
local=/dev/
local=/test/
local=/local/
local=/app/

# Additional hosts file
addn-hosts=%s

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
`, hostsPath)

	// Check if config file exists
	if _, err := os.Stat(configPath); err == nil {
		// Config exists, check if it contains our configuration
		content, err := os.ReadFile(configPath)
		if err == nil && strings.Contains(string(content), "port=5353") {
			logger.Debug("dnsmasq config already contains NSM configuration")
			return nil
		}
	}

	// Write configuration
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		return fmt.Errorf("write dnsmasq config: %w", err)
	}

	logger.Info("Created dnsmasq configuration", "file", configPath)
	return nil
}

func (r *Resolver) ensureDnsmasqHosts(hostsPath string) error {
	if _, err := os.Stat(hostsPath); err == nil {
		return nil // File already exists
	}

	initialContent := `# NSM dnsmasq hosts file
# Development domains will be added here automatically
127.0.0.1 localhost
`

	if err := os.WriteFile(hostsPath, []byte(initialContent), 0o644); err != nil {
		return fmt.Errorf("create dnsmasq hosts file: %w", err)
	}

	logger.Info("Created dnsmasq hosts file", "file", hostsPath)
	return nil
}

func (r *Resolver) entryExists(hostsPath, entry string) bool {
	file, err := os.Open(hostsPath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == entry {
			return true
		}
	}

	return false
}

func (r *Resolver) hasSystemdResolved() bool {
	_, err := exec.LookPath("systemd-resolve")
	return err == nil
}

func (r *Resolver) setupSystemdResolved() error {
	// This would implement systemd-resolved configuration
	// For now, fall back to dnsmasq
	return r.ensureDnsmasq()
}

// Cleanup methods
func (r *Resolver) cleanupMacOS() error {
	hostsPath := r.getDnsmasqHostsPath()
	entry := fmt.Sprintf("127.0.0.1 %s", r.domain)

	return r.removeEntryFromFile(hostsPath, entry)
}

func (r *Resolver) cleanupLinux() error {
	// Similar cleanup for Linux
	return r.cleanupMacOS()
}

func (r *Resolver) removeEntryFromFile(filePath, entry string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != entry {
			lines = append(lines, line)
		}
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}
