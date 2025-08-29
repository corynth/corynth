#!/bin/bash
# Corynth Production Installation Script
# Installs the latest Corynth release with security verification

set -e

# Configuration
GITHUB_REPO="corynth/corynth"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="corynth"
TMP_DIR=$(mktemp -d)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Detect platform and architecture
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $os in
        linux)   PLATFORM="linux" ;;
        darwin)  PLATFORM="darwin" ;;
        mingw*|cygwin*|msys*) PLATFORM="windows" ;;
        *)       log_error "Unsupported operating system: $os" ;;
    esac
    
    case $arch in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)           log_error "Unsupported architecture: $arch" ;;
    esac
    
    BINARY_NAME="corynth-${PLATFORM}-${ARCH}"
    if [ "$PLATFORM" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if curl is available
    if ! command -v curl >/dev/null 2>&1; then
        log_error "curl is required but not installed"
    fi
    
    # Check if we can write to install directory
    if [ ! -w "$INSTALL_DIR" ] && [ "$EUID" -ne 0 ]; then
        log_warning "Installation directory requires sudo privileges"
        NEEDS_SUDO=true
    fi
    
    log_success "Prerequisites check passed"
}

# Get latest release info
get_latest_release() {
    log_info "Getting latest release information..."
    
    local api_url="https://api.github.com/repos/$GITHUB_REPO/releases/latest"
    local release_info=$(curl -s "$api_url")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to fetch release information"
    fi
    
    VERSION=$(echo "$release_info" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        log_error "Could not determine latest version"
    fi
    
    log_info "Latest version: $VERSION"
}

# Download binary and checksums
download_files() {
    log_info "Downloading Corynth $VERSION for $PLATFORM-$ARCH..."
    
    local base_url="https://github.com/$GITHUB_REPO/releases/download/$VERSION"
    local binary_url="${base_url}/${BINARY_NAME}"
    local checksums_url="${base_url}/checksums.txt"
    
    # Download binary
    if ! curl -L -o "${TMP_DIR}/${BINARY_NAME}" "$binary_url"; then
        log_error "Failed to download $BINARY_NAME"
    fi
    
    # Download checksums
    if ! curl -L -o "${TMP_DIR}/checksums.txt" "$checksums_url"; then
        log_error "Failed to download checksums.txt"
    fi
    
    log_success "Downloaded files to temporary directory"
}

# Verify checksums
verify_checksums() {
    log_info "Verifying file integrity..."
    
    cd "$TMP_DIR"
    
    # Extract expected checksum for our binary
    local expected_checksum=$(grep "$BINARY_NAME" checksums.txt | cut -d' ' -f1)
    
    if [ -z "$expected_checksum" ]; then
        log_error "Checksum not found for $BINARY_NAME"
    fi
    
    # Calculate actual checksum
    local actual_checksum=$(shasum -a 256 "$BINARY_NAME" | cut -d' ' -f1)
    
    if [ "$expected_checksum" != "$actual_checksum" ]; then
        log_error "Checksum verification failed!"
    fi
    
    log_success "File integrity verified"
}

# Install binary
install_binary() {
    log_info "Installing Corynth..."
    
    local source_file="${TMP_DIR}/${BINARY_NAME}"
    local dest_file="${INSTALL_DIR}/corynth"
    
    # Make executable
    chmod +x "$source_file"
    
    # Install (with sudo if needed)
    if [ "$NEEDS_SUDO" = true ]; then
        log_info "Installing to $dest_file (requires sudo)..."
        sudo cp "$source_file" "$dest_file"
        sudo chmod +x "$dest_file"
    else
        log_info "Installing to $dest_file..."
        cp "$source_file" "$dest_file"
        chmod +x "$dest_file"
    fi
    
    log_success "Corynth installed successfully"
}

# Verify installation
verify_installation() {
    log_info "Verifying installation..."
    
    if ! command -v corynth >/dev/null 2>&1; then
        log_error "Corynth not found in PATH after installation"
    fi
    
    local installed_version=$(corynth version 2>/dev/null | head -1)
    log_info "Installed: $installed_version"
    
    log_success "Installation verified"
}

# Cleanup
cleanup() {
    log_info "Cleaning up temporary files..."
    rm -rf "$TMP_DIR"
}

# Show completion message
show_completion() {
    echo ""
    log_success "🎉 Corynth installation complete!"
    echo ""
    echo "📚 Quick Start:"
    echo "  corynth --help                    # Show help"
    echo "  corynth sample --template hello-world  # Generate sample"
    echo "  corynth apply hello-world.hcl     # Run workflow"
    echo ""
    echo "🔌 Plugin System:"
    echo "  corynth plugin list               # List plugins"
    echo "  corynth plugin discover           # Find available plugins"
    echo ""
    echo "📖 Documentation: https://github.com/$GITHUB_REPO"
    echo ""
}

# Handle script interruption
trap cleanup EXIT INT TERM

# Main installation flow
main() {
    echo "🚀 Corynth Production Installer"
    echo "==============================="
    echo ""
    
    detect_platform
    log_info "Detected platform: $PLATFORM-$ARCH"
    
    check_prerequisites
    get_latest_release
    download_files
    verify_checksums
    install_binary
    verify_installation
    show_completion
}

# Run main function
main "$@"