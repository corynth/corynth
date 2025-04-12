#!/bin/bash

# Test script for Kubernetes plugin
# This script demonstrates how to use the Kubernetes plugin with Corynth

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Testing Kubernetes plugin for Corynth...${NC}"

# Check if corynth is installed
if ! command -v corynth &> /dev/null; then
    echo -e "${RED}Corynth is not installed. Please install Corynth before continuing.${NC}"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}kubectl is not installed. Please install kubectl before continuing.${NC}"
    exit 1
fi

# Check if kubectl is configured
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}kubectl is not configured or cannot connect to a cluster.${NC}"
    echo -e "${RED}Please configure kubectl to connect to a Kubernetes cluster before continuing.${NC}"
    exit 1
fi

# Create a test directory
TEST_DIR="$(pwd)/test_kubernetes"
echo -e "${YELLOW}Creating test directory at ${TEST_DIR}...${NC}"
mkdir -p "${TEST_DIR}"

# Initialize a new Corynth project
echo -e "${YELLOW}Initializing Corynth project...${NC}"
corynth init "${TEST_DIR}"

# Copy the kubernetes_flow.yaml to the project
echo -e "${YELLOW}Copying kubernetes_flow.yaml to the project...${NC}"
cp "$(pwd)/../../examples/kubernetes_flow.yaml" "${TEST_DIR}/flows/"

# Copy the plugins.yaml to the project
echo -e "${YELLOW}Copying plugins.yaml to the project...${NC}"
cp "$(pwd)/../../examples/plugins.yaml" "${TEST_DIR}/plugins.yaml"

# Create a local plugins directory
echo -e "${YELLOW}Creating local plugins directory...${NC}"
mkdir -p "${TEST_DIR}/plugins/kubernetes"

# Copy the plugin to the local plugins directory
echo -e "${YELLOW}Copying plugin to the local plugins directory...${NC}"
if [ -f "$(pwd)/build/kubernetes.so" ]; then
    cp "$(pwd)/build/kubernetes.so" "${TEST_DIR}/plugins/kubernetes/"
else
    echo -e "${RED}Plugin not found. Please build the plugin first using build.sh.${NC}"
    exit 1
fi

# Change to the test directory
cd "${TEST_DIR}"

# Run the plan command
echo -e "${YELLOW}Running corynth plan...${NC}"
corynth plan

# Ask for confirmation before applying
echo -e "${YELLOW}Do you want to apply the plan? This will create resources in your Kubernetes cluster. (y/n)${NC}"
read -r CONFIRM
if [[ "${CONFIRM}" != "y" ]]; then
    echo -e "${YELLOW}Skipping apply.${NC}"
    echo -e "${GREEN}Test completed. You can examine the plan output above.${NC}"
    exit 0
fi

# Run the apply command
echo -e "${YELLOW}Running corynth apply...${NC}"
corynth apply

echo -e "${GREEN}Test completed successfully!${NC}"
echo -e "${YELLOW}The test created resources in your Kubernetes cluster.${NC}"
echo -e "${YELLOW}You can clean up these resources by running:${NC}"
echo -e "${YELLOW}kubectl delete namespace corynth-demo${NC}"

# Return to the original directory
cd - > /dev/null