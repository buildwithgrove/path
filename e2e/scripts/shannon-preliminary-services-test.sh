#!/bin/bash

################################################################################
# Shannon Services Supplier Report Generator
################################################################################
#
# DESCRIPTION:
#   This script queries supplier counts for a predefined list of service IDs and 
#   optionally tests each service with suppliers using curl requests. It generates 
#   both console output and a detailed CSV report. With the --disqualified_endpoints
#   flag, it also queries and stores disqualified endpoints data in JSON format.
#
# USAGE:
#   ./services-shannon.sh --network <beta|mainnet> --environment <local|production> [--disqualified_endpoints] [--portal_app_id <id>] [--api_key <key>]
#
# ARGUMENTS:
#   -n, --network              Network to use: 'beta' or 'mainnet' (required)
#   -e, --environment          Environment to use: 'local' or 'production' (required)
#   -d, --disqualified_endpoints  Also query disqualified endpoints and update JSON file (optional)
#   -p, --portal_app_id        Portal Application ID for production (required when environment=production)
#   -k, --api_key              API Key for production (required when environment=production)
#
# EXAMPLE USAGE:
#   ./services-shannon.sh --network beta --environment local
#   ./services-shannon.sh --network mainnet --environment production --portal_app_id "your_app_id" --api_key "your_api_key"
#   ./services-shannon.sh --network beta --environment local --disqualified_endpoints
#
# ENVIRONMENT BEHAVIOR:
#   local:      Uses http://localhost:3069 with Target-Service-Id header
#   production: Uses https://<service>.rpc.grove.city URLs with Portal-Application-Id and Authorization headers
#
# OUTPUT:
#   - Console: Summary table showing service IDs, supplier counts, and test results
#   - JSON File: Detailed report with test results (supplier_report_YYYY-MM-DD_HH:MM:SS.json)
#   - JSON File: (with -d flag) Disqualified endpoints data (sanctioned-endpoint-results.json)
#
# FEATURES:
#   - Embedded service list for easy maintenance
#   - Automatic network node configuration
#   - Skips services with 0 suppliers
#   - Tests services with curl requests (5 retries per service)
#   - Captures detailed error information in CSV
#   - Optional disqualified endpoints querying and JSON storage
#   - Clean console output for quick scanning
#
# CSV COLUMNS:
#   service_id, suppliers, test_result, error_message, endpoint_response, unmarshaling_error
#
# JSON STRUCTURE (with -d flag):
#   {"suppliers_passed": {services with working JSON-RPC and suppliers}, 
#    "suppliers_failed": {services with failed JSON-RPC but suppliers, includes errors array},
#    "no_suppliers": {services with no suppliers}}
#
################################################################################

# Embedded Services Array
# Modify this array to include or exclude services
SERVICES=(
    "tia_da"
    "tia_cons"
    "tia_da_test"
    "tia_cons_test"
    "arb_one"
    "arb_sep_test"
    "avax"
    "avax-dfk"
    "base"
    "base-test"
    "bitcoin"
    "blast"
    "bsc"
    "boba"
    "celo"
    "eth"
    "eth_hol_test"
    "eth_sep_test"
    "evmos"
    "fantom"
    "fraxtal"
    "fuse"
    "gnosis"
    "harmony"
    "iotex"
    "kaia"
    "kava"
    "metis"
    "moonbeam"
    "moonriver"
    "near"
    "oasys"
    "op"
    "op_sep_test"
    "opbnb"
    "osmosis"
    "poly"
    "poly_amoy_test"
    "radix"
    "scroll"
    "solana"
    "sui"
    "taiko"
    "taiko_hek_test"
    "poly_zkevm"
    "zklink_nova"
    "zksync_era"
    "xrpl_evm_dev"
    "sonic"
    "tron"
    "linea"
    "ink"
    "mantle"
    "sei"
    "bera"
    "xrpl_evm_testnet"
)

# Configuration Variables
TARGET_SERVICE_HEADER="Target-Service-Id"
JSONRPC_TEST_PAYLOAD='{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}'
MAX_RETRIES=5
RETRY_SLEEP_DURATION=1

