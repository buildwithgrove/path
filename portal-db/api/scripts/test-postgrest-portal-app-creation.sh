#!/bin/bash

# 🧪 Test Script for Portal Application Creation
# This script exercises the PostgREST endpoints for inserting a portal
# application and its RBAC membership using the admin role.

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

# 🔍 Function to validate PostgREST is running
validate_postgrest() {
    print_status $BLUE "🔍 Checking if PostgREST is running..."
    if ! curl -s http://localhost:3000/ > /dev/null 2>&1; then
        print_status $RED "❌ Error: PostgREST is not running on localhost:3000"
        print_status $YELLOW "💡 Start the services first: make portal-db-up"
        exit 1
    fi
    print_status $GREEN "✅ PostgREST is running"
}

# 🔑 Function to generate JWT token
generate_jwt() {
    print_status $BLUE "🔑 Generating JWT token..."
    JWT_TOKEN=$(./postgrest-gen-jwt.sh --token-only portal_db_admin admin@example.com 2>/dev/null)
    if [ -z "$JWT_TOKEN" ]; then
        print_status $RED "❌ Error: Failed to generate JWT token"
        exit 1
    fi
    print_status $GREEN "✅ JWT token generated"
}

# 📱 Function to create portal application
create_portal_app() {
    print_status $BLUE "📱 Creating new portal application..."

    TIMESTAMP=$(date +%s)
    APP_NAME="Test App ${TIMESTAMP}"

    # Generate secret key and hash (hex encoded)
    SECRET_KEY=$(openssl rand -hex 32)
    SECRET_KEY_HASH=$(printf "%s" "$SECRET_KEY" | openssl dgst -sha256 | awk '{print $NF}' | tr -d '\n')

    CREATE_RESPONSE=$(curl -s -X POST http://localhost:3000/portal_applications \
        -H "Content-Type: application/json" \
        -H "Prefer: return=representation" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -d "{
            \"portal_account_id\": \"10000000-0000-0000-0000-000000000004\",
            \"portal_application_name\": \"$APP_NAME\",
            \"portal_application_description\": \"Test application created via automated test\",
            \"emoji\": \"test\",
            \"secret_key_hash\": \"$SECRET_KEY_HASH\",
            \"secret_key_required\": false
        }")

    if echo "$CREATE_RESPONSE" | jq -e '.[0]' >/dev/null 2>&1; then
        APP_ID=$(echo "$CREATE_RESPONSE" | jq -r '.[0].portal_application_id')
    else
        print_status $RED "❌ Error creating portal application:"
        echo "$CREATE_RESPONSE" | jq '.'
        exit 1
    fi

    if [ -z "$APP_ID" ] || [ "$APP_ID" = "null" ]; then
        print_status $RED "❌ Error: Could not extract application ID from response"
        echo "$CREATE_RESPONSE" | jq '.'
        exit 1
    fi

    print_status $GREEN "✅ Portal application created successfully!"
    print_status $CYAN "   📱 Application ID: $APP_ID"
    print_status $CYAN "   🏷️  Application Name: $APP_NAME"
    print_status $CYAN "   🔑 Secret Key (store securely!): $SECRET_KEY"

    echo ""
    print_status $PURPLE "📋 Full Create Response:"
    echo "$CREATE_RESPONSE" | jq '.'
}

# 👥 Function to grant user access via RBAC entry
create_portal_app_rbac() {
    print_status $BLUE "👥 Assigning user to portal application..."

    RBAC_RESPONSE=$(curl -s -X POST http://localhost:3000/portal_application_rbac \
        -H "Content-Type: application/json" \
        -H "Prefer: return=representation" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -d "{
            \"portal_application_id\": \"$APP_ID\",
            \"portal_user_id\": \"30000000-0000-0000-0000-000000000001\"
        }")

    if echo "$RBAC_RESPONSE" | jq -e '.[0]' >/dev/null 2>&1; then
        print_status $GREEN "✅ Portal application RBAC entry created"
    else
        print_status $RED "❌ Error creating RBAC entry:"
        echo "$RBAC_RESPONSE" | jq '.'
        exit 1
    fi
}

