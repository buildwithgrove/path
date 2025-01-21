#!/usr/bin/env bash

# This script updates configuration values in a specified config file using sed.
# It includes checks for required commands, address existence, and supports custom flag and address name overrides.

CONFIG_FILE="./local/path/config/.config.yaml"

# Wrapper function for poktrolld with overridden flags
pkd() {
    poktrolld --keyring-backend="${POKTROLL_TEST_KEYRING_BACKEND:-test}" --home="${POKTROLL_HOME_PROD:-/Users/$(whoami)/.poktroll}" "$@"
}

# Function to check if a command exists on the system
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if a specific address exists
address_exists() {
    local address
    address=$(pkd keys show -a "$1" 2>/dev/null)
    if [[ -z "$address" ]]; then
        return 1 # Address does not exist
    else
        echo "$address" # Address exists, output it
        return 0
    fi
}

# Ensure gsed is installed on macOS
if [[ "$OSTYPE" == "darwin"* ]]; then
    if ! command_exists gsed; then
        echo "Error: gsed is not installed. Please run 'brew install gnu-sed' first."
        exit 1
    fi
    SED_CMD="gsed"
else
    SED_CMD="sed"
fi

# Overridable names for gateway and application
GATEWAY_NAME="${POKTROLL_GATEWAY_NAME:-gateway}"
APPLICATION_NAME="${POKTROLL_APPLICATION_NAME:-application}"

# Check if the 'gateway_address' variable is empty (address does not exist)
gateway_address=$(address_exists "$GATEWAY_NAME")
if [[ -z "$gateway_address" ]]; then
    echo "Error: Gateway address '$GATEWAY_NAME' does not exist."
    exit 1
fi

# Check if the 'application_address' variable is empty (address does not exist)
application_address=$(address_exists "$APPLICATION_NAME")
if [[ -z "$application_address" ]]; then
    echo "Error: Application address '$APPLICATION_NAME' does not exist."
    exit 1
fi

# Export private keys for 'gateway' and 'application'
gateway_private_key_hex=$(pkd keys export ${GATEWAY_NAME} --unsafe --unarmored-hex)
application_private_key_hex=$(pkd keys export ${APPLICATION_NAME} --unsafe --unarmored-hex)

# Update the configuration file
make copy_shannon_e2e_config_to_local

# Replace configuration values
$SED_CMD -i "s|rpc_url: \".*\"|rpc_url: $NODE|" "$CONFIG_FILE"
$SED_CMD -i "s|host_port: \".*\"|host_port: shannon-testnet-grove-grpc.beta.poktroll.com:443|" "$CONFIG_FILE"
$SED_CMD -i "s|gateway_address: .*|gateway_address: $gateway_address|" "$CONFIG_FILE"
$SED_CMD -i "s|gateway_private_key_hex: .*|gateway_private_key_hex: $gateway_private_key_hex|" "$CONFIG_FILE"
$SED_CMD -i '/owned_apps_private_keys_hex:/!b;n;c\      - '"$application_private_key_hex" "$CONFIG_FILE"

echo "Configuration update completed."
