# Makefile for Corynth tests

# Variables
GO = go
GOFLAGS = -v
TEST_DIRS = ./unit ./integration ./e2e
COVERAGE_DIR = ./coverage
COVERAGE_FILE = $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML = $(COVERAGE_DIR)/coverage.html

# Default target
.PHONY: all
all: test

# Create coverage directory
.PHONY: create_coverage_dir
create_coverage_dir:
	mkdir -p $(COVERAGE_DIR)

# Run all tests
.PHONY: test
test:
	$(GO) test $(GOFLAGS) $(TEST_DIRS)

# Run unit tests
.PHONY: unit
unit:
	$(GO) test $(GOFLAGS) ./unit

# Run integration tests
.PHONY: integration
integration:
	$(GO) test $(GOFLAGS) ./integration

# Run end-to-end tests
.PHONY: e2e
e2e:
	$(GO) test $(GOFLAGS) ./e2e

# Run tests with coverage
.PHONY: coverage
coverage: create_coverage_dir
	$(GO) test $(GOFLAGS) -coverprofile=$(COVERAGE_FILE) ./unit ./integration
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated at $(COVERAGE_HTML)"

# Run tests with race detection
.PHONY: race
race:
	$(GO) test $(GOFLAGS) -race $(TEST_DIRS)

# Run tests with verbose output
.PHONY: verbose
verbose:
	$(GO) test $(GOFLAGS) -v $(TEST_DIRS)

# Clean test artifacts
.PHONY: clean
clean:
	rm -rf $(COVERAGE_DIR)

# Help target
.PHONY: help
help:
	@echo "Corynth Tests Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  all       - Run all tests (default)"
	@echo "  unit      - Run unit tests"
	@echo "  integration - Run integration tests"
	@echo "  e2e       - Run end-to-end tests"
	@echo "  coverage  - Run tests with coverage"
	@echo "  race      - Run tests with race detection"
	@echo "  verbose   - Run tests with verbose output"
	@echo "  clean     - Clean test artifacts"
	@echo "  help      - Show this help message"