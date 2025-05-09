#!/usr/bin/env bash

# This script updates configuration values in a specified config file using sed.
# It includes checks for required commands, address existence, and supports custom flag and address name overrides.

CONFIG_FILE="./local/path/.config.yaml"

GREEN='\033[1;32m'
BLUE='\033[1;34m'
YELLOW='\033[1;33m'
RED='\033[1;31m'
NC='\033[0m'

echo -e "\n"
echo -e "${GREEN}🌿  This script will populate the configuration file with the correct values  🌿${NC}"
echo -e ""
echo -e "   Ensure you have completed the ${BLUE}App & PATH Gateway Cheat Sheet${NC} before running this script."
echo -e "   ${BLUE}https://dev.poktroll.com/operate/cheat_sheets/gateway_cheatsheet${NC} (⏰ ~30 min to complete)"
echo -e "   Do you wish to continue? (y/n)${NC}"
echo -n "> "
read -r response
if [[ "$response" != "yes" && "$response" != "y" ]]; then
    echo -e ""
    echo -e "${RED}❌  Operation cancelled  ❌${NC}"
    echo -e ""
    exit 0
fi

# Check if the configuration file already exists
if [[ -f "$CONFIG_FILE" ]]; then
    echo -e "${YELLOW}❕  Warning: The configuration file already exists. Do you want to overwrite it? (y/n)${NC}"
    echo -n "> "
    read -r response
    if [[ "$response" != "yes" && "$response" != "y" ]]; then
        echo "Operation cancelled."
        exit 0
    fi
    echo -e "${RED}❗ Are you sure you want to overwrite the existing configuration file? (y/n)${NC}"
    echo -n "> "
    read -r response
    if [[ "$response" != "yes" && "$response" != "y" ]]; then
        echo "Operation cancelled."
        exit 0
    fi
fi

# Wrapper function for pocketd with overridden flags
pocketd_with_env() {
    pocketd --keyring-backend="${POCKET_TEST_KEYRING_BACKEND:-test}" --home="${POCKET_HOME_PROD:-${HOME}/.poktroll}" "$@"
}

# Function to check if a command exists on the system
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if a specific address exists
address_exists() {
    local address
    address=$(pocketd_with_env keys show -a "$1" 2>/dev/null)
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
        echo -e ""
        echo -e "${RED}❌  gsed is not installed. Run 'brew install gnu-sed' first  ❌${NC}"
        echo -e ""
        exit 1
    fi
    SED_CMD="gsed"
else
    SED_CMD="sed"
fi

# Overridable names for gateway and application
GATEWAY_NAME="${POCKET_GATEWAY_NAME:-gateway}"
APPLICATION_NAME="${POCKET_APPLICATION_NAME:-application}"

# Check if the 'gateway_address' variable is empty (address does not exist)
gateway_address=$(address_exists "$GATEWAY_NAME")
if [[ -z "$gateway_address" ]]; then
    echo -e "${POCKET_TEST_KEYRING_BACKEND}"
    echo -e "${RED}❌  Gateway address '$GATEWAY_NAME' does not exist  ❌${NC}"
    echo -e ""
    echo -e "💡  Refer to the ${BLUE}App & PATH Gateway Cheat Sheet${NC} for instructions on how to create a Gateway address."
    echo -e "   ${BLUE}https://dev.poktroll.com/operate/cheat_sheets/gateway_cheatsheet${NC}"
    exit 1
fi

# Check if the 'application_address' variable is empty (address does not exist)
application_address=$(address_exists "$APPLICATION_NAME")
if [[ -z "$application_address" ]]; then
    echo -e ""
    echo -e "${RED}❌  Application address '$APPLICATION_NAME' does not exist  ❌${NC}"
    echo -e ""
    echo -e "💡  Refer to the ${BLUE}App & PATH Application Cheat Sheet${NC} for instructions on how to create an Application address."
    echo -e "   ${BLUE}https://dev.poktroll.com/operate/cheat_sheets/application_cheatsheet${NC}"
    exit 1
fi

# Export private keys for 'gateway' and 'application'
gateway_private_key_hex=$(pocketd_with_env keys export ${GATEWAY_NAME} --unsafe --unarmored-hex)
application_private_key_hex=$(pocketd_with_env keys export ${APPLICATION_NAME} --unsafe --unarmored-hex)

# Write new configuration file with updated values
cat >"$CONFIG_FILE" <<EOF
shannon_config:
    full_node_config:
        rpc_url: https://shannon-testnet-grove-rpc.beta.poktroll.com
        grpc_config:
            host_port: shannon-testnet-grove-grpc.beta.poktroll.com:443
        lazy_mode: true
    gateway_config:
        gateway_mode: "centralized"
        gateway_address: ${gateway_address}
        gateway_private_key_hex: ${gateway_private_key_hex}
        owned_apps_private_keys_hex:
            - ${application_private_key_hex}
EOF

echo -e "${GREEN}✅   Configuration update completed.${NC}"
echo -e ""
echo -e "You can view the new config file by running: cat local/path/.config.yaml"
