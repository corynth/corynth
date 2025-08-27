#!/bin/bash
set -e

# Corynth Quick Installation Script
# This script automatically downloads and installs the latest version of Corynth

VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
GITHUB_REPO="corynth/corynth"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    case $OS in
        linux)
            PLATFORM="linux-$ARCH"
            ;;
        darwin)
            PLATFORM="darwin-$ARCH"
            ;;
        mingw*|msys*|cygwin*)
            PLATFORM="windows-amd64"
            ;;
        *)
            echo -e "${RED}Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac
}

# Get the latest release version from GitHub
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [ -z "$VERSION" ]; then
            echo -e "${YELLOW}Could not determine latest version, using v1.0.0${NC}"
            VERSION="v1.0.0"
        fi
    fi
}

# Download the binary
download_binary() {
    BINARY_NAME="corynth-$PLATFORM"
    if [[ "$PLATFORM" == "windows-amd64" ]]; then
        BINARY_NAME="$BINARY_NAME.exe"
    fi
    
    DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$VERSION/$BINARY_NAME.gz"
    
    echo -e "${GREEN}Downloading Corynth $VERSION for $PLATFORM...${NC}"
    
    # Create temp directory
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"
    
    # Download the compressed binary
    if ! curl -LO "$DOWNLOAD_URL"; then
        echo -e "${RED}Failed to download from $DOWNLOAD_URL${NC}"
        echo -e "${YELLOW}Trying direct binary download...${NC}"
        
        # Try uncompressed binary as fallback
        DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$VERSION/$BINARY_NAME"
        if ! curl -LO "$DOWNLOAD_URL"; then
            echo -e "${RED}Failed to download Corynth${NC}"
            exit 1
        fi
        cp "$BINARY_NAME" corynth
    else
        # Decompress
        gunzip "$BINARY_NAME.gz"
        mv "$BINARY_NAME" corynth
    fi
    
    # Make executable
    chmod +x corynth
}

# Install the binary
install_binary() {
    echo -e "${GREEN}Installing Corynth to $INSTALL_DIR...${NC}"
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv corynth "$INSTALL_DIR/"
    else
        echo -e "${YELLOW}Need sudo access to install to $INSTALL_DIR${NC}"
        sudo mv corynth "$INSTALL_DIR/"
    fi
    
    # Clean up
    cd /
    rm -rf "$TMP_DIR"
}

# Verify installation
verify_installation() {
    if command -v corynth &> /dev/null; then
        echo -e "${GREEN}✓ Corynth installed successfully!${NC}"
        echo ""
        corynth version
        echo ""
        echo "To get started, run:"
        echo "  corynth init        # Initialize Corynth"
        echo "  corynth sample      # Generate sample workflow"
        echo "  corynth --help      # Show help"
    else
        echo -e "${RED}Installation failed. Please check the error messages above.${NC}"
        exit 1
    fi
}

# Main installation flow
main() {
    echo ""
    echo "================================"
    echo " Corynth Installation Script"
    echo "================================"
    echo ""
    
    detect_platform
    get_latest_version
    download_binary
    install_binary
    verify_installation
}

# Run main function
main