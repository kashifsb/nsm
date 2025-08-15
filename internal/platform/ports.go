package platform

import (
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/kashifsb/nsm/pkg/logger"
	"github.com/shirou/gopsutil/v3/process"
)

type PortManager struct {
	allocatedPorts map[int]bool
}

type PortInfo struct {
	Port        int
	Available   bool
	ProcessName string
	PID         int32
}

func NewPortManager() *PortManager {
	return &PortManager{
		allocatedPorts: make(map[int]bool),
	}
}

func (pm *PortManager) FindFreePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("find free port: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	pm.allocatedPorts[port] = true

	logger.Debug("Found free port", "port", port)
	return port, nil
}

func (pm *PortManager) FindFreePortNear(preferred int) (int, error) {
	// Try preferred port first
	if pm.IsPortAvailable(preferred) {
		pm.allocatedPorts[preferred] = true
		logger.Debug("Using preferred port", "port", preferred)
		return preferred, nil
	}

	// Try nearby ports
	for offset := 1; offset <= 100; offset++ {
		for _, port := range []int{preferred + offset, preferred - offset} {
			if port > 1024 && port < 65535 && pm.IsPortAvailable(port) {
				pm.allocatedPorts[port] = true
				logger.Debug("Found nearby port", "preferred", preferred, "actual", port)
				return port, nil
			}
		}
	}

	// Fall back to any free port
	return pm.FindFreePort()
}

func (pm *PortManager) IsPortAvailable(port int) bool {
	if pm.allocatedPorts[port] {
		return false
	}

	// Try to bind to the port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	listener.Close()

	return true
}

func (pm *PortManager) CanUsePort443() bool {
	// Check if we can bind to port 443
	listener, err := net.Listen("tcp", ":443")
	if err != nil {
		logger.Debug("Cannot bind to port 443", "error", err)
		return false
	}
	listener.Close()

	logger.Debug("Port 443 is available")
	return true
}

func (pm *PortManager) GetPortInfo(port int) PortInfo {
	info := PortInfo{
		Port:      port,
		Available: pm.IsPortAvailable(port),
	}

	if !info.Available {
		// Try to find which process is using the port
		if processInfo := pm.getProcessUsingPort(port); processInfo != nil {
			info.ProcessName = processInfo.Name
			info.PID = processInfo.PID
		}
	}

	return info
}

func (pm *PortManager) WaitForPort(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
		if err == nil {
			conn.Close()
			logger.Debug("Port became available", "port", port)
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("port %d did not become available within %v", port, timeout)
}

func (pm *PortManager) ReleasePort(port int) {
	delete(pm.allocatedPorts, port)
	logger.Debug("Released port", "port", port)
}

func (pm *PortManager) ValidatePortRange(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}

	if port < 1024 && runtime.GOOS != "windows" {
		// On Unix systems, ports below 1024 typically require root privileges
		return fmt.Errorf("port %d requires root privileges on %s", port, runtime.GOOS)
	}

	return nil
}

type processInfo struct {
	Name string
	PID  int32
}

func (pm *PortManager) getProcessUsingPort(port int) *processInfo {
	processes, err := process.Processes()
	if err != nil {
		logger.Debug("Failed to get process list", "error", err)
		return nil
	}

	for _, p := range processes {
		connections, err := p.Connections()
		if err != nil {
			continue
		}

		for _, conn := range connections {
			if conn.Laddr.Port == uint32(port) {
				name, err := p.Name()
				if err != nil {
					name = "unknown"
				}

				return &processInfo{
					Name: name,
					PID:  p.Pid,
				}
			}
		}
	}

	return nil
}

// Platform-specific optimizations
func (pm *PortManager) GetSystemPortPreferences() map[string]int {
	preferences := map[string]int{
		"http":  5173, // Vite default
		"https": 8443, // Common development HTTPS port
	}

	// Platform-specific adjustments
	switch runtime.GOOS {
	case "darwin":
		// macOS specific preferences
		preferences["https"] = 443 // Prefer 443 on macOS for clean URLs
	case "linux":
		// Linux specific preferences
		// Keep defaults
	case "windows":
		// Windows specific preferences
		preferences["http"] = 3000 // Common on Windows
		preferences["https"] = 8443
	}

	return preferences
}

func (pm *PortManager) GetAvailablePortsInRange(start, end int, count int) ([]int, error) {
	if start < 1 || end > 65535 || start >= end {
		return nil, fmt.Errorf("invalid port range: %d-%d", start, end)
	}

	var ports []int

	for port := start; port <= end && len(ports) < count; port++ {
		if pm.IsPortAvailable(port) {
			ports = append(ports, port)
		}
	}

	if len(ports) == 0 {
		return nil, fmt.Errorf("no available ports in range %d-%d", start, end)
	}

	return ports, nil
}
