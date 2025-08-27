package pluginv2

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	
	"github.com/corynth/corynth/pkg/plugin"
)

// GRPCPluginManager manages gRPC-based plugins following Terraform's pattern
type GRPCPluginManager struct {
	localPath     string
	repositories  []plugin.Repository
	plugins       map[string]*GRPCClientPlugin
	mu            sync.RWMutex
}

// NewGRPCPluginManager creates a new gRPC plugin manager
func NewGRPCPluginManager(localPath string, repositories []plugin.Repository) *GRPCPluginManager {
	return &GRPCPluginManager{
		localPath:    localPath,
		repositories: repositories,
		plugins:      make(map[string]*GRPCClientPlugin),
	}
}

// Load loads a plugin by name, downloading if necessary
func (m *GRPCPluginManager) Load(name string) (plugin.Plugin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Return cached plugin if already loaded
	if grpcPlugin, exists := m.plugins[name]; exists {
		return grpcPlugin, nil
	}
	
	// Try to find local executable
	executablePath := m.findLocalExecutable(name)
	if executablePath == "" {
		// Try to install from repositories
		if err := m.installPlugin(name); err != nil {
			return nil, fmt.Errorf("plugin '%s' not found locally and installation failed: %w", name, err)
		}
		
		// Try to find executable again
		executablePath = m.findLocalExecutable(name)
		if executablePath == "" {
			return nil, fmt.Errorf("plugin '%s' installation completed but executable not found", name)
		}
	}
	
	// Create gRPC client plugin
	grpcPlugin, err := NewGRPCClientPlugin(name, executablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC plugin client: %w", err)
	}
	
	// Cache the plugin
	m.plugins[name] = grpcPlugin
	
	return grpcPlugin, nil
}

// findLocalExecutable finds the executable for a plugin
func (m *GRPCPluginManager) findLocalExecutable(name string) string {
	// Platform-specific executable names
	var candidates []string
	if runtime.GOOS == "windows" {
		candidates = []string{
			fmt.Sprintf("corynth-plugin-%s.exe", name),
			fmt.Sprintf("%s.exe", name),
		}
	} else {
		candidates = []string{
			fmt.Sprintf("corynth-plugin-%s", name),
			fmt.Sprintf("%s", name),
		}
	}
	
	// Search in local plugins directory
	for _, candidate := range candidates {
		path := filepath.Join(m.localPath, candidate)
		if _, err := os.Stat(path); err == nil {
			// Check if executable
			if isExecutable(path) {
				return path
			}
		}
	}
	
	return ""
}

// installPlugin downloads and installs a plugin from repositories
func (m *GRPCPluginManager) installPlugin(name string) error {
	for _, repo := range m.repositories {
		err := m.installFromRepository(name, repo)
		if err == nil {
			return nil // Successfully installed
		}
		// Continue to next repository on failure
	}
	
	return fmt.Errorf("plugin '%s' not found in any configured repository", name)
}

// installFromRepository installs a plugin from a specific repository
func (m *GRPCPluginManager) installFromRepository(name string, repo plugin.Repository) error {
	// Clone/update repository
	repoPath := filepath.Join(os.TempDir(), "corynth-plugins", repo.Name)
	if err := m.cloneOrUpdateRepo(repo.URL, repoPath, repo.Branch); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	
	// Look for plugin in various locations
	pluginPaths := []string{
		filepath.Join(repoPath, "official", name),
		filepath.Join(repoPath, "community", name), 
		filepath.Join(repoPath, name),
		filepath.Join(repoPath, "plugins", name),
	}
	
	for _, pluginPath := range pluginPaths {
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			continue
		}
		
		// Try to build and install plugin
		if err := m.buildAndInstallPlugin(name, pluginPath); err == nil {
			return nil // Successfully built and installed
		}
	}
	
	return fmt.Errorf("plugin '%s' source not found in repository '%s'", name, repo.Name)
}