# Disqualified endpoints configuration
DISQUALIFIED_API_URL="http://localhost:3069/disqualified_endpoints"

# Function to display usage
usage() {
    echo "Usage: $0 --network <beta|mainnet> --environment <local|production> [--disqualified_endpoints] [--portal_app_id <id>] [--api_key <key>]"
    echo "  -n, --network              Network to use: 'beta' or 'mainnet' (required)"
    echo "  -e, --environment          Environment to use: 'local' or 'production' (required)"
    echo "  -d, --disqualified_endpoints  Also query disqualified endpoints and add to JSON report (optional)"
    echo "  -p, --portal_app_id        Portal Application ID for production (required when environment=production)"
    echo "  -k, --api_key              API Key for production (required when environment=production)"
    echo ""
    echo "Examples:"
    echo "  $0 --network beta --environment local"
    echo "  $0 --network mainnet --environment production --portal_app_id your_app_id --api_key your_api_key"
    echo "  $0 --network beta --environment local --disqualified_endpoints"
    exit 1
}

# Parse command line arguments
NETWORK=""
ENVIRONMENT=""
QUERY_DISQUALIFIED=false
PORTAL_APP_ID=""
API_KEY=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--network)
            NETWORK="$2"
            shift 2
            ;;
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -d|--disqualified_endpoints)
            QUERY_DISQUALIFIED=true
            shift
            ;;
        -p|--portal_app_id)
            PORTAL_APP_ID="$2"
            shift 2
            ;;
        -k|--api_key)
            API_KEY="$2"
            shift 2
            ;;
        *)
            echo "Error: Unknown argument '$1'"
            usage
            ;;
    esac
done

# Check if network was provided and valid
if [ -z "$NETWORK" ]; then
    echo "Error: --network flag is required"
    usage
fi

if [[ "$NETWORK" != "beta" && "$NETWORK" != "mainnet" ]]; then
    echo "Error: --network must be either 'beta' or 'mainnet'"
    usage
fi

# Check if environment was provided and valid
if [ -z "$ENVIRONMENT" ]; then
    echo "Error: --environment flag is required"
    usage
fi

if [[ "$ENVIRONMENT" != "local" && "$ENVIRONMENT" != "production" ]]; then
    echo "Error: --environment must be either 'local' or 'production'"
    usage
fi

# Check if production-specific parameters are provided when needed
if [ "$ENVIRONMENT" = "production" ]; then
    if [ -z "$PORTAL_APP_ID" ]; then
        echo "Error: --portal_app_id is required when environment is production"
        usage
    fi
    
    if [ -z "$API_KEY" ]; then
        echo "Error: --api_key is required when environment is production"
        usage
    fi
fi

# Configure URLs and headers based on environment
if [ "$ENVIRONMENT" = "local" ]; then
    BASE_PATH_URL="http://localhost:3069/v1"
    BASE_DISQUALIFIED_URL="http://localhost:3069/disqualified_endpoints"
    USE_SUBDOMAIN=false
elif [ "$ENVIRONMENT" = "production" ]; then
    BASE_PATH_URL="https://rpc.grove.city/v1"
    BASE_DISQUALIFIED_URL="https://rpc.grove.city/disqualified_endpoints"
    USE_SUBDOMAIN=true
fi

# Set node flag based on network
if [ "$NETWORK" = "beta" ]; then
    NODE_FLAG="--node https://shannon-testnet-grove-rpc.beta.poktroll.com"
elif [ "$NETWORK" = "mainnet" ]; then
    NODE_FLAG="--node https://shannon-grove-rpc.mainnet.poktroll.com"
fi

