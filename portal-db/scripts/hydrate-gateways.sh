#!/bin/bash

# 🚀 Gateway Hydration Script for Portal DB
# This script ingests gateway addresses and populates the Portal DB gateways table

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
    if [ "$FILE_MODE" != "true" ] && [ -z "$GATEWAY_ADDRESSES" ]; then
        print_status $RED "❌ Error: --gateways parameter is required when not using --file mode"
        exit 1
    fi

    if [ "$FILE_MODE" = "true" ] && [ -z "$GATEWAY_FILE" ]; then
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

# 📊 Function to parse gateway info from pocketd output
parse_gateway_info() {
    local gateway_output="$1"

    # Parse stake amount and denom from YAML output
    local stake_amount=$(echo "$gateway_output" | grep -A 5 "stake:" | grep "amount:" | head -1 | awk '{print $2}' | tr -d '"')
    local stake_denom=$(echo "$gateway_output" | grep -A 5 "stake:" | grep "denom:" | head -1 | awk '{print $2}' | tr -d '"')

    echo "$stake_amount|$stake_denom"
}

# 💾 Function to insert gateway into database
insert_gateway() {
    local address=$1
    local stake_amount=$2
    local stake_denom=$3
    local network_id=$4
    local private_key_hex=$5

    echo -e "   💾 Inserting gateway ${CYAN}$address${NC} into database..."

    # Prepare the private key field - use NULL if empty
    local private_key_field="NULL"
    if [ -n "$private_key_hex" ]; then
        private_key_field="'$private_key_hex'"
    fi

    # Use psql to insert the gateway data
    local db_result
    db_result=$(psql "$DB_CONNECTION_STRING" -c "
        INSERT INTO gateways (gateway_address, stake_amount, stake_denom, network_id, gateway_private_key_hex)
        VALUES ('$address', $stake_amount, '$stake_denom', '$network_id', $private_key_field)
        ON CONFLICT (gateway_address) DO UPDATE SET
            stake_amount = EXCLUDED.stake_amount,
            stake_denom = EXCLUDED.stake_denom,
            network_id = EXCLUDED.network_id,
            gateway_private_key_hex = COALESCE(EXCLUDED.gateway_private_key_hex, gateways.gateway_private_key_hex),
            updated_at = CURRENT_TIMESTAMP;
    " 2>&1)
    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        local key_status="without private key"
        if [ -n "$private_key_hex" ]; then
            key_status="with private key"
        fi
        echo -e "   ✅ Successfully inserted/updated gateway: ${CYAN}$address${NC} ($key_status)"
    else
        echo -e "   ❌ Failed to insert gateway: ${CYAN}$address${NC}"
        echo -e "   📋 Database error: ${RED}$db_result${NC}"
        return 1
    fi
}

# 📁 Function to read gateway addresses from file
read_gateway_file() {
    local file_path=$1

    if [ ! -f "$file_path" ]; then
        print_status $RED "❌ Error: Gateway file '$file_path' not found"
        exit 1
    fi

    if [ ! -r "$file_path" ]; then
        print_status $RED "❌ Error: Gateway file '$file_path' is not readable"
        exit 1
    fi

    # Read file and filter out empty lines and comments
    # Process each line to extract address and optional private key
    local addresses=""
    while IFS= read -r line || [[ -n "$line" ]]; do
        # Skip empty lines and comments
        [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
        
        # Extract address and private key using tab/space separation
        local address=$(echo "$line" | awk '{print $1}')
        local private_key=$(echo "$line" | awk '{print $2}')
        
        # Skip if no address found
        [[ -z "$address" ]] && continue
        
        # Store the data in format: address|private_key (empty if not provided)
        if [ -n "$addresses" ]; then
            addresses="${addresses},${address}|${private_key}"
        else
            addresses="${address}|${private_key}"
        fi
    done < "$file_path"
    
    echo "$addresses"
}

# 🎯 Main function
main() {
    print_status $PURPLE "🚀 Starting Gateway Hydration Process"
    echo -e "📋 Parameters:"
    if [ "$FILE_MODE" = "true" ]; then
        echo -e "   • Gateway File: ${CYAN}${GATEWAY_FILE}${NC}"
    else
        echo -e "   • Gateway Addresses: ${CYAN}${GATEWAY_ADDRESSES}${NC}"
    fi
    echo -e "   • RPC Node: ${CYAN}${NODE}${NC}"
    echo -e "   • Network: ${CYAN}${NETWORK}${NC}"
    echo ""

    # Validate required parameters
    validate_params

    # Check if pocketd command is available
    if ! command -v pocketd &> /dev/null; then
        print_status $RED "❌ Error: pocketd command not found. Please ensure it's installed and in PATH."
        exit 1
    fi

    # Check if psql command is available
    if ! command -v psql &> /dev/null; then
        print_status $RED "❌ Error: psql command not found. Please ensure PostgreSQL client is installed."
        exit 1
    fi

    # Test database connection
    print_status $YELLOW "🔍 Testing database connection..."
    if ! psql "$DB_CONNECTION_STRING" -c "SELECT 1;" > /dev/null 2>&1; then
        print_status $RED "❌ Error: Unable to connect to database"
        exit 1
    fi
    print_status $GREEN "✅ Database connection successful"
    echo ""

    # Get gateway addresses from file or command line
    local gateway_addresses_string
    if [ "$FILE_MODE" = "true" ]; then
        print_status $YELLOW "📁 Reading gateway addresses from file: $GATEWAY_FILE"
        gateway_addresses_string=$(read_gateway_file "$GATEWAY_FILE")
        print_status $GREEN "✅ Read gateway addresses from file"
    else
        # For command line mode, format as address| (no private key)
        gateway_addresses_string=$(echo "$GATEWAY_ADDRESSES" | sed 's/,/|,/g' | sed 's/$/|/')
    fi

    # Convert comma-separated addresses to array
    IFS=',' read -ra ADDR_ARRAY <<< "$gateway_addresses_string"

    total_addresses=${#ADDR_ARRAY[@]}
    processed=0
    successful=0
    failed=0

    print_status $PURPLE "🔄 Processing $total_addresses gateway addresses..."
    echo ""

    # Process each gateway address
    for gateway_entry in "${ADDR_ARRAY[@]}"; do
        # Extract address and private key from entry (format: address|private_key)
        IFS='|' read -r address private_key_hex <<< "$gateway_entry"
        
        # Trim whitespace
        address=$(echo "$address" | xargs)
        private_key_hex=$(echo "$private_key_hex" | xargs)

        processed=$((processed + 1))
        
        # Show status with private key indicator
        local key_indicator=""
        if [ -n "$private_key_hex" ]; then
            key_indicator=" ${GREEN}[+key]${NC}"
        else
            key_indicator=" ${YELLOW}[-key]${NC}"
        fi
        
        echo -e "🔍 Processing gateway ${BLUE}$processed/${total_addresses}${NC}: ${CYAN}$address${NC}$key_indicator"

        # Query gateway information using pocketd with timeout
        print_status $YELLOW "   📡 Fetching gateway info from blockchain..."

        if ! gateway_output=$(timeout 30 pocketd q gateway show-gateway "$address" --node="$NODE" --chain-id="$NETWORK" 2>&1); then
            print_status $RED "   ❌ Failed to fetch gateway info for $address"
            if echo "$gateway_output" | grep -q "timeout"; then
                print_status $RED "   📋 Error: Command timed out after 30 seconds"
            else
                print_status $RED "   📋 Error: $gateway_output"
            fi
            failed=$((failed + 1))
            echo ""
            continue
        fi

        # Check if gateway exists (look for error indicators)
        if echo "$gateway_output" | grep -q "not found\|error\|Error"; then
            print_status $RED "   ❌ Gateway not found or error occurred for $address"
            print_status $RED "   📋 Response: $gateway_output"
            failed=$((failed + 1))
            echo ""
            continue
        fi

        print_status $GREEN "   ✅ Gateway info retrieved successfully"

        # Parse the gateway information
        print_status $YELLOW "   🔧 Parsing gateway information..."
        stake_info=$(parse_gateway_info "$gateway_output")

        if [ -z "$stake_info" ] || [ "$stake_info" = "|" ]; then
            print_status $RED "   ❌ Failed to parse stake information for $address"
            failed=$((failed + 1))
            echo ""
            continue
        fi

        IFS='|' read -r stake_amount stake_denom <<< "$stake_info"

        if [ -z "$stake_amount" ] || [ -z "$stake_denom" ]; then
            print_status $RED "   ❌ Invalid stake information parsed for $address"
            failed=$((failed + 1))
            echo ""
            continue
        fi

        echo -e "   ✅ Parsed - Amount: ${CYAN}$stake_amount${NC}, Denom: ${CYAN}$stake_denom${NC}"

        # Insert into database
        if insert_gateway "$address" "$stake_amount" "$stake_denom" "$NETWORK" "$private_key_hex"; then
            successful=$((successful + 1))
        else
            failed=$((failed + 1))
        fi

        echo ""
    done

    # Print final summary
    print_status $PURPLE "📊 Gateway Hydration Summary:"
    print_status $BLUE "   • Total Processed: $processed"
    print_status $GREEN "   • Successful: $successful"
    print_status $RED "   • Failed: $failed"

    if [ $failed -gt 0 ]; then
        print_status $YELLOW "⚠️  Some gateways failed to process. Check the output above for details."
        exit 1
    else
        print_status $GREEN "🎉 All gateways processed successfully!"
    fi
}

# 📚 Usage information
usage() {
    echo -e "${PURPLE}🔧 Usage:${NC} ${BLUE}$0 [OPTIONS]${NC}"
    echo ""
    echo -e "${YELLOW}📝 Required Parameters:${NC}"
    echo -e "  ${CYAN}--gateways <addrs>${NC}    Comma-separated list of gateway addresses"
    echo -e "  ${CYAN}--file <path>${NC}         Read gateway addresses from file (one per line)"
    echo -e "  ${CYAN}--node <endpoint>${NC}     RPC node endpoint"
    echo -e "  ${CYAN}--chain-id <id>${NC}       Network/chain ID"
    echo -e "  ${CYAN}--db-string <conn>${NC}    PostgreSQL connection string"
    echo ""
    echo -e "${YELLOW}🔧 Optional Parameters:${NC}"
    echo -e "  ${CYAN}-h, --help${NC}            Show this help message"
    echo -e "  ${CYAN}-d, --debug${NC}           Enable debug output"
    echo ""
    echo -e "${YELLOW}📋 Notes:${NC}"
    echo -e "  • Either ${CYAN}--gateways${NC} or ${CYAN}--file${NC} is required (but not both)"
    echo -e "  • All other parameters are required"
    echo -e "  • File format: Each line should contain address and optionally private key separated by tab/space"
    echo -e "  • File format example: ${CYAN}pokt1gateway123${NC} ${CYAN}deadbeef789${NC}"
    echo -e "  • Lines with only address (no private key) are supported"
    echo ""
    echo -e "${YELLOW}💡 Examples:${NC}"
    echo -e "  ${YELLOW}# Using comma-separated gateway addresses (space syntax):${NC}"
    echo -e "  ${GREEN}$0 --gateways 'addr1,addr2,addr3' \\\\${NC}"
    echo -e "  ${GREEN}     --node 'https://rpc.example.com:443' \\\\${NC}"
    echo -e "  ${GREEN}     --chain-id 'pocket' \\\\${NC}"
    echo -e "  ${GREEN}     --db-string 'postgresql://user:pass@localhost:5435/portal_db'${NC}"
    echo ""
    echo -e "  ${YELLOW}# Using comma-separated gateway addresses (equals syntax):${NC}"
    echo -e "  ${GREEN}$0 --gateways='addr1,addr2,addr3' \\\\${NC}"
    echo -e "  ${GREEN}     --node='https://rpc.example.com:443' \\\\${NC}"
    echo -e "  ${GREEN}     --chain-id='pocket' \\\\${NC}"
    echo -e "  ${GREEN}     --db-string='postgresql://user:pass@localhost:5435/portal_db'${NC}"
    echo ""
    echo -e "  ${YELLOW}# Using file mode:${NC}"
    echo -e "  ${GREEN}echo -e 'addr1\\\naddr2\\\naddr3' > gateways.txt${NC}"
    echo -e "  ${GREEN}$0 --file=gateways.txt \\\\${NC}"
    echo -e "  ${GREEN}     --node='https://rpc.example.com:443' \\\\${NC}"
    echo -e "  ${GREEN}     --chain-id='pocket' \\\\${NC}"
    echo -e "  ${GREEN}     --db-string='postgresql://user:pass@localhost:5435/portal_db'${NC}"
    echo ""
    echo -e "  ${YELLOW}# Using file mode with private keys:${NC}"
    echo -e "  ${GREEN}cat > gateways_with_keys.txt << EOF${NC}"
    echo -e "  ${GREEN}pokt1gateway123${TAB}deadbeef123456789abcdef${NC}"
    echo -e "  ${GREEN}pokt1gateway456${TAB}987654321fedcba${NC}"
    echo -e "  ${GREEN}pokt1gateway789${NC}"
    echo -e "  ${GREEN}EOF${NC}"
    echo -e "  ${GREEN}$0 --file=gateways_with_keys.txt \\\\${NC}"
    echo -e "  ${GREEN}     --node='https://rpc.example.com:443' \\\\${NC}"
    echo -e "  ${GREEN}     --chain-id='pocket' \\\\${NC}"
    echo -e "  ${GREEN}     --db-string='postgresql://user:pass@localhost:5435/portal_db'${NC}"
    echo ""
    echo -e "  ${YELLOW}# Using environment variable with mixed syntax:${NC}"
    echo -e "  ${GREEN}export PMAIN='--node=https://rpc.example.com:443 --chain-id=pocket'${NC}"
    echo -e "  ${GREEN}$0 --gateways='addr1' --db-string='postgresql://...' \$PMAIN${NC}"
    echo ""
    echo -e "  ${YELLOW}# With debug output:${NC}"
    echo -e "  ${GREEN}$0 --debug --gateways='addr1' \\\\${NC}"
    echo -e "  ${GREEN}     --node='https://rpc.example.com:443' \\\\${NC}"
    echo -e "  ${GREEN}     --chain-id='pocket' \\\\${NC}"
    echo -e "  ${GREEN}     --db-string='postgresql://user:pass@localhost:5435/portal_db'${NC}"
    echo ""
}

# 🚪 Entry point
# Initialize variables
DEBUG_MODE=false
FILE_MODE=false
GATEWAY_ADDRESSES=""
GATEWAY_FILE=""
NODE=""
NETWORK=""
DB_CONNECTION_STRING=""

# Parse arguments and flags
while [ $# -gt 0 ]; do
    case $1 in
        -h|--help|help)
            usage
            exit 0
            ;;
        -d|--debug)
            DEBUG_MODE=true
            DEBUG=true
            shift
            ;;
        --gateways=*)
            if [ "$FILE_MODE" = "true" ]; then
                print_status $RED "❌ Error: Cannot use both --gateways and --file"
                exit 1
            fi
            GATEWAY_ADDRESSES="${1#*=}"
            if [ -z "$GATEWAY_ADDRESSES" ]; then
                print_status $RED "❌ Error: --gateways requires a value"
                exit 1
            fi
            shift
            ;;
        --gateways)
            if [ -z "$2" ]; then
                print_status $RED "❌ Error: --gateways requires a value"
                exit 1
            fi
            if [ "$FILE_MODE" = "true" ]; then
                print_status $RED "❌ Error: Cannot use both --gateways and --file"
                exit 1
            fi
            GATEWAY_ADDRESSES="$2"
            shift 2
            ;;
        --file=*)
            if [ -n "$GATEWAY_ADDRESSES" ]; then
                print_status $RED "❌ Error: Cannot use both --gateways and --file"
                exit 1
            fi
            GATEWAY_FILE="${1#*=}"
            if [ -z "$GATEWAY_FILE" ]; then
                print_status $RED "❌ Error: --file requires a file path"
                exit 1
            fi
            FILE_MODE=true
            shift
            ;;
        --file)
            if [ -z "$2" ]; then
                print_status $RED "❌ Error: --file requires a file path"
                exit 1
            fi
            if [ -n "$GATEWAY_ADDRESSES" ]; then
                print_status $RED "❌ Error: Cannot use both --gateways and --file"
                exit 1
            fi
            FILE_MODE=true
            GATEWAY_FILE="$2"
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

# Check that we have either gateways or file mode
if [ "$FILE_MODE" != "true" ] && [ -z "$GATEWAY_ADDRESSES" ]; then
    print_status $RED "❌ Error: Either --gateways or --file must be specified"
    echo ""
    usage
    exit 1
fi

main
