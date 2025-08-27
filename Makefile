# Corynth Makefile
# Production-ready build system with gRPC plugin support

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
PLUGIN_SRC_DIR=plugins-src
BUILD_DIR=bin
PLUGINS_DIR=$(BUILD_DIR)/plugins

# Build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.Commit=${COMMIT}"

# Targets
.PHONY: all build clean test plugins install help build-plugins build-k8s build-docker build-terraform test-plugins

## Default target
all: clean build plugins

## Build the main binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## Build all gRPC plugins
plugins: build-plugins

build-plugins: build-http build-k8s build-docker build-terraform
	@echo "✅ All plugins built successfully"

## Build individual plugins
build-http:
	@echo "Building http plugin..."
	@mkdir -p $(PLUGINS_DIR)
	@cd $(PLUGIN_SRC_DIR)/http && $(GOBUILD) -o ../../$(PLUGINS_DIR)/corynth-plugin-http main.go

build-k8s:
	@echo "Building k8s plugin..."
	@mkdir -p $(PLUGINS_DIR)
	@cd $(PLUGIN_SRC_DIR)/k8s && $(GOBUILD) -o ../../$(PLUGINS_DIR)/corynth-plugin-k8s main.go

build-docker:
	@echo "Building docker plugin..."
	@mkdir -p $(PLUGINS_DIR)
	@cd $(PLUGIN_SRC_DIR)/docker && $(GOBUILD) -o ../../$(PLUGINS_DIR)/corynth-plugin-docker main.go

build-terraform:
	@echo "Building terraform plugin..."
	@mkdir -p $(PLUGINS_DIR)
	@cd $(PLUGIN_SRC_DIR)/terraform && $(GOBUILD) -o ../../$(PLUGINS_DIR)/corynth-plugin-terraform main.go

## Run tests
test:
	@echo "Running core tests..."
	@$(GOTEST) -v ./pkg/...
	@echo "✅ Core tests passed"

## Run plugin tests
test-plugins: test-k8s test-docker test-terraform
	@echo "✅ All plugin tests passed"

test-k8s:
	@echo "Testing k8s plugin..."
	@cd $(PLUGIN_SRC_DIR)/k8s && $(GOMOD) verify && $(GOTEST) -v .

test-docker:
	@echo "Testing docker plugin..."
	@cd $(PLUGIN_SRC_DIR)/docker && $(GOMOD) verify && $(GOTEST) -v .

test-terraform:
	@echo "Testing terraform plugin..."
	@cd $(PLUGIN_SRC_DIR)/terraform && $(GOMOD) verify && $(GOTEST) -v .

## Test plugin servers (verify gRPC handshake)
test-plugin-servers: build-plugins
	@echo "Testing plugin servers..."
	@timeout 5s $(PLUGINS_DIR)/corynth-plugin-k8s serve > /dev/null 2>&1 && echo "✓ k8s plugin server" || echo "✗ k8s plugin server failed"
	@timeout 5s $(PLUGINS_DIR)/corynth-plugin-docker serve > /dev/null 2>&1 && echo "✓ docker plugin server" || echo "✗ docker plugin server failed"
	@timeout 5s $(PLUGINS_DIR)/corynth-plugin-terraform serve > /dev/null 2>&1 && echo "✓ terraform plugin server" || echo "✗ terraform plugin server failed"

## Install to system
install: build plugins
	@echo "Installing $(BINARY_NAME) and plugins..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@mkdir -p ~/.corynth/plugins
	@cp $(PLUGINS_DIR)/* ~/.corynth/plugins/
	@chmod +x ~/.corynth/plugins/*
	@echo "✅ Installed to /usr/local/bin/$(BINARY_NAME) and ~/.corynth/plugins/"

## Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@cd $(PLUGIN_SRC_DIR)/k8s && $(GOCLEAN)
	@cd $(PLUGIN_SRC_DIR)/docker && $(GOCLEAN)
	@cd $(PLUGIN_SRC_DIR)/terraform && $(GOCLEAN)
	@echo "✅ Clean complete"

## Update dependencies
deps:
	@echo "Updating dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@cd $(PLUGIN_SRC_DIR)/k8s && $(GOMOD) download && $(GOMOD) tidy
	@cd $(PLUGIN_SRC_DIR)/docker && $(GOMOD) download && $(GOMOD) tidy
	@cd $(PLUGIN_SRC_DIR)/terraform && $(GOMOD) download && $(GOMOD) tidy
	@echo "✅ All dependencies updated"

## Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w cmd pkg
	@cd $(PLUGIN_SRC_DIR)/k8s && gofmt -s -w .
	@cd $(PLUGIN_SRC_DIR)/docker && gofmt -s -w .
	@cd $(PLUGIN_SRC_DIR)/terraform && gofmt -s -w .
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