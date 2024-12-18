#!/usr/bin/env bash

# This script updates configuration values in a specified config file using sed.
# Using a variable for the config file path allows easy updates later.
CONFIG_FILE="./local/path/config/.config.yaml"

# Make a copy of the default config file
make config_shannon_localnet

# Replace endpoints as needed
sed -i "s|rpc_url: \".*\"|rpc_url: $NODE|" "$CONFIG_FILE"
sed -i "s|host_port: \".*\"|host_port: shannon-testnet-grove-grpc.beta.poktroll.com:443|" "$CONFIG_FILE"

# Update gateway and application addresses
sed -i "s|gateway_address: .*|gateway_address: $GATEWAY_ADDR|" "$CONFIG_FILE"
sed -i "s|gateway_private_key_hex: .*|gateway_private_key_hex: $(poktrolld keys export gateway --unsafe --unarmored-hex)|" "$CONFIG_FILE"
sed -i '/owned_apps_private_keys_hex:/!b;n;c\      - '"$(poktrolld keys export application --unsafe --unarmored-hex)" "$CONFIG_FILE"
