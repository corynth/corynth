package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/corynth/corynth/cmd/apply"
	"github.com/corynth/corynth/cmd/initialize"
	"github.com/corynth/corynth/cmd/plan"
)

const (
	version = "0.1.0"
)

func printUsage() {
	fmt.Println("Corynth - Workflow Automation Orchestration Tool")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  corynth [command] [arguments]")
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  init [dir]                   Initialize a new Corynth project")
	fmt.Println("  plan [dir]                   Plan execution and validate flows")
	fmt.Println("  apply [dir] [flow-name]      Execute all flows or a specific flow")
	fmt.Println("  version                      Print the version number")
	fmt.Println("  help                         Print this help message")
	fmt.Println("\nFlags:")
	fmt.Println("  --help, -h                   Print help information")
	fmt.Println("\nExamples:")
	fmt.Println("  corynth init ./my_project")
	fmt.Println("  corynth plan ./my_project")
	fmt.Println("  corynth apply ./my_project deployment_flow")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		dir := "."
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		absDir, err := filepath.Abs(dir)
		if err != nil {
			fmt.Printf("Error resolving path: %s\n", err)
			os.Exit(1)
		}
		initialize.Execute(absDir)

	case "plan":
		dir := "."
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		absDir, err := filepath.Abs(dir)
		if err != nil {
			fmt.Printf("Error resolving path: %s\n", err)
			os.Exit(1)
		}
		plan.Execute(absDir)

	case "apply":
		dir := "."
		flowName := ""
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		if len(os.Args) > 3 {
			flowName = os.Args[3]
		}
		absDir, err := filepath.Abs(dir)
		if err != nil {
			fmt.Printf("Error resolving path: %s\n", err)
			os.Exit(1)
		}
		apply.Execute(absDir, flowName)

	case "version":
		fmt.Printf("Corynth version %s\n", version)

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}