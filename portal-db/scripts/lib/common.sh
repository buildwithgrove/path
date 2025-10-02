#!/bin/bash

# Common utilities for Portal DB scripts
# Source this file at the beginning of scripts with: source "$(dirname "$0")/lib/common.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Print colored status message
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Check if Docker container is running
check_docker_container() {
    local container=$1
    if ! docker ps --format "table {{.Names}}" | grep -q "^$container$"; then
        print_status "$RED" "❌ Docker container '$container' is not running"
        echo "   Please start PostgreSQL and PostgREST services first:"
        echo "   make portal-db-up"
        exit 1
    fi
}

# Validate file exists
validate_file_exists() {
    local file=$1
    local description=$2
    if [ ! -f "$file" ]; then
        print_status "$RED" "❌ $description not found: $file"
        exit 1
    fi
}

# Validate SSL certificates and fix permissions
validate_ssl_certs() {
    local ssl_root_cert=$1
    local ssl_cert=$2
    local ssl_key=$3

    local certs=(
        "$ssl_root_cert:SSL root certificate"
        "$ssl_cert:SSL client certificate"
        "$ssl_key:SSL client key"
    )

    for cert_info in "${certs[@]}"; do
        IFS=':' read -r cert_path cert_desc <<< "$cert_info"
        validate_file_exists "$cert_path" "$cert_desc"
    done

    print_status "$GREEN" "✅ SSL certificates found:"
    echo "   Root CA: $ssl_root_cert"
    echo "   Client Cert: $ssl_cert"
    echo "   Client Key: $ssl_key"

    # Check and fix SSL key permissions (PostgreSQL requires 0600 or stricter)
    local key_perms=$(stat -f "%Lp" "$ssl_key" 2>/dev/null || stat -c "%a" "$ssl_key" 2>/dev/null)
    if [ "$key_perms" != "600" ]; then
        print_status "$YELLOW" "⚠️  Fixing SSL key permissions (current: $key_perms, required: 600)"
        chmod 600 "$ssl_key"
        if [ $? -eq 0 ]; then
            print_status "$GREEN" "✅ SSL key permissions fixed"
        else
            print_status "$RED" "❌ Failed to fix SSL key permissions. Please run: chmod 600 $ssl_key"
            exit 1
        fi
    else
        print_status "$GREEN" "✅ SSL key permissions correct (600)"
    fi
}

# Load and validate environment file
load_env_file() {
    local env_file=$1
    shift
    local required_vars=("$@")

    if [ ! -f "$env_file" ]; then
        print_status "$RED" "❌ .env file not found at: $env_file"
        echo ""
        echo "Please create a .env file with all required variables."
        exit 1
    fi

    echo "📋 Loading configuration from .env file..."
    set -a
    source "$env_file"
    set +a

    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            print_status "$RED" "❌ Required environment variable '$var' is not set in .env file"
            exit 1
        fi
    done

    print_status "$GREEN" "✅ All required environment variables loaded"
}

# Validate database connection
validate_db_connection() {
    local db_connection_string=$1

    if [ -z "$db_connection_string" ]; then
        print_status "$RED" "❌ Error: DB_CONNECTION_STRING environment variable is required"
        print_status "$YELLOW" "💡 Expected format: postgresql://user:password@host:port/database"
        print_status "$YELLOW" "💡 For local development: postgresql://postgres:portal_password@localhost:5435/portal_db"
        exit 1
    fi

    print_status "$BLUE" "🔍 Testing database connection..."
    if ! psql "$db_connection_string" -c "SELECT 1;" > /dev/null 2>&1; then
        print_status "$RED" "❌ Error: Cannot connect to database"
        print_status "$YELLOW" "💡 Make sure the database is running: make portal-db-up"
        exit 1
    fi

    print_status "$GREEN" "✅ Database connection successful"
}

# Check if command exists
check_command() {
    local cmd=$1
    local install_msg=$2

    if ! command -v "$cmd" >/dev/null 2>&1; then
        print_status "$RED" "❌ $cmd is not installed"
        echo "$install_msg"
        exit 1
    fi
}
