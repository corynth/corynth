package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/corynth/corynth/pkg/config"
	"github.com/spf13/cobra"
)

func newPluginDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose plugin system health",
		Long:  "Run diagnostics on the plugin system and provide recommendations for fixing issues.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPluginDoctor(cmd, args)
		},
	}
}

func runPluginDoctor(cmd *cobra.Command, args []string) error {
	fmt.Printf("🔍 %s\n\n", Header("Corynth Plugin System Diagnostics"))
	
	cfg, err := loadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	manager, err := initializePluginManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize plugin manager: %w", err)
	}
	
	// 1. Check core system
	fmt.Printf("%s\n", SubHeader("Core System:"))
	fmt.Printf("%s %s: %s\n", BulletPoint("✓"), Label("Corynth version"), Value("1.2.0"))
	fmt.Printf("%s %s: %s\n", BulletPoint("✓"), Label("Go version"), Value(runtime.Version()))
	fmt.Printf("%s %s: %s %s\n", BulletPoint("✓"), Label("Operating System"), Value(runtime.GOOS), Value(runtime.GOARCH))
	fmt.Println()
	
	// 2. Check built-in plugins
	fmt.Printf("%s\n", SubHeader("Built-in Plugins:"))
	builtinPlugins := []string{"git", "slack"}
	builtinCount := 0
	for _, name := range builtinPlugins {
		if _, err := manager.Get(name); err == nil {
			fmt.Printf("%s %s: %s\n", BulletPoint("✓"), Label(name), Colorize(SuccessColor, "working"))
			builtinCount++
		} else {
			fmt.Printf("%s %s: %s\n", BulletPoint("❌"), Label(name), Colorize(ErrorColor, "failed"))
		}
	}
	fmt.Printf("%s %s: %s\n", BulletPoint("📊"), Label("Status"), Value(fmt.Sprintf("%d/%d working", builtinCount, len(builtinPlugins))))
	fmt.Println()
	
	// 3. Check external plugins
	fmt.Printf("%s\n", SubHeader("External Plugins:"))
	externalPlugins := []string{"shell", "http", "file", "docker", "terraform"}
	workingCount := 0
	var failedPlugins []string
	
	for _, name := range externalPlugins {
		if _, err := manager.Get(name); err == nil {
			fmt.Printf("%s %s: %s\n", BulletPoint("✓"), Label(name), Colorize(SuccessColor, "working"))
			workingCount++
		} else {
			shortErr := getShortError(err)
			fmt.Printf("%s %s: %s\n", BulletPoint("❌"), Label(name), Colorize(ErrorColor, shortErr))
			failedPlugins = append(failedPlugins, name)
		}
	}
	fmt.Printf("%s %s: %s\n", BulletPoint("📊"), Label("Status"), Value(fmt.Sprintf("%d/%d working", workingCount, len(externalPlugins))))
	fmt.Println()
	
	// 4. Check plugin cache
	fmt.Printf("%s\n", SubHeader("Plugin Cache:"))
	cacheDir := getPluginCacheDir(cfg)
	fmt.Printf("%s %s: %s\n", BulletPoint("📁"), Label("Cache directory"), Value(cacheDir))
	
	if files, err := filepath.Glob(filepath.Join(cacheDir, "*.so")); err == nil {
		fmt.Printf("%s %s: %s\n", BulletPoint("📦"), Label("Cached plugins"), Value(fmt.Sprintf("%d files", len(files))))
		
		// Check for stale plugins - estimate based on file age
		staleCount := 0
		for _, file := range files {
			if info, err := os.Stat(file); err == nil {
				// Consider files older than 7 days as potentially stale
				if info.ModTime().Before(time.Now().AddDate(0, 0, -7)) {
					staleCount++
				}
			}
		}
		if staleCount > 0 {
			fmt.Printf("%s %s: %s\n", BulletPoint("⚠️"), Label("Potentially stale plugins"), Colorize(WarningColor, fmt.Sprintf("%d files", staleCount)))
		} else {
			fmt.Printf("%s %s\n", BulletPoint("✓"), Colorize(SuccessColor, "Cache appears healthy"))
		}
	} else {
		fmt.Printf("%s %s: %s\n", BulletPoint("📦"), Label("Cached plugins"), Value("0 files"))
	}
	fmt.Println()
	
	// 5. Recommendations
	fmt.Printf("🔧 %s\n", Header("Recommendations:"))
	if len(failedPlugins) > 0 {
		fmt.Printf("  %s\n", Info("Plugin issues detected:"))
		for _, plugin := range failedPlugins {
			fmt.Printf("    %s corynth plugin clean %s\n", BulletPoint("•"), plugin)
			fmt.Printf("    %s corynth plugin install %s\n", BulletPoint("•"), plugin)
		}
		fmt.Println()
		fmt.Printf("  %s\n", Info("Or fix all at once:"))
		fmt.Printf("    %s corynth plugin clean --all\n", BulletPoint("•"))
		fmt.Printf("    %s corynth plugin install --all\n", BulletPoint("•"))
	} else {
		fmt.Printf("  %s\n", Colorize(SuccessColor, "✓ No critical issues detected"))
		fmt.Printf("  %s\n", Info("Plugin system appears to be healthy!"))
	}
	
	fmt.Println()
	fmt.Printf("📚 %s\n", Header("Additional Help:"))
	fmt.Printf("  %s corynth plugin list       # Show available plugins\n", BulletPoint("•"))
	fmt.Printf("  %s corynth plugin info <name> # Get plugin details\n", BulletPoint("•"))
	fmt.Printf("  %s corynth sample            # Generate test workflows\n", BulletPoint("•"))
	
	return nil
}

