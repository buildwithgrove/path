#!/bin/bash

# 🚀 Application Hydration Script for Portal DB
# This script ingests application addresses and populates the Portal DB applications table

set -e

# 🎨 Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 📝 Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 🔍 Function to validate required parameters
validate_params() {
    if [ "$FILE_MODE" != "true" ] && [ -z "$APPLICATION_ADDRESSES" ]; then
        print_status $RED "❌ Error: Either --apps or --file is required"
        exit 1
    fi

    if [ "$FILE_MODE" = "true" ] && [ -z "$APPLICATION_FILE" ]; then
        print_status $RED "❌ Error: --file parameter requires a file path"
        exit 1
    fi

    if [ -z "$NODE" ]; then
        print_status $RED "❌ Error: --node parameter is required"
        exit 1
    fi

    if [ -z "$NETWORK" ]; then
        print_status $RED "❌ Error: --chain-id parameter is required"
        exit 1
    fi

    if [ -z "$DB_CONNECTION_STRING" ]; then
        print_status $RED "❌ Error: --db-string parameter is required"
        exit 1
    fi
}

# 📊 Function to parse application info from pocketd JSON output
parse_application_info() {
    local app_output="$1"

    # Parse application information from JSON output
    local app_address=$(echo "$app_output" | jq -r '.application.address // empty' 2>/dev/null)
    local gateway_address=$(echo "$app_output" | jq -r '.application.delegatee_gateway_addresses[0] // empty' 2>/dev/null)
    local service_id=$(echo "$app_output" | jq -r '.application.service_configs[0].service_id // empty' 2>/dev/null)
    local stake_amount=$(echo "$app_output" | jq -r '.application.stake.amount // empty' 2>/dev/null)
    local stake_denom=$(echo "$app_output" | jq -r '.application.stake.denom // empty' 2>/dev/null)

    # Fallback to text parsing if JSON parsing fails or jq is not available
    if [ -z "$app_address" ] || ! command -v jq &> /dev/null; then
        print_status $YELLOW "   📝 Using text parsing (jq not available or JSON parsing failed)"
        # Parse application information from YAML/text output
        app_address=$(echo "$app_output" | grep "address:" | head -1 | awk '{print $2}' | tr -d '"')
        gateway_address=$(echo "$app_output" | sed -n '/delegatee_gateway_addresses:/,/^[^ ]/p' | grep "^  - " | head -1 | sed 's/^  - //' | tr -d '"')
        service_id=$(echo "$app_output" | sed -n '/service_configs:/,/^[^ ]/p' | grep "service_id:" | head -1 | sed 's/.*service_id:[[:space:]]*//' | tr -d '"')
        stake_amount=$(echo "$app_output" | grep -A 5 "stake:" | grep "amount:" | head -1 | awk '{print $2}' | tr -d '"')
        stake_denom=$(echo "$app_output" | grep -A 5 "stake:" | grep "denom:" | head -1 | awk '{print $2}' | tr -d '"')
    fi

    echo "$app_address|$gateway_address|$service_id|$stake_amount|$stake_denom"
}

