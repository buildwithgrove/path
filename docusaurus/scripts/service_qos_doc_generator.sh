#!/bin/bash

# Script to parse service_qos_config.go and generate markdown tables
# Usage: ./generate_service_docs.sh <path/to/service_qos_config.go> <output_markdown_file>

# TODO_MVP(@commoddity): Update this file so the entirety of the output file is generated. This is a best practice for auto-generated files.

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
        head -n $((HEADER_LINE - 1)) "$OUTPUT_FILE" >"${OUTPUT_FILE}.tmp"
    else
        # If not found, preserve all content (we'll append to it)
        cat "$OUTPUT_FILE" >"${OUTPUT_FILE}.tmp"
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
done <"$INPUT_FILE"

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
} >"${OUTPUT_FILE}.new"

# Find lines with archival check and extract the service IDs
archival_lines=$(grep -n "evm\.NewEVMArchivalCheckConfig" "$INPUT_FILE")
archival_services=()

# Process each line containing archival check configuration
while IFS= read -r line; do
    # Extract the line number
    line_num=$(echo "$line" | cut -d: -f1)
    
    # Look up to 3 lines before the archival check to find the service ID
    service_id=$(sed -n "$((line_num-3)),$((line_num-1))p" "$INPUT_FILE" | grep -o '"[^"]*"' | head -1 | tr -d '"')
    
    if [[ -n "$service_id" ]]; then
        archival_services+=("$service_id")
    fi
done <<< "$archival_lines"

echo "Detected archival services: ${archival_services[*]}" >&2

# Convert the array to a space-separated string for easier search
archival_services_str=" ${archival_services[*]} "

