#!/bin/bash

# PostgreSQL dump script for remote database
# This script creates a local pg_dump from a remote PostgreSQL database

set -e

# Source common utilities
source "$(dirname "$0")/lib/common.sh"

# Configuration file path
ENV_FILE="$(dirname "$0")/../.env"
PG_DUMP_FILE="$(dirname "$0")/../pg_dump.sql"

# Set default values for local Docker environment
DEFAULT_DOCKER_CONTAINER="portal-db"
DEFAULT_LOCAL_DATABASE="portal_db"
DEFAULT_LOCAL_DB_USER="postgres"

# Check if Docker container is running (check default first, then .env override)
CONTAINER_TO_CHECK="${DOCKER_CONTAINER:-$DEFAULT_DOCKER_CONTAINER}"
check_docker_container "$CONTAINER_TO_CHECK"

# Load and validate environment variables
REQUIRED_VARS=("REMOTE_HOST" "DATABASE" "USERNAME" "PASSWORD" "SSL_ROOT_CERT" "SSL_CERT" "SSL_KEY")
load_env_file "$ENV_FILE" "${REQUIRED_VARS[@]}"

# Apply defaults for local Docker environment variables if not set in .env
DOCKER_CONTAINER="${DOCKER_CONTAINER:-$DEFAULT_DOCKER_CONTAINER}"
LOCAL_DATABASE="${LOCAL_DATABASE:-$DEFAULT_LOCAL_DATABASE}"
LOCAL_DB_USER="${LOCAL_DB_USER:-$DEFAULT_LOCAL_DB_USER}"

print_status "$GREEN" "✅ Using Docker container: $DOCKER_CONTAINER"
print_status "$GREEN" "✅ Using local database: $LOCAL_DATABASE"
print_status "$GREEN" "✅ Using local DB user: $LOCAL_DB_USER"

print_status "$BLUE" "🗄️  PostgreSQL Remote Database Dump"
echo "========================================"
echo -e "${BLUE}Remote Host:${NC} $REMOTE_HOST"
echo -e "${BLUE}Database:${NC} $DATABASE"
echo -e "${BLUE}Username:${NC} $USERNAME"
echo -e "${BLUE}Output File:${NC} $PG_DUMP_FILE"
echo ""

# Clean up any existing dump file
if [ -f "$PG_DUMP_FILE" ]; then
    print_status "$YELLOW" "🧹 Cleaning previous dump file..."
    rm -f "$PG_DUMP_FILE"
    print_status "$GREEN" "✅ Previous dump file removed"
fi

# Check if pg_dump is available
check_command "pg_dump" "   Install PostgreSQL client tools:
   - Mac: brew install postgresql
   - Ubuntu: sudo apt-get install postgresql-client"

print_status "$GREEN" "✅ pg_dump is available: $(pg_dump --version)"

# Validate SSL certificates and fix permissions
validate_ssl_certs "$SSL_ROOT_CERT" "$SSL_CERT" "$SSL_KEY"

# Create output directory if it doesn't exist
OUTPUT_DIR=$(dirname "$PG_DUMP_FILE")
mkdir -p "$OUTPUT_DIR"

# Set password environment variable to avoid prompting
export PGPASSWORD="$PASSWORD"

echo ""
print_status "$YELLOW" "🚀 Starting database dump..."
echo "   This may take several minutes depending on database size"

# Set SSL environment variables for PostgreSQL client
export PGSSLMODE="verify-ca"
export PGSSLCERT="$SSL_CERT"
export PGSSLKEY="$SSL_KEY"
export PGSSLROOTCERT="$SSL_ROOT_CERT"

# Perform the dump in plain SQL format - DATA ONLY from public schema only
print_status "$BLUE" "📥 Dumping database (data only - public schema)..."
if pg_dump \
    --host="$REMOTE_HOST" \
    --port=5432 \
    --username="$USERNAME" \
    --dbname="$DATABASE" \
    --format=plain \
    --verbose \
    --no-password \
    --file="$PG_DUMP_FILE" \
    --schema=public \
    --exclude-table-data='audit.*' \
    --exclude-table-data='logs.*' \
    --no-owner \
    --no-privileges \
    --data-only \
    --inserts; then

    echo ""
    print_status "$GREEN" "✅ Database dump completed successfully!"

    # Display file information
    if [ -f "$PG_DUMP_FILE" ]; then
        FILE_SIZE=$(ls -lh "$PG_DUMP_FILE" | awk '{print $5}')
        print_status "$BLUE" "📊 Dump Information:"
        echo "   File: $PG_DUMP_FILE"
        echo "   Size: $FILE_SIZE"
        echo "   Format: Plain SQL - DATA ONLY (public schema)"
        echo "   Schema required: YES (assumes target schema exists)"
        echo "   Ready for auto-restore: Yes"
        echo ""
        print_status "$BLUE" "📚 Usage Instructions:"
        echo "   💡 Recommended: Use the Makefile target for easier execution:"
        echo "   make hydrate-prod"
        echo ""
        echo "   ⚠️  IMPORTANT: Target database schema must already exist!"
        echo "   Ensure PostgreSQL and PostgREST are running first: make postgrest-up"
        echo ""
        echo "   Manual restore (if needed):"
        echo "   psql --host=localhost --port=5435 --username=postgres --dbname=portal_db < $PG_DUMP_FILE"
    fi

else
    echo ""
    print_status "$RED" "❌ Database dump failed"
    echo "   Check the connection parameters and credentials"
    exit 1
fi

# Clear sensitive environment variables
unset PGPASSWORD PGSSLMODE PGSSLCERT PGSSLKEY PGSSLROOTCERT

echo ""
print_status "$GREEN" "🎉 Dump process completed!"

# ============================================================================
# AUTOMATIC DATABASE IMPORT
# ============================================================================

echo ""
print_status "$BLUE" "🚀 Starting automatic database import..."

# Check if dump file exists and has content
if [ ! -f "$PG_DUMP_FILE" ] || [ ! -s "$PG_DUMP_FILE" ]; then
    print_status "$RED" "❌ Dump file not found or empty, skipping import"
    exit 1
fi

print_status "$BLUE" "📥 Importing data to local PostgreSQL container..."

# Import data with suppressed output, only show result
if cat "$PG_DUMP_FILE" | docker exec -i "$DOCKER_CONTAINER" psql -U "$LOCAL_DB_USER" -d "$LOCAL_DATABASE" >/dev/null 2>&1; then
    print_status "$GREEN" "✅ Database import completed successfully!"

    # Show import summary
    RECORD_COUNT=$(docker exec "$DOCKER_CONTAINER" psql -U "$LOCAL_DB_USER" -d "$LOCAL_DATABASE" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';" 2>/dev/null | xargs)
    print_status "$BLUE" "📊 Import Summary:"
    echo "   Tables imported: $RECORD_COUNT public schema tables"
    echo "   Container: $DOCKER_CONTAINER"
    echo "   Database: $LOCAL_DATABASE"
    echo "   Access: localhost:5435"

else
    print_status "$RED" "❌ Database import failed"
    echo "   Check Docker container status and database connectivity"
    echo "   Manual import: cat $PG_DUMP_FILE | docker exec -i $DOCKER_CONTAINER psql -U $LOCAL_DB_USER -d $LOCAL_DATABASE"
    exit 1
fi

echo ""
print_status "$GREEN" "🎉 Complete! Production data is now available locally."
