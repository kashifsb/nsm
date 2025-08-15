package setup

import (
	"runtime"
)

type Config struct {
	// Runtime options
	Headless bool
	AutoYes  bool
	SkipDeps bool
	Platform string

	// DNS configuration
	TLDs []string

	// Paths
	HomeDir   string
	ConfigDir string
	DataDir   string
	LogDir    string
	BinDir    string

	// System info
	HasHomebrew bool
	HasSystemd  bool
	HasSudo     bool
}

type Dependency struct {
	Name        string
	Description string
	Required    bool
	Installer   func() error
	Checker     func() bool
}

type TLDConfig struct {
	Name         string
	Configured   bool
	ResolverFile string
	DnsmasqEntry string
}

type SystemStatus struct {
	Platform     string `json:"platform"`
	NSMInstalled bool   `json:"nsm_installed"`
	Dependencies struct {
		Mkcert   bool `json:"mkcert"`
		Dnsmasq  bool `json:"dnsmasq"`
		Homebrew bool `json:"homebrew,omitempty"`
	} `json:"dependencies"`
	TLDs      []TLDConfig `json:"tlds"`
	ConfigDir string      `json:"config_dir"`
	LastSetup string      `json:"last_setup,omitempty"`
}

func NewConfig() *Config {
	return &Config{
		Platform: runtime.GOOS,
		TLDs:     []string{"dev", "test", "local", "app"},
	}
}
