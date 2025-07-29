#!/bin/bash

# Script to parse service_qos_config.go and generate markdown tables of supported QoS services
# Usage: ./service_qos_doc_generator.sh <path/to/service_qos_config.go> <output_markdown_file>

if [ $# -lt 2 ]; then
    echo "Usage: $0 <path/to/service_qos_config.go> <output_markdown_file>"
    exit 1
fi

INPUT_FILE="$1"
OUTPUT_FILE="$2"

if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: Input file '$INPUT_FILE' not found."
    exit 1
fi

# Function to convert hex chain ID to decimal
hex_to_decimal() {
    local hex="$1"
    # Remove "0x" prefix if present
    hex="${hex#0x}"
    # Convert hex to decimal using printf
    printf "%d" "0x$hex" 2>/dev/null || echo "N/A"
}

# Function to generate the static content for the beginning of the markdown file
generate_static_content() {
    cat >"$1" <<'EOF'
---
sidebar_position: 1
title: Supported QoS Services
description: Supported Quality of Service Implementations in PATH
---

:::danger DO NOT EDIT

This file was auto-generated via `make gen_service_qos_docs`.

:::

## Configuring PATH QoS Checks

PATH uses an **opt-out** rather than an **opt-in** approach to QoS checks.

This means that PATH **automatically** performs QoS checks for all services the applications it manages are staked for.

### Disable QoS Checks for a particular Service

In order to disable QoS checks for a specific service, the `service_id` field may be specified in the `.config.yaml` file's `qos_disabled_service_ids` field.

For example, to disable QoS checks for the Ethereum service on a Shannon PATH instance, the following configuration would be added to the `.config.yaml` file:

```yaml
hydrator_config:
  qos_disabled_service_ids:
    - "eth"
```

See [PATH Configuration File](../../develop/configs/2_gateway_config.md#hydrator_config-optional) for more details.

## ⛓️ Supported QoS Services

The table below lists the Quality of Service (QoS) implementations currently supported by PATH.

:::warning **🚧 QoS Support 🚧**

If a Service ID is not specified in the tables below, it does not have a QoS implementation in PATH.

:::

EOF
}

# Extract default chain IDs from the config file
extract_default_values() {
    local file="$1"
    local default_evm_chain_id_int=""

    while IFS= read -r line; do
        # Match defaultEVMChainID
        if [[ "$line" =~ defaultEVMChainID[[:space:]]*=[[:space:]]*\"([^\"]+)\"[[:space:]]*//(.*) ]]; then
            default_evm_chain_id_hex="${BASH_REMATCH[1]}"
            # Extract decimal value from comment if available
            if [[ "${BASH_REMATCH[2]}" =~ \(([0-9]+)\) ]]; then
                default_evm_chain_id_int="${BASH_REMATCH[1]}"
            else
                default_evm_chain_id_int="$(hex_to_decimal "$default_evm_chain_id_hex")"
            fi
            break
        fi
    done <"$file"

    echo "$default_evm_chain_id_int"
}

# Process services in the specified section
process_services() {
    local section="$1"
    local default_chain_id="$2"
    local file="$3"
    local output_file="$4"
    local service_name=""
    local comment_buffer=""
    local in_section=false

    while IFS= read -r line; do
        # Check if we're in the specified section
        if [[ "$line" =~ ^var\ ${section}\ = ]]; then
            in_section=true
            continue
        fi

        # Check if we've reached the end of the section
        if [[ "$in_section" == true && "$line" =~ ^}$ ]]; then
            in_section=false
            break
        fi

        # Skip if not in the specified section
        if [[ "$in_section" != true ]]; then
            continue
        fi

        # Capture comments for service name extraction
        if [[ "$line" =~ ^[[:space:]]*//[[:space:]]*(.*)[[:space:]]*$ ]]; then
            comment_text="${BASH_REMATCH[1]}"

            # Skip section headers
            if [[ "$comment_text" =~ ^\*\*\*.*\*\*\*$ || "$comment_text" =~ ^=+$ ]]; then
                comment_buffer=""
                continue
            fi

            # Extract just the service name (part before " - " if present)
            if [[ "$comment_text" == *" - https"* ]]; then
                # Extract service name before the " - " and trim whitespace
                comment_buffer="$(echo "$comment_text" | sed 's/ - https.*//' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
            else
                # No URL pattern found, use the whole comment
                comment_buffer="$comment_text"
            fi
            continue
        fi

        # Process EVM configurations with inline chain ID
        if [[ "$line" =~ evm\.NewEVMServiceQoSConfig\([[:space:]]*\"([^\"]+)\",[[:space:]]*\"([^\"]+)\" ]]; then
            service_id="${BASH_REMATCH[1]}"
            chain_id_hex="${BASH_REMATCH[2]}"
            service_type="EVM"

            # Convert chain ID to decimal
            chain_id="$(hex_to_decimal "$chain_id_hex")"

            # Check if this is an archival service
            archival_check=""
            if [[ "$line" =~ evm\.NewEVMArchivalCheckConfig ]]; then
                archival_check="✅"
            fi

            # Use the most recent comment as the service name
            service_name="${comment_buffer:-Unknown EVM Service}"
            comment_buffer=""

            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >>"$output_file"

        # Process EVM configurations with defaultEVMChainID (Ethereum)
        elif [[ "$line" =~ evm\.NewEVMServiceQoSConfig\([[:space:]]*\"([^\"]+)\",[[:space:]]*defaultEVMChainID ]]; then
            service_id="${BASH_REMATCH[1]}"
            chain_id="$default_chain_id"
            service_type="EVM"

            # Ethereum is always archival
            archival_check="✅"

            # Use the most recent comment as the service name
            service_name="${comment_buffer:-Ethereum}"
            comment_buffer=""

            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >>"$output_file"

        # Process Cosmos SDK configurations
        elif [[ "$line" =~ cosmos\.NewCosmosSDKServiceQoSConfig\([[:space:]]*\"([^\"]+)\",[[:space:]]*\"([^\"]+)\" ]]; then
            service_id="${BASH_REMATCH[1]}"
            chain_id="${BASH_REMATCH[2]}"
            service_type="Cosmos SDK"
            archival_check=""

            # Use the most recent comment as the service name
            service_name="${comment_buffer:-Unknown Cosmos SDK Service}"
            comment_buffer=""

            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >>"$output_file"

        # Process Solana configurations
        elif [[ "$line" =~ solana\.NewSolanaServiceQoSConfig\([[:space:]]*\"([^\"]+)\" ]]; then
            service_id="${BASH_REMATCH[1]}"
            chain_id=""
            service_type="Solana"
            archival_check=""

            # Use the most recent comment as the service name
            service_name="${comment_buffer:-Solana}"
            comment_buffer=""

            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >>"$output_file"
        fi
    done <"$file"
}

# Main execution starts here
default_evm_chain_id=$(extract_default_values "$INPUT_FILE")

# Generate the static content at the beginning of the file
generate_static_content "$OUTPUT_FILE"

# Start the dynamic section
{
    echo "## 🌿 Current PATH QoS Support"
    echo ""
    echo "**🗓️ Document Last Updated: $(date '+%Y-%m-%d')**"
    echo ""

    # Add Shannon services section header
    echo "## Shannon Protocol Services"
    echo ""
    echo "| Service Name | Authoritative Service ID | Service QoS Type | Chain ID (if applicable) | Archival Check Configured |"
    echo "|-------------|------------|-----------------|----------|---------------------------|"
} >>"$OUTPUT_FILE"

# Process Shannon services
process_services "shannonServices" "$default_evm_chain_id" "$INPUT_FILE" "$OUTPUT_FILE"

echo "Documentation successfully updated at $OUTPUT_FILE"
