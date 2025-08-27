# Corynth Makefile
# Production-ready build system

# Variables
BINARY_NAME=corynth
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build paths
CMD_DIR=cmd/corynth
PKG_DIR=pkg
PLUGIN_SRC_DIR=plugins/src
PLUGIN_DIST_DIR=plugins/dist

# Build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.Commit=${COMMIT}"

# Targets
.PHONY: all build clean test plugins install help

## Default target
all: clean build plugins

## Build the main binary
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./$(CMD_DIR)
	@echo "✅ Build complete: $(BINARY_NAME)"

## Build all plugins
plugins:
	@echo "Building plugins..."
	@mkdir -p $(PLUGIN_DIST_DIR)
	@for plugin in $(PLUGIN_SRC_DIR)/*; do \
		name=$$(basename $$plugin); \
		echo "  Building $$name..."; \
		$(GOBUILD) -buildmode=plugin -o $(PLUGIN_DIST_DIR)/$$name.so $$plugin/plugin.go || exit 1; \
	done
	@echo "✅ Plugins built successfully"

## Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v ./pkg/...
	@echo "✅ Tests passed"

## Run integration tests
test-integration:
	@echo "Running integration tests..."
	@$(GOTEST) -v ./tests/integration/...
	@echo "✅ Integration tests passed"

## Install to system
install: build
	@echo "Installing $(BINARY_NAME)..."
	@sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "✅ Installed to /usr/local/bin/$(BINARY_NAME)"

## Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -f $(PLUGIN_DIST_DIR)/*.so
	@echo "✅ Clean complete"

## Update dependencies
deps:
	@echo "Updating dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "✅ Dependencies updated"

## Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w cmd pkg
	@gofmt -s -w $(PLUGIN_SRC_DIR)
	@echo "✅ Code formatted"

## Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run ./cmd/... ./pkg/...
	@echo "✅ Linting complete"

## Generate documentation
docs:
	@echo "Generating documentation..."
	@godoc -http=:6060 &
	@echo "✅ Documentation server running at http://localhost:6060"

## Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "✅ Multi-platform build complete"

## Create release packages
release: build-all
	@echo "Creating release packages..."
	@cd dist && for file in $(BINARY_NAME)-*; do \
		tar czf $$file.tar.gz $$file; \
		rm $$file; \
	done
	@echo "✅ Release packages created in dist/"

## Run development server with hot reload
dev:
	@echo "Starting development server..."
	@air -c .air.toml

## Display help
help:
	@echo "Corynth Build System"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed 's/## /  /'
	@echo ""
	@echo "Examples:"
	@echo "  make              # Build everything"
	@echo "  make build        # Build main binary only"
	@echo "  make plugins      # Build plugins only"
	@echo "  make test         # Run tests"
	@echo "  make install      # Install to system"
	@echo "  make clean        # Clean build artifacts"