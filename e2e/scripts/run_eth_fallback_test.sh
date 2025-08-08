#!/usr/bin/env bash
set -e
set -o nounset

# ETH Fallback Testing Script
# 
# DESCRIPTION:
# This script configures and tests ETH fallback functionality by:
# 1. Updating Shannon config to use a specific fallback URL
# 2. Running E2E tests with fallback enabled
# 3. Restoring the original configuration after testing
#
# USAGE:
#   ./update_shannon_config_fallback_eth.sh <FALLBACK_URL>
#
# EXAMPLE:
#   ./update_shannon_config_fallback_eth.sh "https://eth.rpc.backup.io"
#
# REQUIREMENTS:
# • yq tool for YAML manipulation
# • make command available for running tests
# • Valid Shannon config file at ./e2e/config/.shannon.config.yaml

# Colors for better log readability
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  INFO:${NC} $1"
}

log_success() {
    echo -e "${GREEN}✅ SUCCESS:${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠️  WARNING:${NC} $1"
}

log_error() {
    echo -e "${RED}❌ ERROR:${NC} $1"
}

# Usage function
usage() {
    echo "Usage: $0 <FALLBACK_URL>"
    echo ""
    echo "Configure ETH fallback settings and run E2E tests"
    echo ""
    echo "Arguments:"
    echo "  FALLBACK_URL    The URL to use for ETH fallback endpoint"
    echo ""
    echo "Example:"
    echo "  $0 'https://eth.rpc.backup.io'"
    exit 1
}

# Check if fallback URL argument is provided
if [[ $# -ne 1 ]]; then
    log_error "Missing required argument: FALLBACK_URL"
    usage
fi

FALLBACK_URL="$1"

# Validate the fallback URL format
if [[ ! "$FALLBACK_URL" =~ ^https?:// ]]; then
    log_error "Invalid URL format. Must start with http:// or https://"
    exit 1
fi

log_info "Starting ETH fallback configuration and testing"
log_info "Fallback URL: $FALLBACK_URL"

# Set the current working directory to e2e/config directory
SCRIPT_DIR="$(dirname "$0")"
CONFIG_DIR="$(realpath "$SCRIPT_DIR/../config")"
PROJECT_ROOT="$(realpath "$SCRIPT_DIR/../../")"

log_info "Changing to config directory: $CONFIG_DIR"
cd "$CONFIG_DIR" || {
    log_error "Failed to change to config directory: $CONFIG_DIR"
    exit 1
}

CONFIG_FILE="./.shannon.config.yaml"

# Check if config file exists
if [[ ! -f $CONFIG_FILE ]]; then
    log_error "Config file $CONFIG_FILE not found in $PWD"
    exit 1
fi

log_info "Found config file: $CONFIG_FILE"

# Create backup of original config with absolute path
BACKUP_FILE="$CONFIG_DIR/.shannon.config.yaml.backup"
log_info "Creating backup of original config: $BACKUP_FILE"
cp "$CONFIG_FILE" "$BACKUP_FILE"

# Function to restore original config
restore_config() {
    log_info "Restoring original configuration from backup"
    if [[ -f "$BACKUP_FILE" ]]; then
        # Ensure we're in the config directory for restoration
        cd "$CONFIG_DIR" 2>/dev/null || true
        cp "$BACKUP_FILE" "$CONFIG_FILE"
        rm "$BACKUP_FILE"
        log_success "Configuration restored successfully"
    else
        log_warning "Backup file not found at: $BACKUP_FILE"
        # Try to find backup file in config directory as fallback
        if [[ -f "$CONFIG_DIR/.shannon.config.yaml.backup" ]]; then
            log_info "Found backup in config directory, restoring..."
            cd "$CONFIG_DIR"
            cp ".shannon.config.yaml.backup" ".shannon.config.yaml"
            rm ".shannon.config.yaml.backup"
            log_success "Configuration restored from config directory backup"
        else
            log_error "Unable to restore configuration - no backup found"
        fi
    fi
}

# Set up trap to restore config on script exit (success or failure)
# Handle EXIT, SIGINT (Ctrl+C), SIGTERM, and other signals
trap restore_config EXIT SIGINT SIGTERM

# Configure ETH service fallback settings
log_info "Configuring ETH fallback settings..."
log_info "Setting send_all_traffic=true and fallback URL=$FALLBACK_URL"

yq -i "
    (.shannon_config.gateway_config.service_fallback[] | select(.service_id == \"eth\")).send_all_traffic = true |
    (.shannon_config.gateway_config.service_fallback[] | select(.service_id == \"eth\")).fallback_endpoints[0].default_url = \"$FALLBACK_URL\"
" "$CONFIG_FILE"

if [[ $? -eq 0 ]]; then
    log_success "ETH fallback configuration updated successfully"
else
    log_error "Failed to update ETH fallback configuration"
    exit 1
fi

# Verify the configuration was applied correctly
log_info "Verifying configuration changes..."
CURRENT_URL=$(yq '.shannon_config.gateway_config.service_fallback[] | select(.service_id == "eth").fallback_endpoints[0].default_url' "$CONFIG_FILE")
SEND_ALL_TRAFFIC=$(yq '.shannon_config.gateway_config.service_fallback[] | select(.service_id == "eth").send_all_traffic' "$CONFIG_FILE")

log_info "Current fallback URL: $CURRENT_URL"
log_info "Send all traffic: $SEND_ALL_TRAFFIC"

if [[ "$CURRENT_URL" != "$FALLBACK_URL" ]]; then
    log_error "Configuration verification failed: URL mismatch"
    exit 1
fi

if [[ "$SEND_ALL_TRAFFIC" != "true" ]]; then
    log_error "Configuration verification failed: send_all_traffic not set to true"
    exit 1
fi

log_success "Configuration verification passed"

# Change to project root to run make command
log_info "Changing to project root directory: $PROJECT_ROOT"
cd "$PROJECT_ROOT" || {
    log_error "Failed to change to project root directory: $PROJECT_ROOT"
    exit 1
}

# Run E2E tests with ETH fallback
log_info "Running E2E tests for ETH service with fallback enabled..."
log_info "Executing: make e2e_test eth"

if make e2e_test eth; then
    log_success "E2E tests completed successfully"
else
    log_error "E2E tests failed"
    exit 1
fi

log_success "ETH fallback testing completed successfully"
log_info "Original configuration will be restored automatically"
