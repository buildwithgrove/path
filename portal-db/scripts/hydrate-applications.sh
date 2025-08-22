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

    if [ "$FILE_MODE" != "true" ] && [ -z "$APPLICATION_ADDRESSES" ]; then
        print_status $RED "❌ Error: APPLICATION_ADDRESSES parameter is required when not using file mode"
        exit 1
    fi

    if [ "$FILE_MODE" = "true" ] && [ -z "$APPLICATION_FILE" ]; then
        print_status $RED "❌ Error: APPLICATION_FILE parameter is required when using file mode"
        exit 1
    fi
}

# 📊 Function to parse application info from pocketd output
parse_application_info() {
    local app_output="$1"

    # Parse application information from YAML output
    local app_address=$(echo "$app_output" | grep "address:" | head -1 | awk '{print $2}' | tr -d '"')
    local gateway_address=$(echo "$app_output" | sed -n '/delegatee_gateway_addresses:/,/^[^ ]/p' | grep "^  - " | head -1 | sed 's/^  - //' | tr -d '"')
    local service_id=$(echo "$app_output" | sed -n '/service_configs:/,/^[^ ]/p' | grep "service_id:" | head -1 | sed 's/.*service_id:[[:space:]]*//' | tr -d '"')
    local stake_amount=$(echo "$app_output" | grep -A 5 "stake:" | grep "amount:" | head -1 | awk '{print $2}' | tr -d '"')
    local stake_denom=$(echo "$app_output" | grep -A 5 "stake:" | grep "denom:" | head -1 | awk '{print $2}' | tr -d '"')

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

    echo -e "   💾 Inserting application ${CYAN}$app_address${NC} into database..."

    # Use psql to insert the application data
    local db_result
    db_result=$(psql "$DB_CONNECTION_STRING" -c "
        INSERT INTO applications (application_address, gateway_address, service_id, stake_amount, stake_denom, network_id)
        VALUES ('$app_address', '$gateway_address', '$service_id', $stake_amount, '$stake_denom', '$network_id')
        ON CONFLICT (application_address) DO UPDATE SET
            gateway_address = EXCLUDED.gateway_address,
            service_id = EXCLUDED.service_id,
            stake_amount = EXCLUDED.stake_amount,
            stake_denom = EXCLUDED.stake_denom,
            network_id = EXCLUDED.network_id,
            updated_at = CURRENT_TIMESTAMP;
    " 2>&1)
    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        echo -e "   ✅ Successfully inserted/updated application: ${CYAN}$app_address${NC}"
    else
        echo -e "   ❌ Failed to insert application: ${CYAN}$app_address${NC}"
        echo -e "   📋 Database error: ${RED}$db_result${NC}"
        return 1
    fi
}

# 📁 Function to read application addresses from file
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

    # Read file and filter out empty lines and comments
    grep -v '^#' "$file_path" | grep -v '^[[:space:]]*$' | tr '\n' ','
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

    # Get application addresses from file or command line
    local app_addresses_string
    if [ "$FILE_MODE" = "true" ]; then
        print_status $YELLOW "📁 Reading application addresses from file: $APPLICATION_FILE"
        app_addresses_string=$(read_application_file "$APPLICATION_FILE")
        # Remove trailing comma
        app_addresses_string=${app_addresses_string%,}
        print_status $GREEN "✅ Read application addresses from file"
    else
        app_addresses_string="$APPLICATION_ADDRESSES"
    fi

    # Convert comma-separated application addresses to array
    IFS=',' read -ra APP_ARRAY <<< "$app_addresses_string"

    total_applications=${#APP_ARRAY[@]}
    processed=0
    successful=0
    failed=0

    print_status $PURPLE "🔄 Processing $total_applications application addresses..."
    echo ""

    # Process each application address
    for app_address in "${APP_ARRAY[@]}"; do
        # Trim whitespace
        app_address=$(echo "$app_address" | xargs)

        processed=$((processed + 1))
        echo -e "🔍 Processing application ${BLUE}$processed/${total_applications}${NC}: ${CYAN}$app_address${NC}"

        # Query application information using pocketd with timeout
        print_status $YELLOW "   📡 Fetching application info from blockchain..."

        if ! app_output=$(timeout 30 pocketd q application show-application "$app_address" --node="$NODE" --chain-id="$NETWORK" 2>&1); then
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

        # Check if application exists (look for error indicators)
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

        IFS='|' read -r parsed_address gateway_address service_id stake_amount stake_denom <<< "$app_info"

        if [ -z "$parsed_address" ] || [ -z "$gateway_address" ] || [ -z "$service_id" ] || [ -z "$stake_amount" ] || [ -z "$stake_denom" ]; then
            print_status $RED "   ❌ Invalid application information parsed for $app_address"
            print_status $RED "   📋 Parsed data: address='$parsed_address' gateway='$gateway_address' service='$service_id' amount='$stake_amount' denom='$stake_denom'"
            failed=$((failed + 1))
            echo ""
            continue
        fi

        echo -e "   ✅ Parsed - Gateway: ${CYAN}$gateway_address${NC}, Service: ${CYAN}$service_id${NC}, Stake: ${CYAN}$stake_amount${NC} ${CYAN}$stake_denom${NC}"

        # Insert into database
        if insert_application "$parsed_address" "$gateway_address" "$service_id" "$stake_amount" "$stake_denom" "$NETWORK"; then
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
    echo -e "${PURPLE}🔧 Usage:${NC} ${BLUE}$0 [OPTIONS] <application_addresses|--file application_file> <rpc_node> <network_id>${NC}"
    echo ""
    echo -e "${YELLOW}📝 Parameters:${NC}"
    echo -e "  ${CYAN}application_addresses${NC}  Comma-separated list of application addresses"
    echo -e "  ${CYAN}rpc_node${NC}               RPC node endpoint"
    echo -e "  ${CYAN}network_id${NC}             Network/chain ID"
    echo ""
    echo -e "${YELLOW}🔧 Options:${NC}"
    echo -e "  ${CYAN}-h, --help${NC}            Show this help message"
    echo -e "  ${CYAN}-f, --file${NC}            Use file mode - read application addresses from file (one per line)"
    echo -e "  ${CYAN}-d, --debug${NC}           Enable debug output"
    echo ""
    echo -e "${YELLOW}🌍 Environment Variables:${NC}"
    echo -e "  ${CYAN}DB_CONNECTION_STRING${NC}  PostgreSQL connection string"
    echo -e "  ${CYAN}DEBUG${NC}                 Set to 'true' to enable debug output"
    echo ""
    echo -e "${YELLOW}💡 Examples:${NC}"
    echo -e "  ${YELLOW}# Using comma-separated application addresses:${NC}"
    echo -e "  ${GREEN}export DB_CONNECTION_STRING='postgresql://user:pass@localhost:5435/portal_db'${NC}"
    echo -e "  ${GREEN}$0 'pokt1abc123,pokt1def456' 'https://rpc.example.com:443' 'pocket'${NC}"
    echo ""
    echo -e "  ${YELLOW}# Using file mode:${NC}"
    echo -e "  ${GREEN}echo -e 'pokt1abc123\\\npokt1def456' > applications.txt${NC}"
    echo -e "  ${GREEN}$0 --file applications.txt 'https://rpc.example.com:443' 'pocket'${NC}"
    echo ""
    echo -e "  ${YELLOW}# With debug output:${NC}"
    echo -e "  ${GREEN}$0 --debug 'pokt1abc123' 'https://rpc.example.com:443' 'pocket'${NC}"
    echo ""
}

# 🚪 Entry point
# Parse arguments and flags
DEBUG_MODE=false
FILE_MODE=false

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
        -f|--file)
            FILE_MODE=true
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
if [ "$FILE_MODE" = "true" ]; then
    if [ $# -ne 3 ]; then
        print_status $RED "❌ Error: Invalid number of arguments for file mode"
        echo ""
        usage
        exit 1
    fi
    APPLICATION_FILE="$1"
    NODE="$2"
    NETWORK="$3"
else
    if [ $# -ne 3 ]; then
        print_status $RED "❌ Error: Invalid number of arguments"
        echo ""
        usage
        exit 1
    fi
    APPLICATION_ADDRESSES="$1"
    NODE="$2"
    NETWORK="$3"
fi

# Enable debug mode if DEBUG environment variable is set
if [ "$DEBUG" = "true" ]; then
    DEBUG_MODE=true
fi

main