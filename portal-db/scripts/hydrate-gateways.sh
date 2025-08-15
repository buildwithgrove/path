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
    if [ -z "$GATEWAY_ADDRESSES" ]; then
        print_status $RED "❌ Error: GATEWAY_ADDRESSES parameter is required"
        exit 1
    fi
    
    if [ -z "$NODE" ]; then
        print_status $RED "❌ Error: NODE parameter is required"
        exit 1
    fi
    
    if [ -z "$NETWORK" ]; then
        print_status $RED "❌ Error: NETWORK parameter is required"
        exit 1
    fi
    
    if [ -z "$DB_CONNECTION_STRING" ]; then
        print_status $RED "❌ Error: DB_CONNECTION_STRING environment variable is required"
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
    
    print_status $CYAN "   💾 Inserting gateway $address into database..."
    
    # Use psql to insert the gateway data
    local db_result
    db_result=$(psql "$DB_CONNECTION_STRING" -c "
        INSERT INTO gateways (gateway_address, stake_amount, stake_denom, network_id) 
        VALUES ('$address', $stake_amount, '$stake_denom', '$network_id')
        ON CONFLICT (gateway_address) DO UPDATE SET
            stake_amount = EXCLUDED.stake_amount,
            stake_denom = EXCLUDED.stake_denom,
            network_id = EXCLUDED.network_id,
            updated_at = CURRENT_TIMESTAMP;
    " 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        print_status $GREEN "   ✅ Successfully inserted/updated gateway: $address"
    else
        print_status $RED "   ❌ Failed to insert gateway: $address"
        print_status $RED "   📋 Database error: $db_result"
        return 1
    fi
}

# 🎯 Main function
main() {
    print_status $PURPLE "🚀 Starting Gateway Hydration Process"
    print_status $BLUE "📋 Parameters:"
    print_status $BLUE "   • Gateway Addresses: $GATEWAY_ADDRESSES"
    print_status $BLUE "   • RPC Node: $NODE"
    print_status $BLUE "   • Network: $NETWORK"
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
    
    # Convert comma-separated addresses to array
    IFS=',' read -ra ADDR_ARRAY <<< "$GATEWAY_ADDRESSES"
    
    total_addresses=${#ADDR_ARRAY[@]}
    processed=0
    successful=0
    failed=0
    
    print_status $PURPLE "🔄 Processing $total_addresses gateway addresses..."
    echo ""
    
    # Process each gateway address
    for address in "${ADDR_ARRAY[@]}"; do
        # Trim whitespace
        address=$(echo "$address" | xargs)
        
        processed=$((processed + 1))
        print_status $CYAN "🔍 Processing gateway $processed/$total_addresses: $address"
        
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
        
        print_status $GREEN "   ✅ Parsed - Amount: $stake_amount, Denom: $stake_denom"
        
        # Insert into database
        if insert_gateway "$address" "$stake_amount" "$stake_denom" "$NETWORK"; then
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
    echo "🔧 Usage: $0 [OPTIONS] <gateway_addresses> <rpc_node> <network_id>"
    echo ""
    echo "📝 Parameters:"
    echo "  gateway_addresses  Comma-separated list of gateway addresses"
    echo "  rpc_node          RPC node endpoint"
    echo "  network_id        Network/chain ID"
    echo ""
    echo "🔧 Options:"
    echo "  -h, --help        Show this help message"
    echo "  -d, --debug       Enable debug output"
    echo ""
    echo "🌍 Environment Variables:"
    echo "  DB_CONNECTION_STRING  PostgreSQL connection string"
    echo "  DEBUG               Set to 'true' to enable debug output"
    echo ""
    echo "💡 Examples:"
    echo "  export DB_CONNECTION_STRING='postgresql://user:pass@localhost:5435/portal_db'"
    echo "  $0 'addr1,addr2,addr3' 'https://rpc.example.com:443' 'pocket-mainnet'"
    echo ""
    echo "  # With debug output:"
    echo "  DEBUG=true $0 'addr1' 'https://rpc.example.com:443' 'pocket-mainnet'"
    echo "  # Or:"
    echo "  $0 --debug 'addr1' 'https://rpc.example.com:443' 'pocket-mainnet'"
    echo ""
}

# 🚪 Entry point
# Parse arguments and flags
DEBUG_MODE=false

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
        -*)
            print_status $RED "❌ Error: Unknown option $1"
            echo ""
            usage
            exit 1
            ;;
        *)
            break
            ;;
    esac
done

# Check if we have the right number of remaining arguments
if [ $# -ne 3 ]; then
    print_status $RED "❌ Error: Invalid number of arguments"
    echo ""
    usage
    exit 1
fi

GATEWAY_ADDRESSES="$1"
NODE="$2"
NETWORK="$3"

# Enable debug mode if DEBUG environment variable is set
if [ "$DEBUG" = "true" ]; then
    DEBUG_MODE=true
fi


main