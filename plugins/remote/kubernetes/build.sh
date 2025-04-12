#!/bin/bash

# Build script for Kubernetes plugin
# Optimized for building Linux binaries on Mac M4

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building Kubernetes plugin for Corynth...${NC}"

# Check if we're on a Mac
if [[ "$(uname)" != "Darwin" ]]; then
    echo -e "${RED}This script is optimized for Mac. Please use the Makefile directly on other platforms.${NC}"
    exit 1
fi

# Check for required tools
echo -e "${YELLOW}Checking for required tools...${NC}"

# Check for Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Please install Go before continuing.${NC}"
    exit 1
fi

# Check for cross-compiler for Linux
if ! command -v x86_64-linux-gnu-gcc &> /dev/null; then
    echo -e "${YELLOW}Cross-compiler for Linux (x86_64-linux-gnu-gcc) not found.${NC}"
    echo -e "${YELLOW}Installing cross-compiler using Homebrew...${NC}"
    
    # Check for Homebrew
    if ! command -v brew &> /dev/null; then
        echo -e "${RED}Homebrew is not installed. Please install Homebrew or the cross-compiler manually.${NC}"
        exit 1
    fi
    
    brew install FiloSottile/musl-cross/musl-cross
    
    if ! command -v x86_64-linux-gnu-gcc &> /dev/null; then
        echo -e "${RED}Failed to install cross-compiler. Please install it manually.${NC}"
        exit 1
    fi
fi

# Create build directory
mkdir -p build

# Build for Linux (optimized for Mac M4)
echo -e "${YELLOW}Building plugin for Linux...${NC}"
make linux

# Build for macOS (optional)
echo -e "${YELLOW}Building plugin for macOS...${NC}"
make darwin

# Package the plugin
echo -e "${YELLOW}Packaging plugin...${NC}"
make package

echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}Plugin package is available at: build/kubernetes.tar.gz${NC}"
echo ""
echo -e "${YELLOW}To deploy this plugin to a remote repository:${NC}"
echo "1. Upload the build/kubernetes.tar.gz file to your repository"
echo "2. Update the plugins.yaml file with the correct repository URL and version"
echo "3. Ensure the plugin is accessible via the URL specified in plugins.yaml"
echo ""
echo -e "${YELLOW}Example plugins.yaml entry:${NC}"
echo "plugins:"
echo "  - name: \"kubernetes\""
echo "    repository: \"https://github.com/yourusername/corynth-plugins\""
echo "    version: \"v1.0.0\""
echo "    path: \"kubernetes\""

# Make the plugin executable
chmod +x build/kubernetes.so

echo ""
echo -e "${GREEN}Done!${NC}"