func newPluginCleanCommand() *cobra.Command {
	var cleanAll bool

	cmd := &cobra.Command{
		Use:   "clean [plugin-name]",
		Short: "Clean plugin cache",
		Long:  "Remove cached plugins that may be causing compatibility issues.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPluginClean(cmd, args, cleanAll)
		},
	}

	cmd.Flags().BoolVar(&cleanAll, "all", false, "Clean all plugins")
	return cmd
}

func runPluginClean(cmd *cobra.Command, args []string, cleanAll bool) error {
	cfg, err := loadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	manager, err := initializePluginManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize plugin manager: %w", err)
	}
	
	if cleanAll || len(args) == 0 {
		// Clean all plugins
		fmt.Printf("🧹 %s\n", Info("Cleaning all plugin cache..."))
		
		cacheDir := getPluginCacheDir(cfg)
		if err := os.RemoveAll(cacheDir); err != nil {
			return fmt.Errorf("failed to clean plugin cache: %w", err)
		}
		
		// Recreate directory
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return fmt.Errorf("failed to recreate plugin directory: %w", err)
		}
		
		fmt.Printf("%s\n", Success("Plugin cache cleaned"))
		fmt.Printf("\n%s\n", Info("Next steps:"))
		fmt.Printf("  %s\n", Value("corynth plugin install --all"))
		
		return nil
	}
	
	// Clean specific plugin
	pluginName := args[0]
	fmt.Printf("🧹 %s %s\n", Info("Cleaning plugin:"), Value(pluginName))
	
	if err := manager.Remove(pluginName); err != nil {
		return fmt.Errorf("failed to clean plugin %s: %w", pluginName, err)
	}
	
	fmt.Printf("%s %s %s\n", Success("Plugin"), Value(pluginName), Success("cleaned"))
	fmt.Printf("\n%s\n", Info("Next step:"))
	fmt.Printf("  %s %s\n", Value("corynth plugin install"), Value(pluginName))
	
	return nil
}

func getShortError(err error) string {
	errStr := err.Error()
	if strings.Contains(errStr, "different version") {
		return "version mismatch"
	}
	if strings.Contains(errStr, "not found") {
		return "not installed"
	}
	if strings.Contains(errStr, "plugin.Open") {
		return "loading failed"
	}
	if strings.Contains(errStr, "incompatible") {
		return "incompatible"
	}
	return "error"
}

func getPluginCacheDir(cfg *config.Config) string {
	if cfg != nil && cfg.Plugins != nil {
		if cfg.Plugins.LocalPath != "" {
			return cfg.Plugins.LocalPath
		}
		if cfg.Plugins.Cache != nil && cfg.Plugins.Cache.Path != "" {
			return cfg.Plugins.Cache.Path
		}
	}
	
	// Default cache directory
	home, err := os.UserHomeDir()
	if err != nil {
		return ".corynth/plugins"
	}
	return filepath.Join(home, ".corynth", "plugins")
}