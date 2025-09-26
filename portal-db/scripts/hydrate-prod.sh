#!/bin/bash

# PostgreSQL dump script for remote database
# This script creates a local pg_dump from a remote PostgreSQL database

set -e

# Configuration file path
ENV_FILE="$(dirname "$0")/../.env"
PG_DUMP_FILE="$(dirname "$0")/../pg_dump.sql"

# Check if .env file exists
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}❌ .env file not found at: $ENV_FILE${NC}"
    echo ""
    echo "Please create a .env file with all required variables."
    exit 1
fi

# Load environment variables from .env file
echo "📋 Loading configuration from .env file..."
set -a  # Automatically export all variables
source "$ENV_FILE"
set +a  # Turn off automatic export

# Validate required environment variables
REQUIRED_VARS=("REMOTE_HOST" "DATABASE" "USERNAME" "PASSWORD" "SSL_ROOT_CERT" "SSL_CERT" "SSL_KEY" "DOCKER_CONTAINER" "LOCAL_DATABASE" "LOCAL_DB_USER")

for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ]; then
        echo -e "${RED}❌ Required environment variable '$var' is not set in .env file${NC}"
        exit 1
    fi
done

echo -e "${GREEN}✅ All required environment variables loaded${NC}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🗄️  PostgreSQL Remote Database Dump${NC}"
echo "========================================"
echo -e "${BLUE}Remote Host:${NC} $REMOTE_HOST"
echo -e "${BLUE}Database:${NC} $DATABASE"
echo -e "${BLUE}Username:${NC} $USERNAME"
echo -e "${BLUE}Output File:${NC} $PG_DUMP_FILE"
echo ""

# Clean up any existing dump file
if [ -f "$PG_DUMP_FILE" ]; then
    echo -e "${YELLOW}🧹 Cleaning previous dump file...${NC}"
    rm -f "$PG_DUMP_FILE"
    echo -e "${GREEN}✅ Previous dump file removed${NC}"
fi

# Check if pg_dump is available
if ! command -v pg_dump >/dev/null 2>&1; then
    echo -e "${RED}❌ pg_dump is not installed${NC}"
    echo "   Install PostgreSQL client tools:"
    echo "   - Mac: brew install postgresql"
    echo "   - Ubuntu: sudo apt-get install postgresql-client"
    exit 1
fi

echo -e "${GREEN}✅ pg_dump is available: $(pg_dump --version)${NC}"

# Check if SSL certificates exist
if [ ! -f "$SSL_ROOT_CERT" ]; then
    echo -e "${RED}❌ SSL root certificate not found: $SSL_ROOT_CERT${NC}"
    exit 1
fi

if [ ! -f "$SSL_CERT" ]; then
    echo -e "${RED}❌ SSL client certificate not found: $SSL_CERT${NC}"
    exit 1
fi

if [ ! -f "$SSL_KEY" ]; then
    echo -e "${RED}❌ SSL client key not found: $SSL_KEY${NC}"
    exit 1
fi

echo -e "${GREEN}✅ SSL certificates found:${NC}"
echo "   Root CA: $SSL_ROOT_CERT"
echo "   Client Cert: $SSL_CERT"
echo "   Client Key: $SSL_KEY"

# Check and fix SSL key permissions (PostgreSQL requires 0600 or stricter)
KEY_PERMS=$(stat -f "%Lp" "$SSL_KEY" 2>/dev/null || stat -c "%a" "$SSL_KEY" 2>/dev/null)
if [ "$KEY_PERMS" != "600" ]; then
    echo -e "${YELLOW}⚠️  Fixing SSL key permissions (current: $KEY_PERMS, required: 600)${NC}"
    chmod 600 "$SSL_KEY"
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ SSL key permissions fixed${NC}"
    else
        echo -e "${RED}❌ Failed to fix SSL key permissions. Please run: chmod 600 $SSL_KEY${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✅ SSL key permissions correct (600)${NC}"
fi

