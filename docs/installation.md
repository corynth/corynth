# Installation Guide

This guide covers all installation methods for Corynth on different platforms.

## System Requirements

- **Operating System**: macOS, Linux, or Windows
- **Architecture**: AMD64 or ARM64
- **Dependencies**: Git (required for plugin auto-installation)
- **Optional**: Go 1.21+ (for building from source)

## Installation Methods

### 1. Download Release Binaries (Recommended)

#### macOS

**Apple Silicon (M1/M2/M3)**:
```bash
curl -L https://github.com/corynth/corynth/releases/download/v1.2.0/corynth-darwin-arm64.tar.gz | tar -xz
sudo mv corynth /usr/local/bin/
```

**Intel**:
```bash
curl -L https://github.com/corynth/corynth/releases/download/v1.2.0/corynth-darwin-amd64.tar.gz | tar -xz
sudo mv corynth /usr/local/bin/
```

#### Linux

**AMD64**:
```bash
curl -L https://github.com/corynth/corynth/releases/download/v1.2.0/corynth-linux-amd64.tar.gz | tar -xz
sudo mv corynth /usr/local/bin/
```

**ARM64**:
```bash
curl -L https://github.com/corynth/corynth/releases/download/v1.2.0/corynth-linux-arm64.tar.gz | tar -xz
sudo mv corynth /usr/local/bin/
```

#### Windows

**AMD64**:
```bash
curl -L https://github.com/corynth/corynth/releases/download/v1.2.0/corynth-windows-amd64.exe.tar.gz | tar -xz
# Move corynth.exe to a directory in your PATH
```

### 2. Build from Source

**Prerequisites**: Go 1.21+, Git, Make

```bash
# Clone repository
git clone https://github.com/corynth/corynth.git
cd corynth

# Install dependencies
go mod tidy

# Build binary
make build

# Install globally (optional)
sudo mv corynth /usr/local/bin/
```

### 3. Homebrew (macOS) - Coming Soon

```bash
brew install justynroberts/tap/corynth
```

### 4. Package Managers - Coming Soon

- **apt** (Ubuntu/Debian)
- **yum/dnf** (RHEL/CentOS/Fedora)
- **chocolatey** (Windows)

## Verification

After installation, verify Corynth is working:

```bash
# Check version
corynth version

# Expected output:
# Corynth v1.2.0
# Built with Go 1.21+
# Platform: darwin/arm64
```

## Initial Setup

### 1. Initialize Your First Project

```bash
mkdir my-corynth-project
cd my-corynth-project
corynth init
```

### 2. Verify Installation

```bash
# Check available plugins
corynth plugin list

# Expected output:
# Installed plugins (2)
#   • git v1.0.0 - Git version control operations
#   • slack v1.0.0 - Slack messaging and notification operations
```

### 3. Test with Sample Workflow

```bash
# Generate sample workflow
corynth sample --template hello-world

# Execute the sample
corynth apply hello-world.hcl
```

## Configuration

### Environment Variables

Set these optional environment variables:

```bash
# Plugin repository (default: official repository)
export CORYNTH_PLUGIN_REPO="https://github.com/corynth/corynthplugins.git"

# State directory (default: ~/.corynth/state)
export CORYNTH_STATE_DIR="~/.corynth/state"

# Log level (default: info)
export CORYNTH_LOG_LEVEL="debug"

# Disable color output
export CORYNTH_NO_COLOR="true"
```

### Shell Completion

Add shell completion for better CLI experience:

**Bash**:
```bash
echo 'eval "$(corynth completion bash)"' >> ~/.bashrc
source ~/.bashrc
```

**Zsh**:
```bash
echo 'eval "$(corynth completion zsh)"' >> ~/.zshrc
source ~/.zshrc
```

**Fish**:
```bash
corynth completion fish | source
```

## Troubleshooting

### Common Issues

#### 1. Command Not Found
```bash
corynth: command not found
```
**Solution**: Ensure `/usr/local/bin` is in your PATH or move the binary to a directory in your PATH.

#### 2. Permission Denied
```bash
permission denied: corynth
```
**Solution**: Make the binary executable:
```bash
chmod +x /path/to/corynth
```

#### 3. Plugin Installation Fails
```bash
Error: failed to compile plugin
```
**Solution**: Ensure Git is installed and you have internet connectivity:
```bash
git --version
ping github.com
```

#### 4. Missing Dependencies
```bash
Error: go: command not found
```
**Solution**: Go is only required for building from source. Use release binaries instead, or install Go from [golang.org](https://golang.org/doc/install).

### Platform-Specific Notes

#### macOS
- **Gatekeeper**: You may need to allow the binary in System Preferences → Security & Privacy
- **Homebrew**: Use Homebrew installation method when available for easier updates

#### Linux
- **Distribution packages**: Use your distribution's package manager when available
- **AppImage**: Portable AppImage packages coming soon

#### Windows
- **PowerShell**: Recommended for best experience
- **WSL**: Fully supported in Windows Subsystem for Linux
- **PATH**: Add the binary location to your system PATH

### Getting Help

If you encounter issues:

1. Check this troubleshooting section
2. Review the [FAQ](../user-guide/faq.md)
3. Search [GitHub Issues](https://github.com/corynth/corynth/issues)
4. Create a new issue with:
   - Operating system and version
   - Installation method used
   - Complete error message
   - Output of `corynth version` (if available)

## Next Steps

After successful installation:

1. Read the [Quick Start Guide](quick-start.md)
2. Follow the [First Workflow Tutorial](first-workflow.md)
3. Explore [Example Workflows](../examples/)
4. Learn about [Plugin Development](../plugins/development.md)

---

**Installation complete! Ready to start orchestrating workflows with Corynth.** 🚀