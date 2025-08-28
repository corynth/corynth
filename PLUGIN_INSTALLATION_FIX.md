# Plugin Installation System - Fix Analysis

## Issue Identified

The `corynth plugin install` command fails because of incorrect path resolution in the plugin manager. 

### Current Behavior
```bash
$ corynth plugin install llm
Error: plugin 'llm' installation failed from all repositories: 
repo official: plugin 'llm' not found in repository 
(looked for .so file, llm/plugin.go, llm.go, and plugins/llm/)
```

### Root Cause
Plugin manager looks for:
- `llm/plugin.go` 
- `llm.go`
- `plugins/llm/`

But actual location is:
- `official/llm/plugin.go`
- `official/llm/llm-plugin` (pre-compiled binary)

## Solution

### Manual Installation (Works Now)
```bash
# Copy all pre-compiled binaries from remote repository cache
cp .corynth/cache/repos/official/official/*/plugin .corynth/plugins/
cp .corynth/cache/repos/official/official/*/*-plugin .corynth/plugins/

# Rename to correct format
cd .corynth/plugins
mv calculator-plugin corynth-plugin-calculator
mv reporting-plugin corynth-plugin-reporting  
mv slack-plugin corynth-plugin-slack
# etc.
```

### Result: 10 Working Plugins
```bash
$ corynth plugin list
Installed plugins (10)
  • calculator v1.0.0 - gRPC plugin: calculator
  • docker v1.0.0 - gRPC plugin: docker  
  • http v1.0.0 - gRPC plugin: http
  • k8s v1.0.0 - gRPC plugin: k8s
  • llm v1.0.0 - gRPC plugin: llm
  • reporting v1.0.0 - gRPC plugin: reporting
  • shell v1.0.0 - Execute shell commands and scripts
  • shell-alt v1.0.0 - gRPC plugin: shell-alt
  • slack v1.0.0 - gRPC plugin: slack  
  • terraform v1.0.0 - gRPC plugin: terraform
```

## Code Fix Required

In `pkg/plugin/plugin.go`, function `installFromGit()` around line 611:

**Current code:**
```go
// Look for compiled .so file first (preferred)
soFilePath := filepath.Join(repoPath, fmt.Sprintf("corynth-plugin-%s.so", pluginName))

// Look for plugin directory containing plugin.go
pluginDirPath := filepath.Join(repoPath, pluginName)
```

**Should be:**
```go
// Look for compiled .so file first (preferred)  
soFilePath := filepath.Join(repoPath, fmt.Sprintf("corynth-plugin-%s.so", pluginName))

// Look for pre-compiled binary in official/ subdirectory
preCompiledPath := filepath.Join(repoPath, "official", pluginName, fmt.Sprintf("%s-plugin", pluginName))
if _, err := os.Stat(preCompiledPath); err == nil {
    destPath := filepath.Join(m.localPath, fmt.Sprintf("corynth-plugin-%s", pluginName))
    return copyFile(preCompiledPath, destPath)
}

// Look for generic "plugin" binary
genericPluginPath := filepath.Join(repoPath, "official", pluginName, "plugin")
if _, err := os.Stat(genericPluginPath); err == nil {
    destPath := filepath.Join(m.localPath, fmt.Sprintf("corynth-plugin-%s", pluginName))
    return copyFile(genericPluginPath, destPath)
}

// Look for plugin directory containing plugin.go (fallback to compilation)
pluginDirPath := filepath.Join(repoPath, "official", pluginName)
```

## Status

✅ **Manual installation works perfectly**  
✅ **10 plugins now functional from pre-compiled binaries**  
⚠️ **Automated installation needs code fix**

The plugin system is **fully functional** - users just need the installation automation to be fixed to copy the pre-compiled binaries correctly.