# Create output directory if it doesn't exist
OUTPUT_DIR=$(dirname "$PG_DUMP_FILE")
mkdir -p "$OUTPUT_DIR"

# Set password environment variable to avoid prompting
export PGPASSWORD="$PASSWORD"

echo ""
echo -e "${YELLOW}🚀 Starting database dump...${NC}"
echo "   This may take several minutes depending on database size"

# Set SSL environment variables for PostgreSQL client
export PGSSLMODE="verify-ca"
export PGSSLCERT="$SSL_CERT"
export PGSSLKEY="$SSL_KEY"
export PGSSLROOTCERT="$SSL_ROOT_CERT"

# Perform the dump in plain SQL format - DATA ONLY from public schema only
echo -e "${BLUE}📥 Dumping database (data only - public schema)...${NC}"
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
    echo -e "${GREEN}✅ Database dump completed successfully!${NC}"
    
    # Display file information
    if [ -f "$PG_DUMP_FILE" ]; then
        FILE_SIZE=$(ls -lh "$PG_DUMP_FILE" | awk '{print $5}')
        echo -e "${BLUE}📊 Dump Information:${NC}"
        echo "   File: $PG_DUMP_FILE"
        echo "   Size: $FILE_SIZE"
        echo "   Format: Plain SQL - DATA ONLY (public schema)"
        echo "   Schema required: YES (assumes target schema exists)"
        echo "   Ready for auto-restore: Yes"
        echo ""
        echo -e "${BLUE}📚 Usage Instructions:${NC}"
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
    echo -e "${RED}❌ Database dump failed${NC}"
    echo "   Check the connection parameters and credentials"
    exit 1
fi

# Clear sensitive environment variables
unset PGPASSWORD
unset PGSSLMODE
unset PGSSLCERT
unset PGSSLKEY
unset PGSSLROOTCERT

echo ""
echo -e "${GREEN}🎉 Dump process completed!${NC}"

# ============================================================================
# AUTOMATIC DATABASE IMPORT
# ============================================================================

echo ""
echo -e "${BLUE}🚀 Starting automatic database import...${NC}"

# Check if dump file exists and has content
if [ ! -f "$PG_DUMP_FILE" ] || [ ! -s "$PG_DUMP_FILE" ]; then
    echo -e "${RED}❌ Dump file not found or empty, skipping import${NC}"
    exit 1
fi

# Check if Docker container is running
if ! docker ps --format "table {{.Names}}" | grep -q "^$DOCKER_CONTAINER$"; then
    echo -e "${RED}❌ Docker container '$DOCKER_CONTAINER' is not running${NC}"
    echo "   Please start PostgreSQL and PostgREST services first:"
    echo "   make postgrest-up"
    echo ""
    echo "   Or manually: docker compose up -d"
    exit 1
fi

echo -e "${BLUE}📥 Importing data to local PostgreSQL container...${NC}"

# Import data with suppressed output, only show result
if cat "$PG_DUMP_FILE" | docker exec -i "$DOCKER_CONTAINER" psql -U "$LOCAL_DB_USER" -d "$LOCAL_DATABASE" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Database import completed successfully!${NC}"
    
    # Show import summary
    RECORD_COUNT=$(docker exec "$DOCKER_CONTAINER" psql -U "$LOCAL_DB_USER" -d "$LOCAL_DATABASE" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';" 2>/dev/null | xargs)
    echo -e "${BLUE}📊 Import Summary:${NC}"
    echo "   Tables imported: $RECORD_COUNT public schema tables"
    echo "   Container: $DOCKER_CONTAINER"
    echo "   Database: $LOCAL_DATABASE"
    echo "   Access: localhost:5435"
    
else
    echo -e "${RED}❌ Database import failed${NC}"
    echo "   Check Docker container status and database connectivity"
    echo "   Manual import: cat $PG_DUMP_FILE | docker exec -i $DOCKER_CONTAINER psql -U $LOCAL_DB_USER -d $LOCAL_DATABASE"
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 Complete! Production data is now available locally.${NC}"
