#!/bin/bash

# Setup script for the remote Kubernetes plugin example
# This script sets up a new Corynth project with the remote Kubernetes plugin

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up remote Kubernetes plugin example...${NC}"

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

# Create a new directory for the example
EXAMPLE_DIR="remote-k8s-example"
echo -e "${YELLOW}Creating example directory at ${EXAMPLE_DIR}...${NC}"
mkdir -p "${EXAMPLE_DIR}"

# Initialize a new Corynth project
echo -e "${YELLOW}Initializing Corynth project...${NC}"
corynth init "${EXAMPLE_DIR}"

# Copy the plugins.yaml file to the project
echo -e "${YELLOW}Copying plugins.yaml to the project...${NC}"
cp plugins.yaml "${EXAMPLE_DIR}/"

# Create the flows directory if it doesn't exist
mkdir -p "${EXAMPLE_DIR}/flows"

# Copy the kubernetes_flow.yaml to the project
echo -e "${YELLOW}Copying kubernetes_flow.yaml to the project...${NC}"
cp kubernetes_flow.yaml "${EXAMPLE_DIR}/flows/"

# Change to the example directory
cd "${EXAMPLE_DIR}"

# Run the plan command
echo -e "${YELLOW}Running corynth plan...${NC}"
echo -e "${YELLOW}This will download the Kubernetes plugin automatically.${NC}"
corynth plan

# Ask for confirmation before applying
echo -e "${YELLOW}Do you want to apply the plan? This will create resources in your Kubernetes cluster. (y/n)${NC}"
read -r CONFIRM
if [[ "${CONFIRM}" != "y" ]]; then
    echo -e "${YELLOW}Skipping apply.${NC}"
    echo -e "${GREEN}Setup completed. You can examine the plan output above.${NC}"
    exit 0
fi

# Run the apply command
echo -e "${YELLOW}Running corynth apply...${NC}"
corynth apply

echo -e "${GREEN}Example completed successfully!${NC}"
echo -e "${YELLOW}The example created resources in your Kubernetes cluster and then cleaned them up.${NC}"
echo -e "${YELLOW}You can run the example again by running:${NC}"
echo -e "${YELLOW}cd ${EXAMPLE_DIR} && corynth apply${NC}"

# Return to the original directory
cd - > /dev/null