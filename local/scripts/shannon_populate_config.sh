#!/usr/bin/env bash

# This script updates configuration values in a specified config file using sed.
# Using a variable for the config file path allows easy updates later.
CONFIG_FILE="./local/path/config/.config.yaml"

# Make a copy of the default config file
make config_shannon_localnet

# Check if gsed is installed on macOS
if [[ "$OSTYPE" == "darwin"* ]]; then
    if ! command -v gsed &>/dev/null; then
        echo "Error: gsed is not installed. Please run 'brew install gnu-sed' first."
        exit 1
    fi
    SED_CMD="gsed"
else
    SED_CMD="sed"
fi

# Replace endpoints as needed
$SED_CMD -i "s|rpc_url: \".*\"|rpc_url: $NODE|" "$CONFIG_FILE"
$SED_CMD -i "s|host_port: \".*\"|host_port: shannon-testnet-grove-grpc.beta.poktroll.com:443|" "$CONFIG_FILE"
$SED_CMD -i "s|gateway_address: .*|gateway_address: $(poktrolld keys show -a gateway)|" "$CONFIG_FILE"
$SED_CMD -i "s|gateway_private_key_hex: .*|gateway_private_key_hex: $(poktrolld keys export gateway --unsafe --unarmored-hex)|" "$CONFIG_FILE"
$SED_CMD -i '/owned_apps_private_keys_hex:/!b;n;c\      - '"$(poktrolld keys export application --unsafe --unarmored-hex)" "$CONFIG_FILE"