# Check if PATH service is running via health endpoint
echo ""
if [ "$ENVIRONMENT" = "local" ]; then
    echo "🏥 Checking if PATH service is running..."
    health_response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3069/healthz 2>/dev/null)

    if [ "$health_response" != "200" ]; then
        echo ""
        echo "❌ ERROR: PATH service is not running or not healthy (HTTP $health_response)"
        echo ""
        echo "⚠️ IMPORTANT: Ensure you are running PATH locally before proceeding."
        echo "    👀 See instructions here: https://www.notion.so/buildwithgrove/PATH-on-Shannon-Load-Tests-200a36edfff6805296c9ce10f2066de6?source=copy_link#205a36edfff68087b27dd086a28f21e9"
        echo ""
        echo "🚪 Exiting without testing services."
        exit 1
    fi

    echo "✅ PATH service is healthy - proceeding with service tests"
else
    echo "🌍 Using production environment - skipping local health check"
fi
echo ""

# Initialize temporary files for storing results
TEMP_DIR=$(mktemp -d)
RESULTS_FILE="$TEMP_DIR/results.txt"
JSON_REPORT_FILE="supplier_report_$(date +%Y-%m-%d_%H:%M:%S).json"
trap 'rm -rf "$TEMP_DIR"' EXIT

# Initialize array to store services with suppliers
declare -a services_with_suppliers

echo "=== SUPPLIER COUNT REPORT ==="
echo "Generated: $(date)"
echo "Network: $NETWORK"
echo "Environment: $ENVIRONMENT"
if [ "$QUERY_DISQUALIFIED" = true ]; then
    echo "Mode: Including disqualified endpoints analysis 📊"
else
    echo "Mode: Basic supplier count and testing only 📈"
fi
echo "=============================="
echo ""

# Iterate through each service in the embedded array
for service in "${SERVICES[@]}"; do
    # Skip empty services (shouldn't happen with embedded array, but safety check)
    if [[ -z "$service" ]]; then
        continue
    fi

    echo "🔍 Querying service: $service..."

    # Run the command and capture output
    output=$(pocketd q supplier list-suppliers --service-id "$service" $NODE_FLAG 2>/dev/null)

    if [ $? -eq 0 ]; then
        # Extract the total count from pagination section
        total=$(echo "$output" | grep -E "^\s*total:" | sed 's/.*total: *"\([0-9]*\)".*/\1/')

        if [[ "$total" =~ ^[0-9]+$ ]]; then
            echo "$service:$total" >> "$RESULTS_FILE"
            echo "  ✅ Found $total suppliers"
            
            # Only add to array if suppliers were found
            if [ "$total" -gt 0 ]; then
                services_with_suppliers+=("$service:$total")
            fi
        else
            echo "$service:ERROR" >> "$RESULTS_FILE"
            echo "  🚫 Found 0 suppliers for $service, skipping..."
        fi
    else
        echo "$service:FAILED" >> "$RESULTS_FILE"
        echo "  💥 Command failed"
    fi

done

