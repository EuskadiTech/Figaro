#!/bin/bash
# Build script for Figaro Go application

set -e

echo "🏗️ Building Figaro..."

# Set build variables
VERSION=${VERSION:-"dev"}
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS="-s -w -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}' -X 'main.GitCommit=${GIT_COMMIT}'"

# Default to current platform if not specified
GOOS=${GOOS:-$(go env GOOS)}
GOARCH=${GOARCH:-$(go env GOARCH)}

# Output binary name
OUTPUT="figaro"
if [ "$GOOS" = "windows" ]; then
    OUTPUT="figaro.exe"
fi

echo "📋 Build Information:"
echo "   Version: ${VERSION}"
echo "   Platform: ${GOOS}/${GOARCH}"
echo "   Build Time: ${BUILD_TIME}"
echo "   Git Commit: ${GIT_COMMIT}"
echo ""

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -f figaro figaro.exe figaro-*

# Download dependencies
echo "📦 Downloading dependencies..."
go mod download

# Run tests
echo "🧪 Running tests..."
go test ./...

# Build the application
echo "🔨 Building binary..."

# Windows-specific CGO setup
if [ "$GOOS" = "windows" ]; then
    echo "🪟 Configuring Windows CGO build..."
    export CC=gcc
    export CGO_ENABLED=1
    
    # Check if GCC is available
    if ! command -v gcc &> /dev/null; then
        echo "❌ Error: GCC not found. Please install TDM-GCC or MinGW-w64:"
        echo "   - Using Chocolatey: choco install mingw"
        echo "   - Using winget: winget install mingw-w64" 
        echo "   - Or download from: https://jmeubank.github.io/tdm-gcc/"
        exit 1
    fi
    
    CGO_ENABLED=1 CC=gcc GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags="${LDFLAGS}" -o ${OUTPUT} ./cmd/figaro
else
    CGO_ENABLED=1 GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags="${LDFLAGS}" -o ${OUTPUT} ./cmd/figaro
fi

echo "✅ Build completed successfully!"
echo "   Binary: ${OUTPUT}"
echo "   Size: $(du -h ${OUTPUT} | cut -f1)"
echo ""
echo "🚀 To run the application:"
echo "   ./${OUTPUT}"
echo ""
echo "🐳 To build Docker image:"
echo "   docker build -t figaro:${VERSION} ."