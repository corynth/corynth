package main

import (
	"fmt"
	"os"
	
	"github.com/corynth/corynth/pkg/cli"
	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
	Commit    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "corynth",
		Short: "Workflow orchestration platform",
		Long:  `Corynth is a powerful workflow orchestration platform built in Go that enables you to define, execute, and manage complex multi-step workflows using HCL (HashiCorp Configuration Language).`,
		Version: fmt.Sprintf("%s (built %s, commit %s)", Version, BuildDate, Commit),
	}

	// Core workflow commands
	rootCmd.AddCommand(cli.NewInitCommand())
	rootCmd.AddCommand(cli.NewValidateCommand())
	rootCmd.AddCommand(cli.NewPlanCommand())
	rootCmd.AddCommand(cli.NewApplyCommand())
	
	// State management
	rootCmd.AddCommand(cli.NewStateCommand())
	
	// Plugin management
	rootCmd.AddCommand(cli.NewPluginCommand())
	
	// Configuration
	rootCmd.AddCommand(cli.NewConfigCommand())
	
	// Sample workflows
	rootCmd.AddCommand(cli.NewSampleCommand())
	
	// Newbie experience commands
	rootCmd.AddCommand(cli.NewStartCommand())
	rootCmd.AddCommand(cli.NewGalleryCommand())
	
	// Add version command explicitly
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Corynth version %s (built %s, commit %s)\n", Version, BuildDate, Commit)
		},
	}
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}