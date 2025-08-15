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
    
    if [ "$FILE_MODE" != "true" ] && [ -z "$SERVICE_IDS" ]; then
        print_status $RED "❌ Error: SERVICE_IDS parameter is required when not using file mode"
        exit 1
    fi
    
    if [ "$FILE_MODE" = "true" ] && [ -z "$SERVICE_FILE" ]; then
        print_status $RED "❌ Error: SERVICE_FILE parameter is required when using file mode"
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
    
    print_status $CYAN "   💾 Inserting service $service_id into database..."
    
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
        print_status $GREEN "   ✅ Successfully inserted/updated service: $service_id"
    else
        print_status $RED "   ❌ Failed to insert service: $service_id"
        print_status $RED "   📋 Database error: $db_result"
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
    print_status $BLUE "📋 Parameters:"
    if [ "$FILE_MODE" = "true" ]; then
        print_status $BLUE "   • Service File: $SERVICE_FILE"
    else
        print_status $BLUE "   • Service IDs: $SERVICE_IDS"
    fi
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
        print_status $CYAN "🔍 Processing service $processed/$total_services: $service_id"
        
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
        
        print_status $GREEN "   ✅ Parsed - Name: $service_name, Units: $compute_units, Owner: $owner_address"
        
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
    echo "🔧 Usage: $0 [OPTIONS] <service_ids|--file service_file> <rpc_node> <network_id>"
    echo ""
    echo "📝 Parameters:"
    echo "  service_ids       Comma-separated list of service IDs"
    echo "  rpc_node         RPC node endpoint"
    echo "  network_id       Network/chain ID"
    echo ""
    echo "🔧 Options:"
    echo "  -h, --help       Show this help message"
    echo "  -f, --file       Use file mode - read service IDs from file (one per line)"
    echo "  -d, --debug      Enable debug output"
    echo ""
    echo "🌍 Environment Variables:"
    echo "  DB_CONNECTION_STRING  PostgreSQL connection string"
    echo "  DEBUG               Set to 'true' to enable debug output"
    echo ""
    echo "💡 Examples:"
    echo "  # Using comma-separated service IDs:"
    echo "  export DB_CONNECTION_STRING='postgresql://user:pass@localhost:5435/portal_db'"
    echo "  $0 'eth,poly,sol' 'https://rpc.example.com:443' 'pocket'"
    echo ""
    echo "  # Using file mode:"
    echo "  echo -e 'eth\\npoly\\nsol' > services.txt"
    echo "  $0 --file services.txt 'https://rpc.example.com:443' 'pocket'"
    echo ""
    echo "  # With debug output:"
    echo "  $0 --debug 'eth' 'https://rpc.example.com:443' 'pocket'"
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
    SERVICE_FILE="$1"
    NODE="$2"
    NETWORK="$3"
else
    if [ $# -ne 3 ]; then
        print_status $RED "❌ Error: Invalid number of arguments"
        echo ""
        usage
        exit 1
    fi
    SERVICE_IDS="$1"
    NODE="$2"
    NETWORK="$3"
fi

# Enable debug mode if DEBUG environment variable is set
if [ "$DEBUG" = "true" ]; then
    DEBUG_MODE=true
fi

main