# Display services with suppliers as a table
if [ ${#services_with_suppliers[@]} -gt 0 ]; then
    echo ""
    echo "=============================="
    echo "👥 SERVICES WITH SUPPLIERS:"
    echo "=============================="
    printf "%-20s | %s\n" "SERVICE ID" "SUPPLIERS"
    printf "%-20s-+-%s\n" "--------------------" "----------"
    
    for item in "${services_with_suppliers[@]}"; do
        IFS=':' read -r service count <<< "$item"
        printf "%-20s | %s\n" "$service" "$count"
    done
    
    echo "=============================="
    
    echo "=============================="
    echo "🧪 TESTING SERVICES..."
    echo "=============================="
    
    # Initialize JSON report structure with three categories
    echo '{"suppliers_failed": {}, "suppliers_passed": {}, "no_suppliers": [], "summary": {}}' > "$TEMP_DIR/json_report.json"
    
    # Array to store all results for sorting
    declare -a all_results
    
    for item in "${services_with_suppliers[@]}"; do
        IFS=':' read -r service count <<< "$item"
        echo "🚀 Testing $service..."
        
        # Execute 5 curl requests and require ALL to succeed
        request_count=0
        successful_requests=0
        failed_requests=0
        max_requests=5
        declare -a error_responses=()
        declare -a detailed_errors=()
        
        while [ $request_count -lt $max_requests ]; do
            request_count=$((request_count + 1))
            echo "    📡 Request $request_count/$max_requests..."
            
            # Construct URL and headers based on environment
            if [ "$USE_SUBDOMAIN" = true ]; then
                # Production: use subdomain format with required headers
                service_url="https://${service}.rpc.grove.city/v1"
                curl_result=$(curl -s "$service_url" \
                    -H "Portal-Application-Id: $PORTAL_APP_ID" \
                    -H "Authorization: $API_KEY" \
                    -d "$JSONRPC_TEST_PAYLOAD" 2>/dev/null)
            else
                # Local: use header format
                curl_result=$(curl -s "$BASE_PATH_URL" \
                    -H "$TARGET_SERVICE_HEADER: $service" \
                    -d "$JSONRPC_TEST_PAYLOAD" 2>/dev/null)
            fi
            
            # Check if curl was successful and returned valid JSON
            if [ $? -eq 0 ] && echo "$curl_result" | jq -e . >/dev/null 2>&1; then
                # Check if response contains an error field
                if echo "$curl_result" | jq -e '.error' >/dev/null 2>&1; then
                    echo "      ❌ Failed (JSON-RPC error)"
                    failed_requests=$((failed_requests + 1))
                    # Collect error response
                    error_response=$(echo "$curl_result" | jq -c '.error')
                    error_responses+=("$error_response")
                    # Collect detailed error for JSON report - with safe parsing
                    error_message=$(echo "$curl_result" | jq -r '.error.message // .error // "Unknown error"' 2>/dev/null || echo "Parse error")
                    endpoint_response=$(echo "$curl_result" | jq -r '.error.data.endpoint_response // ""' 2>/dev/null || echo "")
                    unmarshaling_error=$(echo "$curl_result" | jq -r '.error.data.unmarshaling_error // ""' 2>/dev/null || echo "")
                    detailed_errors+=("{\"error_message\":\"$error_message\",\"endpoint_response\":\"$endpoint_response\",\"unmarshaling_error\":\"$unmarshaling_error\"}")
                else
                    echo "      ✅ Success"
                    successful_requests=$((successful_requests + 1))
                fi
            else
                echo "      ❌ Failed (connection/invalid JSON)"
                failed_requests=$((failed_requests + 1))
                # Create error object for connection/JSON parsing failures
                error_response='{"code":-32700,"message":"Connection error or invalid JSON response"}'
                error_responses+=("$error_response")
                detailed_errors+=("{\"error_message\":\"Connection error or invalid JSON response\",\"endpoint_response\":\"\",\"unmarshaling_error\":\"\"}")
            fi
            
            # Brief pause between requests
            if [ $request_count -lt $max_requests ]; then
                sleep $RETRY_SLEEP_DURATION
            fi
        done
        
        # Determine overall result - ALL requests must succeed
        if [ $successful_requests -eq $max_requests ]; then
            echo "  🎉 SUCCESS ($successful_requests/$max_requests requests succeeded)"
            test_result="✅"
            overall_status="success"
        else
            echo "  💔 FAILED ($successful_requests/$max_requests requests succeeded, $failed_requests failed)"
            test_result="❌"
            overall_status="failed"
        fi
        
        # If disqualified endpoints flag is set, query disqualified endpoints
        disqualified_response=""
        if [ "$QUERY_DISQUALIFIED" = true ]; then
            echo "    📊 Querying disqualified endpoints for $service..."
            
            # Construct URL and headers based on environment
            if [ "$USE_SUBDOMAIN" = true ]; then
                # Production: use subdomain format with required headers
                disqualified_url="https://${service}.rpc.grove.city/disqualified_endpoints"
                disqualified_response=$(curl -s "$disqualified_url" \
                    -H "Portal-Application-Id: $PORTAL_APP_ID" \
                    -H "Authorization: $API_KEY" 2>/dev/null)
            else
                # Local: use header format
                disqualified_response=$(curl -s "$BASE_DISQUALIFIED_URL" -H "Target-Service-Id: $service" 2>/dev/null)
            fi
            curl_exit_code=$?
            
            if [ $curl_exit_code -eq 0 ] && [ -n "$disqualified_response" ]; then
                # Check if response is valid JSON
                if echo "$disqualified_response" | jq empty 2>/dev/null; then
                    echo "    ✅ Successfully retrieved disqualified endpoints data"
                else
                    echo "    ❌ Invalid JSON from disqualified endpoints"
                    disqualified_response=""
                fi
            else
                echo "    💥 Disqualified endpoints call failed"
                disqualified_response=""
            fi
        fi
        
        # Create JSON entry for this service
        errors_json="[]"
        for error_detail in "${detailed_errors[@]}"; do
            errors_json=$(echo "$errors_json" | jq --argjson err "$error_detail" '. += [$err]')
        done
        
        # Build service JSON with optional disqualified endpoints response
        if [ "$QUERY_DISQUALIFIED" = true ] && [ -n "$disqualified_response" ]; then
            service_json=$(jq -n \
                --arg service_id "$service" \
                --arg suppliers "$count" \
                --arg test_result "$test_result" \
                --arg status "$overall_status" \
                --arg successful_requests "$successful_requests" \
                --arg failed_requests "$failed_requests" \
                --arg total_requests "$max_requests" \
                --argjson errors "$errors_json" \
                --argjson disqualified_response "$disqualified_response" \
                '{
                    service_id: $service_id,
                    suppliers: ($suppliers | tonumber),
                    test_result: $test_result,
                    status: $status,
                    successful_requests: ($successful_requests | tonumber),
                    failed_requests: ($failed_requests | tonumber),
                    total_requests: ($total_requests | tonumber),
                    errors: $errors,
                    disqualifed_endpoints_response: $disqualified_response
                }')
        else
            service_json=$(jq -n \
                --arg service_id "$service" \
                --arg suppliers "$count" \
                --arg test_result "$test_result" \
                --arg status "$overall_status" \
                --arg successful_requests "$successful_requests" \
                --arg failed_requests "$failed_requests" \
                --arg total_requests "$max_requests" \
                --argjson errors "$errors_json" \
                '{
                    service_id: $service_id,
                    suppliers: ($suppliers | tonumber),
                    test_result: $test_result,
                    status: $status,
                    successful_requests: ($successful_requests | tonumber),
                    failed_requests: ($failed_requests | tonumber),
                    total_requests: ($total_requests | tonumber),
                    errors: $errors
                }')
        fi
        
        # Add to appropriate category in JSON report based on test results
        if [ "$overall_status" = "success" ]; then
            jq --arg service "$service" --argjson data "$service_json" '.suppliers_passed[$service] = $data' "$TEMP_DIR/json_report.json" > "$TEMP_DIR/temp_report.json" && \
               mv "$TEMP_DIR/temp_report.json" "$TEMP_DIR/json_report.json"
        else
            jq --arg service "$service" --argjson data "$service_json" '.suppliers_failed[$service] = $data' "$TEMP_DIR/json_report.json" > "$TEMP_DIR/temp_report.json" && \
               mv "$TEMP_DIR/temp_report.json" "$TEMP_DIR/json_report.json"
        fi
        
        # Keep track for console output
        all_results+=("$service,$count,$test_result")
        
        # Clear the error arrays for next service
        unset error_responses
        unset detailed_errors
    done
    
    # Add services with no suppliers to the no_suppliers array
    for service in "${SERVICES[@]}"; do
        # Check if this service is in our services_with_suppliers array
        found=false
        for item in "${services_with_suppliers[@]}"; do
            IFS=':' read -r service_with_suppliers count <<< "$item"
            if [ "$service" = "$service_with_suppliers" ]; then
                found=true
                break
            fi
        done
        
        # If not found in services_with_suppliers, add to no_suppliers
        if [ "$found" = false ]; then
            jq --arg service "$service" '.no_suppliers += [$service]' "$TEMP_DIR/json_report.json" > "$TEMP_DIR/temp_report.json" && \
               mv "$TEMP_DIR/temp_report.json" "$TEMP_DIR/json_report.json"
        fi
    done
    
    # Add summary to JSON report
    total_services=${#services_with_suppliers[@]}
    successful_services=$(jq '.suppliers_passed | length' "$TEMP_DIR/json_report.json")
    failed_services=$(jq '.suppliers_failed | length' "$TEMP_DIR/json_report.json")
    no_suppliers_count=$(jq '.no_suppliers | length' "$TEMP_DIR/json_report.json")
    
    summary_json=$(jq -n \
        --arg total "$total_services" \
        --arg successful "$successful_services" \
        --arg failed "$failed_services" \
        --arg no_suppliers "$no_suppliers_count" \
        --arg timestamp "$(date)" \
        --arg disqualified_flag "$QUERY_DISQUALIFIED" \
        '{
            total_services_with_suppliers: ($total | tonumber),
            successful_services: ($successful | tonumber),
            failed_services: ($failed | tonumber),
            no_suppliers_count: ($no_suppliers | tonumber),
            timestamp: $timestamp,
            includes_disqualified_endpoints: ($disqualified_flag == "true")
        }')
    
    jq --argjson summary "$summary_json" '.summary = $summary' "$TEMP_DIR/json_report.json" > "$TEMP_DIR/temp_report.json" && \
       mv "$TEMP_DIR/temp_report.json" "$TEMP_DIR/json_report.json"
    
    # Sort suppliers_passed and suppliers_failed by service_id, and sort no_suppliers array
    jq '.suppliers_passed = (.suppliers_passed | to_entries | sort_by(.key) | from_entries) | 
        .suppliers_failed = (.suppliers_failed | to_entries | sort_by(.key) | from_entries) |
        .no_suppliers = (.no_suppliers | sort)' "$TEMP_DIR/json_report.json" > "$JSON_REPORT_FILE"
    
    echo ""
    echo "=============================="
    echo "📋 FINAL REPORT"
    echo "=============================="
    echo "💾 Report saved to: $JSON_REPORT_FILE"
    
    if [ "$QUERY_DISQUALIFIED" = true ]; then
        echo "📊 Report includes disqualified endpoints data for each service"
    fi
    
    echo ""
    
    # Display as a nice table
    printf "%-20s | %-9s | %s\n" "SERVICE ID" "SUPPLIERS" "RESULT"
    printf "%-20s-+-%9s-+-%s\n" "--------------------" "---------" "-------"
    
    # Sort and display the results from the all_results array
    printf '%s\n' "${all_results[@]}" | sort -t, -k3,3r -k1,1 | while IFS=, read -r service count result; do
        printf "%-20s | %-9s | %s\n" "$service" "$count" "$result"
    done
    
    echo "=============================="
else
    echo ""
    echo "🚫 No services with suppliers found. Skipping tests."
    
    # Add all services to no_suppliers since none have suppliers
    no_suppliers_json="[]"
    for service in "${SERVICES[@]}"; do
        no_suppliers_json=$(echo "$no_suppliers_json" | jq --arg service "$service" '. += [$service]')
    done
    
    # Create JSON report with all services in no_suppliers
    summary_json=$(jq -n \
        --arg no_suppliers_count "${#SERVICES[@]}" \
        --arg timestamp "$(date)" \
        --arg disqualified_flag "$QUERY_DISQUALIFIED" \
        '{
            total_services_with_suppliers: 0,
            successful_services: 0,
            failed_services: 0,
            no_suppliers_count: ($no_suppliers_count | tonumber),
            timestamp: $timestamp,
            includes_disqualified_endpoints: ($disqualified_flag == "true")
        }')
    
    jq -n \
        --argjson no_suppliers "$no_suppliers_json" \
        --argjson summary "$summary_json" \
        '{
            suppliers_failed: {},
            suppliers_passed: {},
            no_suppliers: ($no_suppliers | sort),
            summary: $summary
        }' > "$JSON_REPORT_FILE"
    
    echo "💾 Empty report saved to: $JSON_REPORT_FILE"
fi
