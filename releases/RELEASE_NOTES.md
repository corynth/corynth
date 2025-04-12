# 🚀 Corynth v0.1.0 Release Notes

## 📦 Overview

This is the initial release of Corynth, a powerful automation orchestration tool that enables users to define, plan, and execute sequential workflows through declarative YAML configurations.

## ✨ Features

- **Core Engine**: The main execution engine written in Go
- **Plugin System**: Extensible plugin architecture with three core plugins:
  - Git: Repository operations
  - Ansible: Infrastructure automation
  - Shell: Command execution
- **Flow Parser**: YAML parser for flow definitions
- **Dependency Resolver**: Handles step dependencies
- **Execution Manager**: Orchestrates flow execution
- **State Management**: Tracks execution state

## 📥 Downloads

- [corynth-linux-amd64](https://github.com/corynth/corynth/releases/download/v0.1.0/corynth-linux-amd64) - Linux (64-bit)
- [corynth-darwin-arm64](https://github.com/corynth/corynth/releases/download/v0.1.0/corynth-darwin-arm64) - macOS (Apple Silicon)

## 🔧 Installation

### Linux

```bash
# Download the binary
curl -LO https://github.com/corynth/corynth/releases/download/v0.1.0/corynth-linux-amd64

# Make it executable
chmod +x corynth-linux-amd64

# Move to your PATH
sudo mv corynth-linux-amd64 /usr/local/bin/corynth
```

### macOS (Apple Silicon)

```bash
# Download the binary
curl -LO https://github.com/corynth/corynth/releases/download/v0.1.0/corynth-darwin-arm64

# Make it executable
chmod +x corynth-darwin-arm64

# Move to your PATH
sudo mv corynth-darwin-arm64 /usr/local/bin/corynth
```

## 📝 Documentation

- [🚀 Getting Started](../docs/getting_started.md)
- [🔌 Plugin Development](../docs/plugin_development.md)
- [🌊 Advanced Flow Configuration](../docs/advanced_flows.md)
- [💾 State Management](../docs/state_management.md)

## 🐛 Known Issues

- None reported yet. Please submit issues on GitHub.

## 🙏 Acknowledgements

- Inspired by Terraform's workflow model
- Built with Go