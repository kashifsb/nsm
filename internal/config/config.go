package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type ProjectType string

const (
	ProjectTypeVite   ProjectType = "vite"
	ProjectTypeReact  ProjectType = "react"
	ProjectTypeGo     ProjectType = "go"
	ProjectTypeRust   ProjectType = "rust"
	ProjectTypePython ProjectType = "python"
	ProjectTypeJava   ProjectType = "java"
	ProjectTypeDotNet ProjectType = "dotnet"
	ProjectTypeNode   ProjectType = "node"
	ProjectTypeNext   ProjectType = "next"
)

type Config struct {
	// Project configuration
	ProjectType ProjectType `json:"project_type"`
	ProjectDir  string      `json:"project_dir"`
	ProjectName string      `json:"project_name"`
	Command     string      `json:"command"`

	// Network configuration
	Domain     string `json:"domain"`
	HTTPPort   int    `json:"http_port"`
	HTTPSPort  int    `json:"https_port"`
	UsePort443 bool   `json:"use_port_443"`

	// Feature flags
	EnableHTTPS bool `json:"enable_https"`
	EnableDNS   bool `json:"enable_dns"`
	EnableProxy bool `json:"enable_proxy"`

	// Runtime options
	Debug    bool `json:"debug"`
	AutoYes  bool `json:"auto_yes"`
	Headless bool `json:"headless"`

	// Paths
	DataDir  string `json:"data_dir"`
	CertPath string `json:"cert_path"`
	KeyPath  string `json:"key_path"`
}

func ParseFromFlags(cmd *cobra.Command) (*Config, error) {
	cfg := &Config{
		EnableHTTPS: true,
		EnableDNS:   true,
		EnableProxy: true,
		UsePort443:  true,
	}

	// Parse flags
	var err error

	if projectType, _ := cmd.Flags().GetString("project-type"); projectType != "" {
		cfg.ProjectType = ProjectType(projectType)
	}

	cfg.Domain, _ = cmd.Flags().GetString("domain")
	cfg.Command, _ = cmd.Flags().GetString("command")
	cfg.HTTPPort, _ = cmd.Flags().GetInt("http-port")
	cfg.HTTPSPort, _ = cmd.Flags().GetInt("https-port")
	cfg.Debug, _ = cmd.Flags().GetBool("debug")
	cfg.AutoYes, _ = cmd.Flags().GetBool("auto-yes")
	cfg.Headless, _ = cmd.Flags().GetBool("headless")

	if no443, _ := cmd.Flags().GetBool("no-443"); no443 {
		cfg.UsePort443 = false
	}

	// Set project directory and name
	cfg.ProjectDir, err = os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	cfg.ProjectName = filepath.Base(cfg.ProjectDir)
	cfg.ProjectName = strings.ToLower(strings.ReplaceAll(cfg.ProjectName, " ", "-"))

	// Setup data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}

	cfg.DataDir = filepath.Join(homeDir, ".nsm", cfg.ProjectName)

	// Auto-detect project type if not specified
	if cfg.ProjectType == "" {
		cfg.ProjectType = detectProjectType(cfg.ProjectDir)
	}

	// Set default command if not specified
	if cfg.Command == "" {
		cfg.Command = getDefaultCommand(cfg.ProjectType)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.ProjectType == "" {
		return fmt.Errorf("project type is required")
	}

	if c.Command == "" {
		return fmt.Errorf("development command is required")
	}

	return nil
}
