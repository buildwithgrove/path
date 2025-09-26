#!/bin/bash

# 🚀 Service Hydration Script for Portal DB
# This script ingests service IDs and populates the Portal DB services table

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

    if [ "$FILE_MODE" != "true" ] && [ -z "$SERVICE_IDS" ]; then
        print_status $RED "❌ Error: --services parameter is required when not using --file mode"
        exit 1
    fi

    if [ "$FILE_MODE" = "true" ] && [ -z "$SERVICE_FILE" ]; then
        print_status $RED "❌ Error: --file parameter requires a file path"
        exit 1
    fi
}

# 📊 Function to parse service info from pocketd output
parse_service_info() {
    local service_output="$1"

    # Parse service information from YAML output
    local service_name=$(echo "$service_output" | grep "name:" | head -1 | awk '{print $2}' | tr -d '"')
    local compute_units=$(echo "$service_output" | grep "compute_units_per_relay:" | head -1 | awk '{print $2}' | tr -d '"')
    local owner_address=$(echo "$service_output" | grep "owner_address:" | head -1 | awk '{print $2}' | tr -d '"')

    echo "$service_name|$compute_units|$owner_address"
}

# 💾 Function to insert service into database
insert_service() {
    local service_id=$1
    local service_name=$2
    local compute_units=$3
    local owner_address=$4
    local network_id=$5

    echo -e "   💾 Inserting service ${CYAN}$service_id${NC} into database..."

    # Use psql to insert the service data
    local db_result
    db_result=$(psql "$DB_CONNECTION_STRING" -c "
        INSERT INTO services (service_id, service_name, compute_units_per_relay, service_domains, service_owner_address, network_id, active)
        VALUES ('$service_id', '$service_name', $compute_units, '{}', '$owner_address', '$network_id', false)
        ON CONFLICT (service_id) DO UPDATE SET
            service_name = EXCLUDED.service_name,
            compute_units_per_relay = EXCLUDED.compute_units_per_relay,
            service_owner_address = EXCLUDED.service_owner_address,
            network_id = EXCLUDED.network_id,
            active = EXCLUDED.active,
            updated_at = CURRENT_TIMESTAMP;
    " 2>&1)
    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        echo -e "   ✅ Successfully inserted/updated service: ${CYAN}$service_id${NC}"
    else
        echo -e "   ❌ Failed to insert service: ${CYAN}$service_id${NC}"
        echo -e "   📋 Database error: ${RED}$db_result${NC}"
        return 1
    fi
}

# 📁 Function to read service IDs from file
read_service_file() {
    local file_path=$1

    if [ ! -f "$file_path" ]; then
        print_status $RED "❌ Error: Service file '$file_path' not found"
        exit 1
    fi

    if [ ! -r "$file_path" ]; then
        print_status $RED "❌ Error: Service file '$file_path' is not readable"
        exit 1
    fi

    # Read file and filter out empty lines and comments
    grep -v '^#' "$file_path" | grep -v '^[[:space:]]*$' | tr '\n' ','
}

# 🎯 Main function
main() {
    print_status $PURPLE "🚀 Starting Service Hydration Process"
    echo -e "📋 Parameters:"
    if [ "$FILE_MODE" = "true" ]; then
        echo -e "   • Service File: ${CYAN}${SERVICE_FILE}${NC}"
    else
        echo -e "   • Service IDs: ${CYAN}${SERVICE_IDS}${NC}"
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

    # Get service IDs from file or command line
    local service_ids_string
    if [ "$FILE_MODE" = "true" ]; then
        print_status $YELLOW "📁 Reading service IDs from file: $SERVICE_FILE"
        service_ids_string=$(read_service_file "$SERVICE_FILE")
        # Remove trailing comma
        service_ids_string=${service_ids_string%,}
        print_status $GREEN "✅ Read service IDs from file"
    else
        service_ids_string="$SERVICE_IDS"
    fi

    # Convert comma-separated service IDs to array
    IFS=',' read -ra SERVICE_ARRAY <<< "$service_ids_string"

    total_services=${#SERVICE_ARRAY[@]}
    processed=0
    successful=0
    failed=0

    print_status $PURPLE "🔄 Processing $total_services service IDs..."
    echo ""

    # Process each service ID
    for service_id in "${SERVICE_ARRAY[@]}"; do
        # Trim whitespace
        service_id=$(echo "$service_id" | xargs)

        processed=$((processed + 1))
        echo -e "🔍 Processing service ${BLUE}$processed/${total_services}${NC}: ${CYAN}$service_id${NC}"

        # Query service information using pocketd with timeout
        print_status $YELLOW "   📡 Fetching service info from blockchain..."

        if ! service_output=$(timeout 30 pocketd q service show-service "$service_id" --node="$NODE" --chain-id="$NETWORK" 2>&1); then
            print_status $RED "   ❌ Failed to fetch service info for $service_id"
            if echo "$service_output" | grep -q "timeout"; then
                print_status $RED "   📋 Error: Command timed out after 30 seconds"
            else
                print_status $RED "   📋 Error: $service_output"
            fi
            failed=$((failed + 1))
            echo ""
            continue
        fi

        # Check if service exists (look for error indicators)
        if echo "$service_output" | grep -q "not found\|error\|Error"; then
            print_status $RED "   ❌ Service not found or error occurred for $service_id"
            print_status $RED "   📋 Response: $service_output"
            failed=$((failed + 1))
            echo ""
            continue
        fi

        print_status $GREEN "   ✅ Service info retrieved successfully"

        # Parse the service information
        print_status $YELLOW "   🔧 Parsing service information..."
        service_info=$(parse_service_info "$service_output")

        if [ -z "$service_info" ] || [ "$service_info" = "||" ]; then
            print_status $RED "   ❌ Failed to parse service information for $service_id"
            failed=$((failed + 1))
            echo ""
            continue
        fi

        IFS='|' read -r service_name compute_units owner_address <<< "$service_info"

        if [ -z "$service_name" ] || [ -z "$compute_units" ] || [ -z "$owner_address" ]; then
            print_status $RED "   ❌ Invalid service information parsed for $service_id"
            failed=$((failed + 1))
            echo ""
            continue
        fi

        echo -e "   ✅ Parsed - Name: ${CYAN}$service_name${NC}, Units: ${CYAN}$compute_units${NC}, Owner: ${CYAN}$owner_address${NC}"

        # Insert into database
        if insert_service "$service_id" "$service_name" "$compute_units" "$owner_address" "$NETWORK"; then
            successful=$((successful + 1))
        else
            failed=$((failed + 1))
        fi

        echo ""
    done

    # Print final summary
    print_status $PURPLE "📊 Service Hydration Summary:"
    print_status $BLUE "   • Total Processed: $processed"
    print_status $GREEN "   • Successful: $successful"
    print_status $RED "   • Failed: $failed"

    if [ $failed -gt 0 ]; then
        print_status $YELLOW "⚠️  Some services failed to process. Check the output above for details."
        exit 1
    else
        print_status $GREEN "🎉 All services processed successfully!"
    fi
}

# 📚 Usage information
usage() {
    echo -e "${PURPLE}🔧 Usage:${NC} ${BLUE}$0 [OPTIONS]${NC}"
    echo ""
    echo -e "${YELLOW}📝 Required Parameters:${NC}"
    echo -e "  ${CYAN}--services <ids>${NC}      Comma-separated list of service IDs"
    echo -e "  ${CYAN}--file <path>${NC}         Read service IDs from file (one per line)"
    echo -e "  ${CYAN}--node <endpoint>${NC}     RPC node endpoint"
    echo -e "  ${CYAN}--chain-id <id>${NC}       Network/chain ID"
    echo -e "  ${CYAN}--db-string <conn>${NC}    PostgreSQL connection string"
    echo ""
    echo -e "${YELLOW}🔧 Optional Parameters:${NC}"
    echo -e "  ${CYAN}-h, --help${NC}            Show this help message"
    echo -e "  ${CYAN}-d, --debug${NC}           Enable debug output"
    echo ""
    echo -e "${YELLOW}📋 Notes:${NC}"
    echo -e "  • Either ${CYAN}--services${NC} or ${CYAN}--file${NC} is required (but not both)"
    echo -e "  • All other parameters are required"
    echo ""
    echo -e "${YELLOW}💡 Examples:${NC}"
    echo -e "  ${YELLOW}# Using comma-separated service IDs (space syntax):${NC}"
    echo -e "  ${GREEN}$0 --services 'eth,poly,solana,xrplevm' \\\\${NC}"
    echo -e "  ${GREEN}     --node 'https://rpc.example.com:443' \\\\${NC}"
    echo -e "  ${GREEN}     --chain-id 'pocket' \\\\${NC}"
    echo -e "  ${GREEN}     --db-string 'postgresql://user:pass@localhost:5435/portal_db'${NC}"
    echo ""
    echo -e "  ${YELLOW}# Using comma-separated service IDs (equals syntax):${NC}"
    echo -e "  ${GREEN}$0 --services='eth,poly,solana,xrplevm' \\\\${NC}"
    echo -e "  ${GREEN}     --node='https://rpc.example.com:443' \\\\${NC}"
    echo -e "  ${GREEN}     --chain-id='pocket' \\\\${NC}"
    echo -e "  ${GREEN}     --db-string='postgresql://user:pass@localhost:5435/portal_db'${NC}"
    echo ""
    echo -e "  ${YELLOW}# Using file mode:${NC}"
    echo -e "  ${GREEN}echo -e 'eth\\\npoly\\\nsolana\\\nxrplevm' > services.txt${NC}"
    echo -e "  ${GREEN}$0 --file=services.txt \\\\${NC}"
    echo -e "  ${GREEN}     --node='https://rpc.example.com:443' \\\\${NC}"
    echo -e "  ${GREEN}     --chain-id='pocket' \\\\${NC}"
    echo -e "  ${GREEN}     --db-string='postgresql://user:pass@localhost:5435/portal_db'${NC}"
    echo ""
    echo -e "  ${YELLOW}# Using environment variable with mixed syntax:${NC}"
    echo -e "  ${GREEN}export PMAIN='--node=https://rpc.example.com:443 --chain-id=pocket'${NC}"
    echo -e "  ${GREEN}$0 --services='eth' --db-string='postgresql://...' \$PMAIN${NC}"
    echo ""
    echo -e "  ${YELLOW}# With debug output:${NC}"
    echo -e "  ${GREEN}$0 --debug --services='eth' \\\\${NC}"
    echo -e "  ${GREEN}     --node='https://rpc.example.com:443' \\\\${NC}"
    echo -e "  ${GREEN}     --chain-id='pocket' \\\\${NC}"
    echo -e "  ${GREEN}     --db-string='postgresql://user:pass@localhost:5435/portal_db'${NC}"
    echo ""
}

# 🚪 Entry point
# Initialize variables
DEBUG_MODE=false
FILE_MODE=false
SERVICE_IDS=""
SERVICE_FILE=""
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
        --services=*)
            if [ "$FILE_MODE" = "true" ]; then
                print_status $RED "❌ Error: Cannot use both --services and --file"
                exit 1
            fi
            SERVICE_IDS="${1#*=}"
            if [ -z "$SERVICE_IDS" ]; then
                print_status $RED "❌ Error: --services requires a value"
                exit 1
            fi
            shift
            ;;
        --services)
            if [ -z "$2" ]; then
                print_status $RED "❌ Error: --services requires a value"
                exit 1
            fi
            if [ "$FILE_MODE" = "true" ]; then
                print_status $RED "❌ Error: Cannot use both --services and --file"
                exit 1
            fi
            SERVICE_IDS="$2"
            shift 2
            ;;
        --file=*)
            if [ -n "$SERVICE_IDS" ]; then
                print_status $RED "❌ Error: Cannot use both --services and --file"
                exit 1
            fi
            SERVICE_FILE="${1#*=}"
            if [ -z "$SERVICE_FILE" ]; then
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
            if [ -n "$SERVICE_IDS" ]; then
                print_status $RED "❌ Error: Cannot use both --services and --file"
                exit 1
            fi
            FILE_MODE=true
            SERVICE_FILE="$2"
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

# Check that we have either services or file mode
if [ "$FILE_MODE" != "true" ] && [ -z "$SERVICE_IDS" ]; then
    print_status $RED "❌ Error: Either --services or --file must be specified"
    echo ""
    usage
    exit 1
fi

main
