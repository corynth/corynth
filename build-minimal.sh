#!/bin/bash
# Build Corynth with only shell plugin built-in
# All other plugins should be loaded remotely

set -e

echo "Building minimal Corynth with shell-only plugin..."

# Set version info
VERSION=${VERSION:-"1.0.1-minimal"}
BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS="-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.Commit=${COMMIT}"

# Build directory
BUILD_DIR="bin-minimal"
mkdir -p $BUILD_DIR

echo "Building for current platform..."
go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/corynth ./cmd/corynth

echo "✅ Minimal build complete: ${BUILD_DIR}/corynth"
echo ""
echo "Testing plugin list..."
${BUILD_DIR}/corynth plugin list

echo ""
echo "This build includes ONLY the shell plugin as built-in."
echo "All other plugins (http, file, git, slack, reporting) must be installed remotely."