// buildAndInstallPlugin compiles and installs a plugin
func (m *GRPCPluginManager) buildAndInstallPlugin(name, sourcePath string) error {
	// Determine executable name
	execName := fmt.Sprintf("corynth-plugin-%s", name)
	if runtime.GOOS == "windows" {
		execName += ".exe"
	}
	
	destPath := filepath.Join(m.localPath, execName)
	
	// Build the plugin
	var buildCmd *exec.Cmd
	
	// Check for Go source (main.go or plugin.go)
	goFiles := []string{"main.go", "plugin.go"}
	var goFile string
	for _, file := range goFiles {
		if _, err := os.Stat(filepath.Join(sourcePath, file)); err == nil {
			goFile = file
			break
		}
	}
	
	if goFile != "" {
		// Build Go plugin
		buildCmd = exec.Command("go", "build", "-o", destPath, goFile)
		buildCmd.Dir = sourcePath
	} else {
		// Check for other languages
		if _, err := os.Stat(filepath.Join(sourcePath, "main.py")); err == nil {
			// Python plugin - copy and make executable
			srcFile := filepath.Join(sourcePath, "main.py")
			if err := copyFile(srcFile, destPath); err != nil {
				return err
			}
			// Make executable
			return os.Chmod(destPath, 0755)
		} else if _, err := os.Stat(filepath.Join(sourcePath, "main.js")); err == nil {
			// Node.js plugin - copy and make executable
			srcFile := filepath.Join(sourcePath, "main.js")
			if err := copyFile(srcFile, destPath); err != nil {
				return err
			}
			// Make executable
			return os.Chmod(destPath, 0755)
		}
		
		return fmt.Errorf("no supported plugin source found (go, python, or node.js)")
	}
	
	// Execute build command
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %w\nOutput: %s", err, string(output))
	}
	
	// Make executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(destPath, 0755); err != nil {
			return fmt.Errorf("failed to make executable: %w", err)
		}
	}
	
	return nil
}

// List returns metadata for all loaded plugins
func (m *GRPCPluginManager) List() []plugin.Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var metadata []plugin.Metadata
	for _, grpcPlugin := range m.plugins {
		metadata = append(metadata, grpcPlugin.Metadata())
	}
	return metadata
}

// Close closes all plugin connections
func (m *GRPCPluginManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var errs []error
	for name, grpcPlugin := range m.plugins {
		if err := grpcPlugin.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close plugin %s: %w", name, err))
		}
	}
	
	// Clear plugin cache
	m.plugins = make(map[string]*GRPCClientPlugin)
	
	if len(errs) > 0 {
		return fmt.Errorf("errors closing plugins: %v", errs)
	}
	
	return nil
}

// cloneOrUpdateRepo clones or updates a git repository
func (m *GRPCPluginManager) cloneOrUpdateRepo(repoURL, localPath, branch string) error {
	// Check if repo already exists
	if _, err := os.Stat(filepath.Join(localPath, ".git")); err == nil {
		// Repository exists, update it
		repo, err := git.PlainOpen(localPath)
		if err != nil {
			return fmt.Errorf("failed to open existing repository: %w", err)
		}
		
		worktree, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get worktree: %w", err)
		}
		
		err = worktree.Pull(&git.PullOptions{
			ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("failed to pull repository: %w", err)
		}
	} else {
		// Repository doesn't exist, clone it
		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		
		_, err := git.PlainClone(localPath, false, &git.CloneOptions{
			URL:           repoURL,
			ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		})
		if err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
	}
	
	return nil
}

// isExecutable checks if a file is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	
	mode := info.Mode()
	if runtime.GOOS == "windows" {
		// On Windows, check file extension
		return strings.HasSuffix(strings.ToLower(path), ".exe")
	}
	
	// On Unix systems, check execute permissions
	return mode&0111 != 0
}

// copyFile copies a file from src to dst  
func copyFile(src, dst string) error {
	return fmt.Errorf("copyFile not implemented yet - use os.Link or io.Copy")
}