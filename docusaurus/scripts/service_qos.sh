#!/bin/bash

# Script to parse service_qos.go and generate markdown tables
# Usage: ./generate_service_docs.sh <path/to/service_qos.go> <output_markdown_file>

if [ $# -lt 2 ]; then
    echo "Usage: $0 <path/to/service_qos.go> <output_markdown_file>"
    exit 1
fi

INPUT_FILE="$1"
OUTPUT_FILE="$2"

if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: Input file '$INPUT_FILE' not found."
    exit 1
fi

# Function to convert hex to decimal
hex_to_decimal() {
    local hex="$1"
    # Remove "0x" prefix if present
    hex="${hex#0x}"
    # Convert hex to decimal using printf
    printf "%d" "0x$hex" 2>/dev/null || echo "N/A"
}

# Get the current content of the file and extract content up to the QoS section
if [ -f "$OUTPUT_FILE" ]; then
    # Find the line number where "# 🌿 Current PATH QoS Support" starts
    HEADER_LINE=$(grep -n "# 🌿 Current PATH QoS Support" "$OUTPUT_FILE" | head -1 | cut -d: -f1)
    
    if [ -n "$HEADER_LINE" ]; then
        # Save the content before "# 🌿 Current PATH QoS Support"
        head -n $((HEADER_LINE - 1)) "$OUTPUT_FILE" > "${OUTPUT_FILE}.tmp"
    else
        # If not found, preserve all content (we'll append to it)
        cat "$OUTPUT_FILE" > "${OUTPUT_FILE}.tmp"
    fi
else
    # Create an empty temp file if output file doesn't exist
    touch "${OUTPUT_FILE}.tmp"
fi

# First, extract default values
default_evm_chain_id_hex=""
default_evm_chain_id_int=""
default_cometbft_chain_id=""

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
    fi
    
    # Match defaultCometBFTChainID
    if [[ "$line" =~ defaultCometBFTChainID[[:space:]]*=[[:space:]]*\"([^\"]+)\" ]]; then
        default_cometbft_chain_id="${BASH_REMATCH[1]}"
    fi
done < "$INPUT_FILE"

# Start the new section
{
    echo "# 🌿 Current PATH QoS Support"
    echo ""
    echo "**🗓️ Document Last Updated: $(date '+%Y-%m-%d')**"
    echo ""
    
    # Process Shannon services
    echo "## Shannon Protocol Services"
    echo ""
    echo "| Service Name | Authoritative Service ID | Service QoS Type | Chain ID (if applicable) | Archival Check Configured |"
    echo "|-------------|------------|-----------------|----------|---------------------------|"
} > "${OUTPUT_FILE}.new"

# Parse Shannon services
in_shannon_section=false
current_service=""
service_id=""
service_name=""
service_type=""
chain_id=""
archival_check=""

while IFS= read -r line; do
    # Check if we're in the Shannon services section
    if [[ "$line" =~ ^var\ shannonServices\ = ]]; then
        in_shannon_section=true
        continue
    fi

    # Check if we've reached the end of Shannon services
    if [[ "$in_shannon_section" == true && "$line" =~ ^}$ ]]; then
        # Process the last service before exiting
        if [[ -n "$service_id" ]]; then
            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >> "${OUTPUT_FILE}.new"
        fi
        in_shannon_section=false
        continue
    fi

    # Skip if not in Shannon section
    if [[ "$in_shannon_section" != true ]]; then
        continue
    fi

    # Check for new service
    if [[ "$line" =~ ServiceConfig\{ ]]; then
        # Process the previous service if exists
        if [[ -n "$service_id" ]]; then
            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >> "${OUTPUT_FILE}.new"
        fi
        
        # Reset variables for new service
        service_id=""
        service_name=""
        chain_id=""
        archival_check=""
        
        # Determine service type
        if [[ "$line" =~ evm\.ServiceConfig ]]; then
            service_type="EVM"
        elif [[ "$line" =~ cometbft\.ServiceConfig ]]; then
            service_type="CometBFT"
        elif [[ "$line" =~ solana\.ServiceConfig ]]; then
            service_type="Solana"
        else
            service_type="Unknown"
        fi
    fi

    # Parse service ID and name
    if [[ "$line" =~ ServiceID:[[:space:]]*\"([^\"]+)\",[[:space:]]*//(.*) ]]; then
        service_id="${BASH_REMATCH[1]}"
        service_name="${BASH_REMATCH[2]}"
        # Remove leading/trailing spaces from service name
        service_name="$(echo "$service_name" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
    fi

    # Parse EVM Chain ID from explicit value
    if [[ "$line" =~ EVMChainID:[[:space:]]*\"([^\"]+)\",[[:space:]]*//(.*) ]]; then
        chain_id_hex="${BASH_REMATCH[1]}"
        # Extract decimal value from comment if available
        if [[ "${BASH_REMATCH[2]}" =~ \(([0-9]+)\) ]]; then
            chain_id="${BASH_REMATCH[1]}"
        else
            chain_id="$(hex_to_decimal "$chain_id_hex")"
        fi
    # Parse EVM Chain ID from default value
    elif [[ "$line" =~ EVMChainID:[[:space:]]*defaultEVMChainID ]]; then
        chain_id="$default_evm_chain_id_int"
    fi

    # Parse CometBFT Chain ID from explicit value
    if [[ "$line" =~ CometBFTChainID:[[:space:]]*\"([^\"]+)\" ]]; then
        chain_id="${BASH_REMATCH[1]}"
    # Parse CometBFT Chain ID from default value
    elif [[ "$line" =~ CometBFTChainID:[[:space:]]*defaultCometBFTChainID ]]; then
        chain_id="$default_cometbft_chain_id"
    fi

    # Check for archival check configuration
    if [[ "$line" =~ ArchivalCheckConfig ]]; then
        archival_check="✅"
    fi
done < "$INPUT_FILE"

echo "" >> "${OUTPUT_FILE}.new"

# Process Morse services
{
    echo "## Morse Protocol Services"
    echo ""
    echo "| Service Name | Authoritative Service ID | Service QoS Type | Chain ID (if applicable) | Archival Check Configured |"
    echo "|-------------|------------|-----------------|----------|---------------------------|"
} >> "${OUTPUT_FILE}.new"

# Parse Morse services
in_morse_section=false
current_service=""
service_id=""
service_name=""
service_type=""
chain_id=""
archival_check=""

while IFS= read -r line; do
    # Check if we're in the Morse services section
    if [[ "$line" =~ ^var\ morseServices\ = ]]; then
        in_morse_section=true
        continue
    fi

    # Check if we've reached the end of Morse services
    if [[ "$in_morse_section" == true && "$line" =~ ^}$ ]]; then
        # Process the last service before exiting
        if [[ -n "$service_id" ]]; then
            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >> "${OUTPUT_FILE}.new"
        fi
        in_morse_section=false
        continue
    fi

    # Skip if not in Morse section
    if [[ "$in_morse_section" != true ]]; then
        continue
    fi

    # Check for new service
    if [[ "$line" =~ ServiceConfig\{ ]]; then
        # Process the previous service if exists
        if [[ -n "$service_id" ]]; then
            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >> "${OUTPUT_FILE}.new"
        fi
        
        # Reset variables for new service
        service_id=""
        service_name=""
        chain_id=""
        archival_check=""
        
        # Determine service type
        if [[ "$line" =~ evm\.ServiceConfig ]]; then
            service_type="EVM"
        elif [[ "$line" =~ cometbft\.ServiceConfig ]]; then
            service_type="CometBFT"
        elif [[ "$line" =~ solana\.ServiceConfig ]]; then
            service_type="Solana"
        else
            service_type="Unknown"
        fi
    fi

    # Parse service ID and name
    if [[ "$line" =~ ServiceID:[[:space:]]*\"([^\"]+)\",[[:space:]]*//(.*) ]]; then
        service_id="${BASH_REMATCH[1]}"
        service_name="${BASH_REMATCH[2]}"
        # Remove leading/trailing spaces from service name
        service_name="$(echo "$service_name" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
    fi

    # Parse EVM Chain ID from explicit value
    if [[ "$line" =~ EVMChainID:[[:space:]]*\"([^\"]+)\",[[:space:]]*//(.*) ]]; then
        chain_id_hex="${BASH_REMATCH[1]}"
        # Extract decimal value from comment if available
        if [[ "${BASH_REMATCH[2]}" =~ \(([0-9]+)\) ]]; then
            chain_id="${BASH_REMATCH[1]}"
        else
            chain_id="$(hex_to_decimal "$chain_id_hex")"
        fi
    # Parse EVM Chain ID from default value
    elif [[ "$line" =~ EVMChainID:[[:space:]]*defaultEVMChainID ]]; then
        chain_id="$default_evm_chain_id_int"
    fi

    # Parse CometBFT Chain ID from explicit value
    if [[ "$line" =~ CometBFTChainID:[[:space:]]*\"([^\"]+)\" ]]; then
        chain_id="${BASH_REMATCH[1]}"
    # Parse CometBFT Chain ID from default value
    elif [[ "$line" =~ CometBFTChainID:[[:space:]]*defaultCometBFTChainID ]]; then
        chain_id="$default_cometbft_chain_id"
    fi

    # Check for archival check configuration
    if [[ "$line" =~ ArchivalCheckConfig ]]; then
        archival_check="✅"
    fi
done < "$INPUT_FILE"

# Create the final file by combining the preserved content and new content
cat "${OUTPUT_FILE}.tmp" > "$OUTPUT_FILE"
cat "${OUTPUT_FILE}.new" >> "$OUTPUT_FILE"

# Clean up temp files
rm "${OUTPUT_FILE}.tmp" "${OUTPUT_FILE}.new"

echo "Documentation successfully updated at $OUTPUT_FILE"