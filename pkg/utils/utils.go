package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// File system utilities
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func HomeDir() (string, error) {
	return os.UserHomeDir()
}

func TempDir() string {
	return os.TempDir()
}

func WorkingDir() (string, error) {
	return os.Getwd()
}

// String utilities
func SanitizeFilename(name string) string {
	// Replace problematic characters
	replacements := map[string]string{
		" ":  "-",
		"/":  "-",
		"\\": "-",
		":":  "-",
		"*":  "-",
		"?":  "-",
		"\"": "-",
		"<":  "-",
		">":  "-",
		"|":  "-",
	}

	result := name
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// Remove consecutive dashes
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	// Trim dashes from start and end
	result = strings.Trim(result, "-")

	// Ensure it's not empty
	if result == "" {
		result = "unnamed"
	}

	return strings.ToLower(result)
}

func GenerateID(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	if maxLength <= 3 {
		return s[:maxLength]
	}

	return s[:maxLength-3] + "..."
}

// System utilities
func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func GetOSInfo() map[string]string {
	return map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	}
}

func GetBrewPrefix() (string, error) {
	if runtime.GOOS != "darwin" {
		return "", fmt.Errorf("brew is only available on macOS")
	}

	cmd := exec.Command("brew", "--prefix")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get brew prefix: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Network utilities
func IsValidPort(port int) bool {
	return port > 0 && port <= 65535
}

func IsPrivilegedPort(port int) bool {
	return port < 1024
}

func ParsePort(s string) (int, error) {
	port, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid port: %w", err)
	}

	if !IsValidPort(port) {
		return 0, fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}

	return port, nil
}

// Process utilities
func IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix systems, sending signal 0 checks if process exists
	err = process.Signal(os.Signal(syscall.Signal(0)))
	return err == nil
}

func KillProcess(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid: %d", pid)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	return process.Kill()
}

// JSON utilities
func PrettyJSON(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Time utilities
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}

	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}

	return fmt.Sprintf("%.1fh", d.Hours())
}

func HumanizeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	}

	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	days := int(diff.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	if days < 7 {
		return fmt.Sprintf("%d days ago", days)
	}

	return t.Format("2006-01-02")
}

// Configuration utilities
type Config struct {
	data map[string]interface{}
}

func NewConfig() *Config {
	return &Config{
		data: make(map[string]interface{}),
	}
}

func (c *Config) Set(key string, value interface{}) {
	c.data[key] = value
}

func (c *Config) Get(key string) interface{} {
	return c.data[key]
}

func (c *Config) GetString(key string) string {
	if v, ok := c.data[key].(string); ok {
		return v
	}
	return ""
}

func (c *Config) GetInt(key string) int {
	if v, ok := c.data[key].(int); ok {
		return v
	}
	return 0
}

func (c *Config) GetBool(key string) bool {
	if v, ok := c.data[key].(bool); ok {
		return v
	}
	return false
}

// Error utilities
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

func IsNotFound(err error) bool {
	return os.IsNotExist(err)
}

// Retry utilities
func Retry(attempts int, delay time.Duration, fn func() error) error {
	var lastErr error

	for i := 0; i < attempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if i < attempts-1 {
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", attempts, lastErr)
}

func RetryWithBackoff(attempts int, initialDelay time.Duration, maxDelay time.Duration, fn func() error) error {
	var lastErr error
	delay := initialDelay

	for i := 0; i < attempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if i < attempts-1 {
			time.Sleep(delay)
			delay *= 2
			if delay > maxDelay {
				delay = maxDelay
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", attempts, lastErr)
}
