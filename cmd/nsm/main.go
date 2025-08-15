package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/kashifsb/nsm/internal/config"
	"github.com/kashifsb/nsm/internal/ui"
	"github.com/kashifsb/nsm/pkg/logger"
)

var (
	version = "1.0.0"
	commit  = "dev"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rootCmd := &cobra.Command{
		Use:     "nsm",
		Short:   "Enterprise Development Environment Manager",
		Long:    "NSM provides clean URLs, automatic HTTPS, and professional DNS resolution for development environments",
		Version: fmt.Sprintf("%s (%s)", version, commit),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNSM(ctx, cmd)
		},
	}

	// Add flags
	rootCmd.Flags().StringP("project-type", "t", "", "Project type (vite, react, go, rust, python, java, dotnet)")
	rootCmd.Flags().StringP("domain", "d", "", "Custom domain (e.g., api.dev)")
	rootCmd.Flags().StringP("command", "c", "", "Development command to run")
	rootCmd.Flags().IntP("http-port", "p", 0, "HTTP port (0 = auto)")
	rootCmd.Flags().IntP("https-port", "s", 0, "HTTPS port (0 = auto, prefers 443)")
	rootCmd.Flags().BoolP("no-443", "n", false, "Don't use port 443")
	rootCmd.Flags().BoolP("debug", "v", false, "Enable debug logging")
	rootCmd.Flags().BoolP("headless", "h", false, "Run without interactive UI")
	rootCmd.Flags().BoolP("auto-yes", "y", false, "Auto-confirm prompts")

	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}

func runNSM(ctx context.Context, cmd *cobra.Command) error {
	// Initialize logger
	debug, _ := cmd.Flags().GetBool("debug")
	logger.Init(debug)

	// Parse configuration
	cfg, err := config.ParseFromFlags(cmd)
	if err != nil {
		return fmt.Errorf("parse configuration: %w", err)
	}

	// Check if headless mode
	headless, _ := cmd.Flags().GetBool("headless")
	if headless {
		return runHeadless(ctx, cfg)
	}

	// Run interactive UI
	return runInteractive(ctx, cfg)
}

func runInteractive(ctx context.Context, cfg *config.Config) error {
	model := ui.NewModel(cfg)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	go func() {
		<-ctx.Done()
		p.Send(ui.ShutdownMsg{})
	}()

	_, err := p.Run()
	return err
}

func runHeadless(ctx context.Context, cfg *config.Config) error {
	// Direct execution without UI
	logger.Info("Running in headless mode")
	// Implementation for headless mode
	return nil
}