# 💾 Function to insert application into database
insert_application() {
    local app_address=$1
    local gateway_address=$2
    local service_id=$3
    local stake_amount=$4
    local stake_denom=$5
    local network_id=$6
    local private_key_hex=$7

    echo -e "   💾 Inserting application ${CYAN}$app_address${NC} into database..."

    # Prepare the private key field - use NULL if empty
    local private_key_field="NULL"
    if [ -n "$private_key_hex" ]; then
        private_key_field="'$private_key_hex'"
    fi

    # Use psql to insert the application data
    local db_result
    db_result=$(psql "$DB_CONNECTION_STRING" -c "
        INSERT INTO applications (application_address, gateway_address, service_id, stake_amount, stake_denom, network_id, application_private_key_hex)
        VALUES ('$app_address', '$gateway_address', '$service_id', $stake_amount, '$stake_denom', '$network_id', $private_key_field)
        ON CONFLICT (application_address) DO UPDATE SET
            gateway_address = EXCLUDED.gateway_address,
            service_id = EXCLUDED.service_id,
            stake_amount = EXCLUDED.stake_amount,
            stake_denom = EXCLUDED.stake_denom,
            network_id = EXCLUDED.network_id,
            application_private_key_hex = COALESCE(EXCLUDED.application_private_key_hex, applications.application_private_key_hex),
            updated_at = CURRENT_TIMESTAMP;
    " 2>&1)
    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        local key_status="without private key"
        if [ -n "$private_key_hex" ]; then
            key_status="with private key"
        fi
        echo -e "   ✅ Successfully inserted/updated application: ${CYAN}$app_address${NC} ($key_status)"
    else
        echo -e "   ❌ Failed to insert application: ${CYAN}$app_address${NC}"
        echo -e "   📋 Database error: ${RED}$db_result${NC}"
        return 1
    fi
}

# 📁 Function to read application data from file
read_application_file() {
    local file_path=$1

    if [ ! -f "$file_path" ]; then
        print_status $RED "❌ Error: Application file '$file_path' not found"
        exit 1
    fi

    if [ ! -r "$file_path" ]; then
        print_status $RED "❌ Error: Application file '$file_path' is not readable"
        exit 1
    fi

    # Read file and process each line
    # Format: address [service_id] [private_key]
    local app_data=""
    while IFS= read -r line || [[ -n "$line" ]]; do
        # Skip empty lines and comments
        [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
        
        # Parse the line: address service_id private_key (space/tab separated)
        local address=$(echo "$line" | awk '{print $1}')
        local service_id=$(echo "$line" | awk '{print $2}')
        local private_key=$(echo "$line" | awk '{print $3}')
        
        # Skip if no address found
        [[ -z "$address" ]] && continue
        
        # Store the data in format: address|service_id|private_key
        # Empty fields will be empty strings
        if [ -n "$app_data" ]; then
            app_data="${app_data},${address}|${service_id}|${private_key}"
        else
            app_data="${address}|${service_id}|${private_key}"
        fi
    done < "$file_path"
    
    echo "$app_data"
}

# 🎯 Main function
main() {
    print_status $PURPLE "🚀 Starting Application Hydration Process"
    echo -e "📋 Parameters:"
    if [ "$FILE_MODE" = "true" ]; then
        echo -e "   • Application File: ${CYAN}${APPLICATION_FILE}${NC}"
    else
        echo -e "   • Application Addresses: ${CYAN}${APPLICATION_ADDRESSES}${NC}"
    fi
    echo -e "   • RPC Node: ${CYAN}${NODE}${NC}"
    echo -e "   • Network: ${CYAN}${NETWORK}${NC}"
    echo ""

    # Validate required parameters
    validate_params

    # Check if required commands are available
    if ! command -v pocketd &> /dev/null; then
        print_status $RED "❌ Error: pocketd command not found. Please ensure it's installed and in PATH."
        exit 1
    fi

    if ! command -v psql &> /dev/null; then
        print_status $RED "❌ Error: psql command not found. Please ensure PostgreSQL client is installed."
        exit 1
    fi

    # Check if jq is available for JSON parsing
    if ! command -v jq &> /dev/null; then
        print_status $YELLOW "⚠️  jq not found. Will use text parsing as fallback."
    fi

    # Test database connection
    print_status $YELLOW "🔍 Testing database connection..."
    if ! psql "$DB_CONNECTION_STRING" -c "SELECT 1;" > /dev/null 2>&1; then
        print_status $RED "❌ Error: Unable to connect to database"
        exit 1
    fi
    print_status $GREEN "✅ Database connection successful"
    echo ""

    # Get application data from file or command line
    local app_data_string
    
    if [ "$FILE_MODE" = "true" ]; then
        print_status $YELLOW "📁 Reading application data from file: $APPLICATION_FILE"
        app_data_string=$(read_application_file "$APPLICATION_FILE")
        print_status $GREEN "✅ Read application data from file"
    else
        # For command line mode, format as address|| (no service override or private key)
        app_data_string=$(echo "$APPLICATION_ADDRESSES" | sed 's/,/||,/g' | sed 's/$/||/')
    fi

    # Convert comma-separated data to array
    IFS=',' read -ra APP_ARRAY <<< "$app_data_string"

    total_applications=${#APP_ARRAY[@]}
    processed=0
    successful=0
    failed=0

    print_status $PURPLE "🔄 Processing $total_applications application addresses..."
    echo ""

    # Process each application
    for app_entry in "${APP_ARRAY[@]}"; do
        # Extract address, service_id, and private key from entry (format: address|service_id|private_key)
        IFS='|' read -r app_address service_override private_key_hex <<< "$app_entry"
        
        # Trim whitespace
        app_address=$(echo "$app_address" | xargs)
        service_override=$(echo "$service_override" | xargs)
        private_key_hex=$(echo "$private_key_hex" | xargs)

        processed=$((processed + 1))
        
        # Show status indicators
        local key_indicator=""
        local service_indicator=""
        
        if [ -n "$private_key_hex" ]; then
            key_indicator=" ${GREEN}[+key]${NC}"
        else
            key_indicator=" ${YELLOW}[-key]${NC}"
        fi
        
        if [ -n "$service_override" ]; then
            service_indicator=" ${BLUE}[+svc:$service_override]${NC}"
        else
            service_indicator=" ${YELLOW}[-svc]${NC}"
        fi
        
        echo -e "🔍 Processing application ${BLUE}$processed/${total_applications}${NC}: ${CYAN}$app_address${NC}$key_indicator$service_indicator"

        # Query application information using pocketd with timeout
        print_status $YELLOW "   📡 Fetching application info from blockchain..."

        if ! app_output=$(timeout 30 pocketd q application show-application "$app_address" --node "$NODE" --chain-id "$NETWORK" --output json 2>&1); then
            print_status $RED "   ❌ Failed to fetch application info for $app_address"
            if echo "$app_output" | grep -q "timeout"; then
                print_status $RED "   📋 Error: Command timed out after 30 seconds"
            else
                print_status $RED "   📋 Error: $app_output"
            fi
            failed=$((failed + 1))
            echo ""
            continue
        fi

        # Check if application exists
        if echo "$app_output" | grep -q "not found\|error\|Error"; then
            print_status $RED "   ❌ Application not found or error occurred for $app_address"
            print_status $RED "   📋 Response: $app_output"
            failed=$((failed + 1))
            echo ""
            continue
        fi

        print_status $GREEN "   ✅ Application info retrieved successfully"

        # Parse the application information
        print_status $YELLOW "   🔧 Parsing application information..."
        app_info=$(parse_application_info "$app_output")

        if [ -z "$app_info" ] || [ "$app_info" = "||||" ]; then
            print_status $RED "   ❌ Failed to parse application information for $app_address"
            failed=$((failed + 1))
            echo ""
            continue
        fi

        IFS='|' read -r parsed_address gateway_address blockchain_service_id stake_amount stake_denom <<< "$app_info"

        if [ -z "$parsed_address" ] || [ -z "$gateway_address" ] || [ -z "$blockchain_service_id" ] || [ -z "$stake_amount" ] || [ -z "$stake_denom" ]; then
            print_status $RED "   ❌ Invalid application information parsed for $app_address"
            print_status $RED "   📋 Parsed data: address='$parsed_address' gateway='$gateway_address' service='$blockchain_service_id' amount='$stake_amount' denom='$stake_denom'"
            failed=$((failed + 1))
            echo ""
            continue
        fi

        # Determine which service ID to use
        local final_service_id="$blockchain_service_id"
        if [ -n "$service_override" ]; then
            final_service_id="$service_override"
            echo -e "   🔄 Using service override: ${YELLOW}$blockchain_service_id${NC} → ${CYAN}$final_service_id${NC}"
        fi

        echo -e "   ✅ Parsed - Gateway: ${CYAN}$gateway_address${NC}, Service: ${CYAN}$final_service_id${NC}, Stake: ${CYAN}$stake_amount${NC} ${CYAN}$stake_denom${NC}"

        # Insert application into database
        if insert_application "$parsed_address" "$gateway_address" "$final_service_id" "$stake_amount" "$stake_denom" "$NETWORK" "$private_key_hex"; then
            successful=$((successful + 1))
        else
            failed=$((failed + 1))
        fi

        echo ""
    done

    # Print final summary
    print_status $PURPLE "📊 Application Hydration Summary:"
    print_status $BLUE "   • Total Processed: $processed"
    print_status $GREEN "   • Successful: $successful"
    print_status $RED "   • Failed: $failed"

    if [ $failed -gt 0 ]; then
        print_status $YELLOW "⚠️  Some applications failed to process. Check the output above for details."
        exit 1
    else
        print_status $GREEN "🎉 All applications processed successfully!"
    fi
}

# 📚 Usage information
usage() {
    echo -e "${PURPLE}🔧 Usage:${NC} ${BLUE}$0 [OPTIONS]${NC}"
    echo ""
    echo -e "${YELLOW}📝 Required Parameters:${NC}"
    echo -e "  ${CYAN}--apps <addresses>${NC}     Comma-separated list of application addresses"
    echo -e "  ${CYAN}--file <path>${NC}          Read application data from file"
    echo -e "  ${CYAN}-f <path>${NC}              Short form of --file"
    echo -e "  ${CYAN}--node <endpoint>${NC}      RPC node endpoint"
    echo -e "  ${CYAN}--chain-id <id>${NC}        Network/chain ID"
    echo -e "  ${CYAN}--db-string <conn>${NC}     PostgreSQL connection string"
    echo ""
    echo -e "${YELLOW}🔧 Optional Parameters:${NC}"
    echo -e "  ${CYAN}-h, --help${NC}             Show this help message"
    echo -e "  ${CYAN}-d, --debug${NC}            Enable debug output"
    echo ""
    echo -e "${YELLOW}📋 File Format:${NC}"
    echo -e "  • Format: ${CYAN}address [service_id] [private_key]${NC}"
    echo -e "  • Only address is required, service_id and private_key are optional"
    echo -e "  • Fields separated by spaces or tabs"
    echo -e "  • Lines starting with # are ignored"
    echo ""
    echo -e "${YELLOW}📋 File Examples:${NC}"
    echo -e "  ${GREEN}# applications.txt${NC}"
    echo -e "  ${GREEN}pokt1address123${NC}                                    ${GRAY}# address only${NC}"
    echo -e "  ${GREEN}pokt1address456 ethereum${NC}                          ${GRAY}# address + service override${NC}"
    echo -e "  ${GREEN}pokt1address789 polygon deadbeef123${NC}               ${GRAY}# address + service + private key${NC}"
    echo -e "  ${GREEN}pokt1address012 \"\" abc123456${NC}                       ${GRAY}# address + private key (no service override)${NC}"
    echo ""
    echo -e "${YELLOW}💡 Examples:${NC}"
    echo -e "  ${YELLOW}# Using comma-separated addresses:${NC}"
    echo -e "  ${GREEN}$0 --apps 'pokt1abc123,pokt1def456' \\\\${NC}"
    echo -e "  ${GREEN}     --node 'https://rpc.example.com:443' \\\\${NC}"
    echo -e "  ${GREEN}     --chain-id 'pocket' \\\\${NC}"
    echo -e "  ${GREEN}     --db-string 'postgresql://user:pass@localhost:5432/portal_db'${NC}"
    echo ""
    echo -e "  ${YELLOW}# Using file mode:${NC}"
    echo -e "  ${GREEN}$0 --file applications.txt \\\\${NC}"
    echo -e "  ${GREEN}     --node 'https://rpc.example.com:443' \\\\${NC}"
    echo -e "  ${GREEN}     --chain-id 'pocket' \\\\${NC}"
    echo -e "  ${GREEN}     --db-string 'postgresql://user:pass@localhost:5432/portal_db'${NC}"
    echo ""
}

# 🚪 Entry point - Initialize variables
DEBUG_MODE=false
FILE_MODE=false
APPLICATION_ADDRESSES=""
APPLICATION_FILE=""
NODE=""
NETWORK=""
DB_CONNECTION_STRING=""

# Parse arguments
while [ $# -gt 0 ]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -d|--debug)
            DEBUG_MODE=true
            shift
            ;;
        --apps=*)
            if [ "$FILE_MODE" = "true" ]; then
                print_status $RED "❌ Error: Cannot use both --apps and --file"
                exit 1
            fi
            APPLICATION_ADDRESSES="${1#*=}"
            if [ -z "$APPLICATION_ADDRESSES" ]; then
                print_status $RED "❌ Error: --apps requires a value"
                exit 1
            fi
            shift
            ;;
        --apps)
            if [ -z "$2" ]; then
                print_status $RED "❌ Error: --apps requires a value"
                exit 1
            fi
            if [ "$FILE_MODE" = "true" ]; then
                print_status $RED "❌ Error: Cannot use both --apps and --file"
                exit 1
            fi
            APPLICATION_ADDRESSES="$2"
            shift 2
            ;;
        --file=*|-f=*)
            if [ -n "$APPLICATION_ADDRESSES" ]; then
                print_status $RED "❌ Error: Cannot use both --apps and --file"
                exit 1
            fi
            APPLICATION_FILE="${1#*=}"
            if [ -z "$APPLICATION_FILE" ]; then
                print_status $RED "❌ Error: --file requires a file path"
                exit 1
            fi
            FILE_MODE=true
            shift
            ;;
        --file|-f)
            if [ -z "$2" ]; then
                print_status $RED "❌ Error: --file requires a file path"
                exit 1
            fi
            if [ -n "$APPLICATION_ADDRESSES" ]; then
                print_status $RED "❌ Error: Cannot use both --apps and --file"
                exit 1
            fi
            FILE_MODE=true
            APPLICATION_FILE="$2"
            shift 2
            ;;
        --node=*)
            NODE="${1#*=}"
            if [ -z "$NODE" ]; then
                print_status $RED "❌ Error: --node requires a value"
                exit 1
            fi
            shift
            ;;
        --node)
            if [ -z "$2" ]; then
                print_status $RED "❌ Error: --node requires a value"
                exit 1
            fi
            NODE="$2"
            shift 2
            ;;
        --chain-id=*)
            NETWORK="${1#*=}"
            if [ -z "$NETWORK" ]; then
                print_status $RED "❌ Error: --chain-id requires a value"
                exit 1
            fi
            shift
            ;;
        --chain-id)
            if [ -z "$2" ]; then
                print_status $RED "❌ Error: --chain-id requires a value"
                exit 1
            fi
            NETWORK="$2"
            shift 2
            ;;
        --db-string=*)
            DB_CONNECTION_STRING="${1#*=}"
            if [ -z "$DB_CONNECTION_STRING" ]; then
                print_status $RED "❌ Error: --db-string requires a value"
                exit 1
            fi
            shift
            ;;
        --db-string)
            if [ -z "$2" ]; then
                print_status $RED "❌ Error: --db-string requires a value"
                exit 1
            fi
            DB_CONNECTION_STRING="$2"
            shift 2
            ;;
        -*)
            print_status $RED "❌ Error: Unknown option $1"
            echo ""
            usage
            exit 1
            ;;
        *)
            print_status $RED "❌ Error: Unexpected argument $1. All parameters must be specified with flags."
            echo ""
            usage
            exit 1
            ;;
    esac
done

# Enable debug mode if DEBUG environment variable is set
if [ "$DEBUG" = "true" ]; then
    DEBUG_MODE=true
fi

# Validate that we have either apps or file mode
if [ "$FILE_MODE" != "true" ] && [ -z "$APPLICATION_ADDRESSES" ]; then
    print_status $RED "❌ Error: Either --apps or --file must be specified"
    echo ""
    usage
    exit 1
fi

# Run the main function
main
