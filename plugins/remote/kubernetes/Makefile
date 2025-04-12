# Makefile for Kubernetes plugin

# Variables
PLUGIN_NAME = kubernetes
PLUGIN_DIR = $(shell pwd)
BUILD_DIR = $(PLUGIN_DIR)/build
CORYNTH_ROOT = $(shell cd ../../.. && pwd)

# Go build flags
GO = go
GOFLAGS = -buildmode=plugin -trimpath

# Platform-specific settings
LINUX_ARCH = amd64
LINUX_OS = linux
DARWIN_ARCH = arm64
DARWIN_OS = darwin

# Default target
.PHONY: all
all: linux darwin

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build for Linux (for Mac M4)
.PHONY: linux
linux: $(BUILD_DIR)
	CGO_ENABLED=1 \
	GOOS=$(LINUX_OS) \
	GOARCH=$(LINUX_ARCH) \
	CC=x86_64-linux-gnu-gcc \
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME).so .

# Build for macOS
.PHONY: darwin
darwin: $(BUILD_DIR)
	CGO_ENABLED=1 \
	GOOS=$(DARWIN_OS) \
	GOARCH=$(DARWIN_ARCH) \
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME)_darwin.so .

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

# Install the plugin to the Corynth plugins directory
.PHONY: install
install: linux
	mkdir -p $(CORYNTH_ROOT)/plugins/remote/$(PLUGIN_NAME)
	cp $(BUILD_DIR)/$(PLUGIN_NAME).so $(CORYNTH_ROOT)/plugins/remote/$(PLUGIN_NAME)/

# Package the plugin for distribution
.PHONY: package
package: linux darwin
	mkdir -p $(BUILD_DIR)/package
	cp $(BUILD_DIR)/$(PLUGIN_NAME).so $(BUILD_DIR)/package/
	cp $(BUILD_DIR)/$(PLUGIN_NAME)_darwin.so $(BUILD_DIR)/package/
	cp README.md $(BUILD_DIR)/package/ || echo "No README.md found"
	cd $(BUILD_DIR) && tar -czf $(PLUGIN_NAME).tar.gz -C package .

# Help target
.PHONY: help
help:
	@echo "Kubernetes Plugin Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  all      - Build for all platforms (default)"
	@echo "  linux    - Build for Linux (for Mac M4)"
	@echo "  darwin   - Build for macOS"
	@echo "  clean    - Remove build artifacts"
	@echo "  install  - Install the plugin to the Corynth plugins directory"
	@echo "  package  - Package the plugin for distribution"
	@echo "  help     - Show this help message"