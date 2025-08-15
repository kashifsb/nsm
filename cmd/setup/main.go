package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kashifsb/nsm/internal/setup"
	"github.com/kashifsb/nsm/pkg/logger"
)

var (
	version = "3.0.0"
	commit  = "dev"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rootCmd := &cobra.Command{
		Use:     "nsm-setup",
		Short:   "NSM Setup & Configuration Tool",
		Long:    "Enterprise development environment setup and configuration manager for NSM",
		Version: fmt.Sprintf("%s (%s)", version, commit),
	}

	// Setup command
	setupCmd := &cobra.Command{
		Use:   "install",
		Short: "Install and configure NSM development environment",
		Long:  "Installs dependencies, configures DNS, and sets up local development environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(ctx, cmd)
		},
	}

	// TLD management commands
	tldCmd := &cobra.Command{
		Use:   "tld",
		Short: "Manage Top-Level Domains",
		Long:  "Add, remove, and list configured development TLDs",
	}

	tldAddCmd := &cobra.Command{
		Use:   "add [tld]",
		Short: "Add a new development TLD",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTLDAdd(ctx, args[0])
		},
	}

	tldRemoveCmd := &cobra.Command{
		Use:   "remove [tld]",
		Short: "Remove a development TLD",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTLDRemove(ctx, args[0])
		},
	}

	tldListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured TLDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTLDList(ctx)
		},
	}

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check NSM environment status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(ctx)
		},
	}

	// Reset command
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset NSM configuration",
		Long:  "Remove all NSM configuration and return system to original state",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReset(ctx)
		},
	}

	// Example command
	exampleCmd := &cobra.Command{
		Use:   "example [framework]",
		Short: "Create example project",
		Long:  "Create a new example project for the specified framework",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateExample(ctx, args[0])
		},
	}

	// Add flags
	setupCmd.Flags().Bool("headless", false, "Run setup without interactive UI")
	setupCmd.Flags().Bool("auto-yes", false, "Auto-confirm all prompts")
	setupCmd.Flags().Bool("skip-deps", false, "Skip dependency installation")
	setupCmd.Flags().StringSlice("tlds", []string{"dev", "test", "local"}, "TLDs to configure")

	statusCmd.Flags().Bool("json", false, "Output status as JSON")

	resetCmd.Flags().Bool("force", false, "Force reset without confirmation")

	exampleCmd.Flags().String("name", "", "Project name (default: auto-generated)")
	exampleCmd.Flags().String("path", ".", "Output directory")

	// Build command tree
	tldCmd.AddCommand(tldAddCmd, tldRemoveCmd, tldListCmd)
	rootCmd.AddCommand(setupCmd, tldCmd, statusCmd, resetCmd, exampleCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}

func runSetup(ctx context.Context, cmd *cobra.Command) error {
	headless, _ := cmd.Flags().GetBool("headless")
	autoYes, _ := cmd.Flags().GetBool("auto-yes")
	skipDeps, _ := cmd.Flags().GetBool("skip-deps")
	tlds, _ := cmd.Flags().GetStringSlice("tlds")

	logger.Init(false)
	logger.Info("Starting NSM setup", "platform", runtime.GOOS)

	config := setup.Config{
		Headless: headless,
		AutoYes:  autoYes,
		SkipDeps: skipDeps,
		TLDs:     tlds,
		Platform: runtime.GOOS,
	}

	if headless {
		return setup.RunHeadless(ctx, config)
	}

	return setup.RunInteractive(ctx, config)
}

func runTLDAdd(ctx context.Context, tld string) error {
	logger.Init(false)
	return setup.AddTLD(ctx, tld)
}

func runTLDRemove(ctx context.Context, tld string) error {
	logger.Init(false)
	return setup.RemoveTLD(ctx, tld)
}

func runTLDList(ctx context.Context) error {
	logger.Init(false)
	return setup.ListTLDs(ctx)
}

func runStatus(ctx context.Context) error {
	logger.Init(false)
	return setup.ShowStatus(ctx)
}

func runReset(ctx context.Context) error {
	logger.Init(false)
	return setup.Reset(ctx)
}

func runCreateExample(ctx context.Context, framework string) error {
	logger.Init(false)
	return setup.CreateExample(ctx, framework)
}
