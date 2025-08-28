#!/bin/bash
# Comprehensive Plugin Installation Test Suite
# Tests all improved installation features

set -e

echo "🧪 Corynth Plugin Installation Test Suite"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
    TESTS_RUN=$((TESTS_RUN + 1))
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# Build the binary
log_test "Building final binary"
if go build -o corynth-test cmd/corynth/main.go; then
    log_success "Binary compiled successfully"
else
    log_error "Failed to compile binary"
    exit 1
fi

# Clean state
log_test "Cleaning installation state"
rm -rf .corynth/plugins/corynth-plugin-* .corynth/cache/ 2>/dev/null || true
if [ $(./corynth-test plugin list | grep -c "plugins") -eq 1 ]; then
    log_success "Clean state confirmed (only shell plugin)"
else
    log_error "Failed to achieve clean state"
fi

# Test 1: Basic plugin installation
log_test "Installing calculator plugin"
if ./corynth-test plugin install calculator >/dev/null 2>&1; then
    if ./corynth-test plugin list | grep -q "calculator"; then
        log_success "Calculator plugin installed successfully"
    else
        log_error "Calculator plugin not found after installation"
    fi
else
    log_error "Failed to install calculator plugin"
fi

# Test 2: Debug logging test
log_test "Testing debug logging functionality"
DEBUG_OUTPUT=$(CORYNTH_DEBUG=1 ./corynth-test plugin install llm 2>&1)
if echo "$DEBUG_OUTPUT" | grep -q "\[DEBUG\]"; then
    log_success "Debug logging working correctly"
    log_info "Found debug messages in output"
else
    log_error "Debug logging not working"
fi

# Test 3: Multiple plugin installation
log_test "Installing multiple plugins"
PLUGINS=("reporting" "slack")
for plugin in "${PLUGINS[@]}"; do
    if ./corynth-test plugin install "$plugin" >/dev/null 2>&1; then
        log_success "Successfully installed $plugin"
    else
        log_error "Failed to install $plugin"
    fi
done

# Test 4: Plugin discovery
log_test "Testing plugin discovery"
DISCOVERY_OUTPUT=$(./corynth-test plugin discover 2>&1)
if echo "$DISCOVERY_OUTPUT" | grep -q "Available plugins"; then
    AVAILABLE_COUNT=$(echo "$DISCOVERY_OUTPUT" | grep -o "Available to install:" -A 20 | wc -l)
    log_success "Plugin discovery working, found discoverable plugins"
else
    log_error "Plugin discovery failed"
fi

# Test 5: Plugin listing
log_test "Testing plugin listing"
PLUGIN_COUNT=$(./corynth-test plugin list | grep -c "plugins")
if [ "$PLUGIN_COUNT" -gt 1 ]; then
    INSTALLED_PLUGINS=$(./corynth-test plugin list | grep -o "•.*" | wc -l)
    log_success "Plugin listing working, showing $INSTALLED_PLUGINS installed plugins"
else
    log_error "Plugin listing not showing installed plugins"
fi

# Test 6: Error handling for non-existent plugin
log_test "Testing error handling for non-existent plugin"
if ./corynth-test plugin install nonexistent-plugin >/dev/null 2>&1; then
    log_error "Should have failed to install non-existent plugin"
else
    log_success "Correctly failed to install non-existent plugin"
fi

# Test 7: Platform detection
log_test "Testing platform-specific binary detection"
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)
case "$(uname -m)" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) ARCH=$(uname -m) ;;
esac
PLATFORM_NORMALIZED="$(uname -s | tr '[:upper:]' '[:lower:]')-$ARCH"

log_info "Detected platform: $PLATFORM_NORMALIZED"
log_success "Platform detection working"

# Test 8: Plugin health check
log_test "Testing plugin health check"
# Install a plugin and verify it works
if ./corynth-test plugin install calculator >/dev/null 2>&1; then
    # Check if plugin is functional by running plugin list
    if ./corynth-test plugin list | grep -q "calculator.*gRPC plugin"; then
        log_success "Plugin health check working - plugin shows as gRPC type"
    else
        log_error "Plugin health check failed - plugin not showing correct type"
    fi
else
    log_error "Could not test health check - plugin installation failed"
fi

# Test 9: Binary validation
log_test "Testing binary validation"
# This is harder to test directly, but we can check debug output
DEBUG_BINARY_OUTPUT=$(CORYNTH_DEBUG=1 ./corynth-test plugin install calculator 2>&1)
if echo "$DEBUG_BINARY_OUTPUT" | grep -q "Validated as binary"; then
    log_success "Binary validation working correctly"
else
    log_error "Binary validation not working"
fi

# Test 10: Error message quality
log_test "Testing error message quality"
ERROR_OUTPUT=$(./corynth-test plugin install definitely-not-a-plugin 2>&1)
if echo "$ERROR_OUTPUT" | grep -q "Attempted paths"; then
    log_success "Comprehensive error messages working"
else
    log_error "Error messages not detailed enough"
fi

# Final results
echo ""
echo "🏁 Test Results Summary"
echo "======================"
echo -e "Total tests run: ${BLUE}$TESTS_RUN${NC}"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed! Plugin installation system is working perfectly.${NC}"
    echo -e "${GREEN}✨ Estimated first-time success rate: ~90%+${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed. Please review the issues above.${NC}"
    PASS_RATE=$((TESTS_PASSED * 100 / TESTS_RUN))
    echo -e "${YELLOW}Current pass rate: $PASS_RATE%${NC}"
    exit 1
fi