# Create a map of service IDs to chain IDs for archival services
declare -A archival_chain_ids
# Extract chain IDs for archival services
for service_id in "${archival_services[@]}"; do
    # For each archival service, extract the chain ID from the config file
    chain_id=""
    
    # Check for F00C (Ethereum) - uses defaultEVMChainID
    if [[ "$service_id" == "F00C" ]]; then
        chain_id="$default_evm_chain_id_int"
    else
        # Extract chain ID from the line containing the service ID
        chain_line=$(grep -A 2 "\"$service_id\"" "$INPUT_FILE" | grep -v "evm\.NewEVMArchivalCheckConfig")
        if [[ "$chain_line" =~ \"([^\"]+)\", ]]; then
            chain_id_hex="${BASH_REMATCH[1]}"
            # Check if it's a hex value or defaultEVMChainID
            if [[ "$chain_id_hex" == "defaultEVMChainID" ]]; then
                chain_id="$default_evm_chain_id_int"
            else
                chain_id="$(hex_to_decimal "$chain_id_hex")"
            fi
        fi
    fi
    
    # Store in the map
    archival_chain_ids["$service_id"]="$chain_id"
    echo "Archival service $service_id has chain ID: ${archival_chain_ids[$service_id]}" >&2
done

# Parse Shannon services
in_shannon_section=false
service_id=""
service_name=""
service_type=""
chain_id=""
archival_check=""
comment_buffer=""

# First pass: process Shannon services
while IFS= read -r line; do
    # Check if we're in the Shannon services section
    if [[ "$line" =~ ^var\ shannonServices\ = ]]; then
        in_shannon_section=true
        continue
    fi

    # Check if we've reached the end of Shannon services section
    if [[ "$in_shannon_section" == true && "$line" =~ ^}$ ]]; then
        # Process the last service before exiting
        if [[ -n "$service_id" ]]; then
            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >>"${OUTPUT_FILE}.new"
        fi
        in_shannon_section=false
        continue
    fi

    # Skip if not in Shannon section
    if [[ "$in_shannon_section" != true ]]; then
        continue
    fi

    # Capture comments for service name extraction
    if [[ "$line" =~ ^[[:space:]]*//[[:space:]]*(.*)[[:space:]]*$ ]]; then
        comment_text="${BASH_REMATCH[1]}"

        # If it's a section header like "*** EVM Services ***", skip it
        if [[ "$comment_text" =~ ^\*\*\*.*\*\*\*$ || "$comment_text" =~ ^=+$ ]]; then
            comment_buffer=""
            continue
        fi

        # Store comment for potential service name
        comment_buffer="$comment_text"
        continue
    fi

    # Check for new service
    if [[ "$line" =~ evm\.NewEVMServiceQoSConfig ]]; then
        # Process the previous service if exists
        if [[ -n "$service_id" ]]; then
            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >>"${OUTPUT_FILE}.new"
        fi

        # Reset variables for new service
        service_type="EVM"
        chain_id=""
        archival_check=""

        # Determine service ID and set service_name
        # Single-line definition with ID on the same line
        if [[ "$line" =~ evm\.NewEVMServiceQoSConfig\([[:space:]]*\"([^\"]+)\" ]]; then
            service_id="${BASH_REMATCH[1]}"
            remainder_line="$line"
        else
            # Multi-line definition: read next line for service ID
            read -r remainder_line
            if [[ "$remainder_line" =~ ^[[:space:]]*\"([^\"]+)\" ]]; then
                service_id="${BASH_REMATCH[1]}"
            else
                service_id=""
            fi
        fi

        # Use the most recent comment as the service name
        if [[ -n "$comment_buffer" ]]; then
            service_name="$comment_buffer"
            comment_buffer=""
        else
            service_name="Unknown EVM Service"
        fi

        # Attempt to extract chain ID from remainder_line
        if [[ "$remainder_line" =~ \"([^\"]+)\",[[:space:]]*(nil|evm\.New) ]]; then
            chain_id_hex="${BASH_REMATCH[1]}"
            if [[ "$remainder_line" =~ //.*\(([0-9]+)\) ]]; then
                chain_id="${BASH_REMATCH[1]}"
            elif [[ "$chain_id_hex" == "defaultEVMChainID" ]]; then
                chain_id="$default_evm_chain_id_int"
            else
                chain_id="$(hex_to_decimal "$chain_id_hex")"
            fi
        fi

        # Check if this service ID is in the list of archival services
        if [[ "$archival_services_str" == *" $service_id "* ]]; then
            archival_check="✅"
        fi
    elif [[ "$line" =~ cometbft\.NewCometBFTServiceQoSConfig\([[:space:]]*\"([^\"]+)\",[[:space:]]*\"([^\"]+)\" ]]; then
        # Process the previous service if exists
        if [[ -n "$service_id" ]]; then
            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >>"${OUTPUT_FILE}.new"
        fi

        # Reset variables for new service
        service_id="${BASH_REMATCH[1]}"
        service_type="CometBFT"
        chain_id="${BASH_REMATCH[2]}"
        archival_check=""

        # Use the most recent comment as the service name
        if [[ -n "$comment_buffer" ]]; then
            service_name="$comment_buffer"
            comment_buffer=""
        else
            service_name="Unknown CometBFT Service"
        fi
    elif [[ "$line" =~ solana\.NewSolanaServiceQoSConfig\([[:space:]]*\"([^\"]+)\" ]]; then
        # Process the previous service if exists
        if [[ -n "$service_id" ]]; then
            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >>"${OUTPUT_FILE}.new"
        fi

        # Reset variables for new service
        service_id="${BASH_REMATCH[1]}"
        service_type="Solana"
        chain_id=""
        archival_check=""

        # Use the most recent comment as the service name
        if [[ -n "$comment_buffer" ]]; then
            service_name="$comment_buffer"
            comment_buffer=""
        else
            service_name="Solana"
        fi
    fi
done <"$INPUT_FILE"

echo "" >>"${OUTPUT_FILE}.new"

# Process Morse services
{
    echo "## Morse Protocol Services"
    echo ""
    echo "| Service Name | Authoritative Service ID | Service QoS Type | Chain ID (if applicable) | Archival Check Configured |"
    echo "|-------------|------------|-----------------|----------|---------------------------|"
} >>"${OUTPUT_FILE}.new"

# Parse Morse services
in_morse_section=false
service_id=""
service_name=""
service_type=""
chain_id=""
archival_check=""
comment_buffer=""

# Debug - special handling for F00C, F01C, F021, F036
echo "Checking for known archival services in morse section" >&2

# Reset to the beginning of the file for second pass
while IFS= read -r line; do
    # Check if we're in the Morse services section
    if [[ "$line" =~ ^var\ morseServices\ = ]]; then
        in_morse_section=true
        continue
    fi

    # Check if we've reached the end of Morse services section
    if [[ "$in_morse_section" == true && "$line" =~ ^}$ ]]; then
        # Process the last service before exiting
        if [[ -n "$service_id" ]]; then
            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >>"${OUTPUT_FILE}.new"
        fi
        in_morse_section=false
        break
    fi

    # Skip if not in Morse section
    if [[ "$in_morse_section" != true ]]; then
        continue
    fi

    # Capture comments for service name extraction
    if [[ "$line" =~ ^[[:space:]]*//[[:space:]]*(.*)[[:space:]]*$ ]]; then
        comment_text="${BASH_REMATCH[1]}"

        # If it's a section header like "*** EVM Services ***", skip it
        if [[ "$comment_text" =~ ^\*\*\*.*\*\*\*$ || "$comment_text" =~ ^=+$ ]]; then
            comment_buffer=""
            continue
        fi

        # Store comment for potential service name
        comment_buffer="$comment_text"
        continue
    fi

    # Check for new service - this should be done before processing any other attributes
    if [[ "$line" =~ evm\.NewEVMServiceQoSConfig ]]; then
        # Process the previous service if exists
        if [[ -n "$service_id" ]]; then
            # Debug output for special cases
            if [[ "$service_id" == "F00C" || "$service_id" == "F01C" || "$service_id" == "F021" || "$service_id" == "F036" ]]; then
                echo "Processing service $service_id with archival=$archival_check and chain_id=$chain_id" >&2
            fi
            
            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >>"${OUTPUT_FILE}.new"
        fi

        # Reset variables for new service
        service_type="EVM"
        chain_id=""
        archival_check=""

        # Determine service ID and set service_name
        # Single-line definition with ID on the same line
        if [[ "$line" =~ evm\.NewEVMServiceQoSConfig\([[:space:]]*\"([^\"]+)\" ]]; then
            service_id="${BASH_REMATCH[1]}"
            remainder_line="$line"
        else
            # Multi-line definition: read next line for service ID
            read -r remainder_line
            if [[ "$remainder_line" =~ ^[[:space:]]*\"([^\"]+)\" ]]; then
                service_id="${BASH_REMATCH[1]}"
            else
                service_id=""
            fi
        fi

        # Use the most recent comment as the service name
        if [[ -n "$comment_buffer" ]]; then
            service_name="$comment_buffer"
            comment_buffer=""
        else
            service_name="Unknown EVM Service"
        fi

        # Check if this service ID is in the list of archival services
        if [[ "$archival_services_str" == *" $service_id "* ]]; then
            archival_check="✅"
            echo "Marking $service_id as archival" >&2
            
            # For archival services, use the pre-extracted chain ID if available
            if [[ -n "${archival_chain_ids[$service_id]}" ]]; then
                chain_id="${archival_chain_ids[$service_id]}"
                echo "Using pre-extracted chain ID for $service_id: $chain_id" >&2
            else
                # Attempt to extract from the remainder_line as fallback
                if [[ "$remainder_line" =~ \"([^\"]+)\",[[:space:]]*(nil|evm\.New) ]]; then
                    chain_id_hex="${BASH_REMATCH[1]}"
                    if [[ "$remainder_line" =~ //.*\(([0-9]+)\) ]]; then
                        chain_id="${BASH_REMATCH[1]}"
                    elif [[ "$chain_id_hex" == "defaultEVMChainID" ]]; then
                        chain_id="$default_evm_chain_id_int"
                    else
                        chain_id="$(hex_to_decimal "$chain_id_hex")"
                    fi
                fi
            fi
        else
            # Normal (non-archival) service - extract chain ID as usual
            if [[ "$remainder_line" =~ \"([^\"]+)\",[[:space:]]*(nil|evm\.New) ]]; then
                chain_id_hex="${BASH_REMATCH[1]}"
                if [[ "$remainder_line" =~ //.*\(([0-9]+)\) ]]; then
                    chain_id="${BASH_REMATCH[1]}"
                elif [[ "$chain_id_hex" == "defaultEVMChainID" ]]; then
                    chain_id="$default_evm_chain_id_int"
                else
                    chain_id="$(hex_to_decimal "$chain_id_hex")"
                fi
            fi
        fi
    elif [[ "$line" =~ cometbft\.NewCometBFTServiceQoSConfig\([[:space:]]*\"([^\"]+)\",[[:space:]]*\"([^\"]+)\" ]]; then
        # Process the previous service if exists
        if [[ -n "$service_id" ]]; then
            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >>"${OUTPUT_FILE}.new"
        fi

        # Reset variables for new service
        service_id="${BASH_REMATCH[1]}"
        service_type="CometBFT"
        chain_id="${BASH_REMATCH[2]}"
        archival_check=""

        # Use the most recent comment as the service name
        if [[ -n "$comment_buffer" ]]; then
            service_name="$comment_buffer"
            comment_buffer=""
        else
            service_name="Unknown CometBFT Service"
        fi
    elif [[ "$line" =~ solana\.NewSolanaServiceQoSConfig\([[:space:]]*\"([^\"]+)\" ]]; then
        # Process the previous service if exists
        if [[ -n "$service_id" ]]; then
            echo "| $service_name | $service_id | $service_type | $chain_id | $archival_check |" >>"${OUTPUT_FILE}.new"
        fi

        # Reset variables for new service
        service_id="${BASH_REMATCH[1]}"
        service_type="Solana"
        chain_id=""
        archival_check=""

        # Use the most recent comment as the service name
        if [[ -n "$comment_buffer" ]]; then
            service_name="$comment_buffer"
            comment_buffer=""
        else
            service_name="Solana"
        fi
    fi
done <"$INPUT_FILE"

# Create the final file by combining the preserved content and new content
cat "${OUTPUT_FILE}.tmp" >"$OUTPUT_FILE"
cat "${OUTPUT_FILE}.new" >>"$OUTPUT_FILE"

# Clean up temp files
rm "${OUTPUT_FILE}.tmp" "${OUTPUT_FILE}.new"

echo "Documentation successfully updated at $OUTPUT_FILE"
