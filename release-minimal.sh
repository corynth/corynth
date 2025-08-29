#!/bin/bash
# Build minimal Corynth releases for all platforms
# Only shell plugin is built-in, all others are remote

set -e

VERSION=${VERSION:-"1.0.1-minimal"}
BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS="-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.Commit=${COMMIT}"

# Release directory
RELEASE_DIR="release-minimal"
mkdir -p $RELEASE_DIR

echo "Building minimal Corynth releases (shell-only)..."
echo "Version: $VERSION"
echo ""

# Build for all platforms
platforms=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a platform_split <<< "$platform"
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    
    output_name="corynth-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "Building for ${GOOS}/${GOARCH}..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "${LDFLAGS}" -o "${RELEASE_DIR}/${output_name}" ./cmd/corynth
    
    # Compress the binary
    echo "Compressing ${output_name}..."
    gzip -c "${RELEASE_DIR}/${output_name}" > "${RELEASE_DIR}/${output_name}.gz"
done

# Create checksums
echo ""
echo "Creating checksums..."
cd $RELEASE_DIR
shasum -a 256 *.gz > checksums.txt
cd ..

echo ""
echo "✅ Minimal releases built successfully in ${RELEASE_DIR}/"
echo ""
echo "Files created:"
ls -lh ${RELEASE_DIR}/

echo ""
echo "These releases include ONLY the shell plugin as built-in."
echo "All other plugins must be installed from remote repositories."