# 🔍 Function to retrieve portal application
retrieve_portal_app() {
    print_status $BLUE "🔍 Retrieving portal application by ID..."

    RETRIEVE_RESPONSE=$(curl -s -X GET \
        "http://localhost:3000/portal_applications?portal_application_id=eq.$APP_ID" \
        -H "Authorization: Bearer $JWT_TOKEN")

    # Check if we got results
    APP_COUNT=$(echo "$RETRIEVE_RESPONSE" | jq '. | length')
    if [ "$APP_COUNT" -eq "0" ]; then
        print_status $RED "❌ Error: Application not found in database"
        exit 1
    fi

    print_status $GREEN "✅ Portal application retrieved successfully!"

    echo ""
    print_status $PURPLE "📋 Full Retrieve Response:"
    echo "$RETRIEVE_RESPONSE" | jq '.'

    # Extract key fields for comparison
    RETRIEVED_NAME=$(echo "$RETRIEVE_RESPONSE" | jq -r '.[0].portal_application_name')
    RETRIEVED_DESCRIPTION=$(echo "$RETRIEVE_RESPONSE" | jq -r '.[0].portal_application_description')
    RETRIEVED_EMOJI=$(echo "$RETRIEVE_RESPONSE" | jq -r '.[0].emoji')

    print_status $CYAN "   🏷️  Retrieved Name: $RETRIEVED_NAME"
    print_status $CYAN "   📝 Retrieved Description: $RETRIEVED_DESCRIPTION"
    print_status $CYAN "   😊 Retrieved Emoji: $RETRIEVED_EMOJI"
}

# 🔍 Function to test retrieval by name
retrieve_by_name() {
    print_status $BLUE "🔍 Testing retrieval by application name..."

    # URL encode the app name
    ENCODED_NAME=$(echo "$APP_NAME" | sed 's/ /%20/g')

    RETRIEVE_BY_NAME_RESPONSE=$(curl -s -X GET \
        "http://localhost:3000/portal_applications?portal_application_name=eq.$ENCODED_NAME" \
        -H "Authorization: Bearer $JWT_TOKEN")

    APP_COUNT_BY_NAME=$(echo "$RETRIEVE_BY_NAME_RESPONSE" | jq '. | length')
    if [ "$APP_COUNT_BY_NAME" -eq "0" ]; then
        print_status $RED "❌ Error: Application not found by name"
        exit 1
    fi

    print_status $GREEN "✅ Portal application found by name!"
    print_status $CYAN "   Found $APP_COUNT_BY_NAME application(s) with name: $APP_NAME"
}

# 📊 Function to show test summary
show_summary() {
    echo ""
    print_status $PURPLE "======================================"
    print_status $PURPLE "🎉 TEST SUMMARY"
    print_status $PURPLE "======================================"
    print_status $GREEN "✅ PostgREST API connectivity"
    print_status $GREEN "✅ JWT token generation"
    print_status $GREEN "✅ Portal application creation"
    print_status $GREEN "✅ Portal application retrieval by ID"
    print_status $GREEN "✅ Portal application retrieval by name"
    print_status $PURPLE "======================================"
    print_status $CYAN "💡 Created application ID: $APP_ID"
    print_status $CYAN "💡 Application name: $APP_NAME"
    print_status $CYAN "💡 Secret key: $SECRET_KEY"
    print_status $PURPLE "======================================"
}

# 🎯 Main execution
main() {
    print_status $PURPLE "🧪 Portal Application Creation Test"
    print_status $PURPLE "===================================="
    echo ""

    # Run all test steps
    validate_postgrest
    generate_jwt
    create_portal_app
    create_portal_app_rbac
    retrieve_portal_app
    retrieve_by_name
    show_summary

    print_status $GREEN "🎉 All tests passed successfully!"
}

# 🏁 Execute main function
main "$@"
