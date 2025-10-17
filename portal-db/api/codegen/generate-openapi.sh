#!/bin/bash

# Generate OpenAPI specification from PostgREST
# This script fetches the OpenAPI spec and saves it to openapi/

set -e

# Color codes for terminal output
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
BOLD='\033[1m'
RESET='\033[0m'

# Configuration
POSTGREST_URL="${POSTGREST_URL:-http://localhost:3000}"
OUTPUT_DIR="../openapi"
OUTPUT_FILE="$OUTPUT_DIR/openapi.json"

echo -e "${BLUE}🔍 Generating OpenAPI specification from PostgREST...${RESET}"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Generate JWT token for authenticated access to get all endpoints
echo -e "${BLUE}🔑 Generating JWT token for authenticated OpenAPI spec...${RESET}"
JWT_TOKEN=$(cd ../scripts && ./postgrest-gen-jwt.sh --token-only portal_db_admin 2>/dev/null)

if [ -z "$JWT_TOKEN" ]; then
    echo -e "${YELLOW}⚠️  Could not generate JWT token, fetching public endpoints only...${RESET}"
    AUTH_HEADER=""
else
    echo -e "${GREEN}✅ JWT token generated, will fetch all endpoints (public + protected)${RESET}"
    AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"
fi

# Wait for PostgREST to be ready
echo -e "${BLUE}💡 Ensure you ran ${CYAN}make portal-db-up${RESET}"
echo -e "${CYAN}⏳ Waiting for PostgREST to be ready at ${BOLD}$POSTGREST_URL${RESET}${CYAN}...${RESET}"
max_attempts=30
attempt=1

while [ $attempt -le $max_attempts ]; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
        ${AUTH_HEADER:+-H "$AUTH_HEADER"} \
        "$POSTGREST_URL/" || echo "000")

    if [ "$STATUS" != "000" ]; then
        echo -e "${GREEN}✅ PostgREST is ready! (HTTP $STATUS)${RESET}"
        break
    fi

    if [ $attempt -eq $max_attempts ]; then
        echo -e "${RED}❌ PostgREST is not responding after $max_attempts attempts${RESET}"
        echo -e "   Make sure PostgREST is running: ${CYAN}docker compose up postgrest${RESET}"
        exit 1
    fi

    echo -e "   ${YELLOW}Attempt $attempt/$max_attempts - waiting 2 seconds...${RESET}"
    sleep 2
    ((attempt++))
done

# Fetch OpenAPI specification (Swagger 2.0 format from PostgREST)
echo -e "${BLUE}📥 Fetching OpenAPI specification...${RESET}"
SWAGGER_FILE="${OUTPUT_FILE%.json}-swagger.json"

if curl -s -f -H "Accept: application/openapi+json" ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$POSTGREST_URL/" -o "$SWAGGER_FILE"; then
    echo -e "${GREEN}✅ Swagger 2.0 specification fetched${RESET}"

    # Convert Swagger 2.0 to OpenAPI 3.x
    echo -e "${BLUE}🔄 Converting Swagger 2.0 to OpenAPI 3.x...${RESET}"

    # Check if swagger2openapi is available (PostgREST outputs Swagger 2.0, we need OpenAPI 3.x for SDK generators)
    if ! command -v swagger2openapi >/dev/null 2>&1; then
        echo -e "${BLUE}📦 Installing swagger2openapi converter...${RESET}"
        if command -v npm >/dev/null 2>&1; then
            npm install -g swagger2openapi
        else
            echo -e "${RED}❌ npm not found. Please install Node.js and npm first.${RESET}"
            echo "   - Mac: brew install node"
            echo "   - Or download from: https://nodejs.org/"
            exit 1
        fi
    fi

    if ! swagger2openapi "$SWAGGER_FILE" -o "$OUTPUT_FILE"; then
        echo -e "${RED}❌ Failed to convert Swagger 2.0 to OpenAPI 3.x${RESET}"
        exit 1
    fi

    # Fix format issues in the converted spec
    # PostgreSQL outputs custom format types that are invalid in OpenAPI (e.g., "character varying", "timestamp with time zone")
    # We need to remove all non-standard OpenAPI formats while keeping valid ones like int32, int64, float, double, date, date-time
    echo -e "${BLUE}🔧 Fixing invalid format fields...${RESET}"

    # Use a Python/Node script to only keep valid OpenAPI formats
    # Valid OpenAPI formats for strings: byte, binary, date, date-time, password, email, uuid, uri, hostname, ipv4, ipv6
    # Valid OpenAPI formats for numbers: int32, int64, float, double
    # Everything else should be removed

    if command -v python3 >/dev/null 2>&1; then
        python3 -c "
import json
import re

# Valid OpenAPI format values
VALID_FORMATS = {
    'int32', 'int64', 'float', 'double',  # number formats
    'byte', 'binary', 'date', 'date-time', 'password',  # string formats
    'email', 'uuid', 'uri', 'hostname', 'ipv4', 'ipv6'  # more string formats
}

with open('$OUTPUT_FILE', 'r') as f:
    spec = json.load(f)

def clean_formats(obj):
    if isinstance(obj, dict):
        # Remove format if it's not in the valid list
        if 'format' in obj and obj.get('format') not in VALID_FORMATS:
            del obj['format']
        # Recursively clean nested objects
        for value in obj.values():
            clean_formats(value)
    elif isinstance(obj, list):
        for item in obj:
            clean_formats(item)

clean_formats(spec)

with open('$OUTPUT_FILE', 'w') as f:
    json.dump(spec, f, indent=2)
"
    else
        # Fallback to sed if Python is not available
        sed -i.bak \
            -e 's/"format": "boolean",//g' \
            -e 's/"format": "integer",//g' \
            -e 's/"format": "character varying[^"]*",//g' \
            -e 's/"format": "timestamp[^"]*",//g' \
            -e 's/"format": "text",//g' \
            -e 's/"format": "jsonb",//g' \
            -e 's/"format": "json",//g' \
            -e 's/"format": "bytea",//g' \
            -e 's/"format": "public\.[^"]*",//g' \
            "$OUTPUT_FILE"
        rm -f "${OUTPUT_FILE}.bak"
    fi

    # Clean up temporary Swagger file
    rm -f "$SWAGGER_FILE"

    echo -e "${GREEN}✅ OpenAPI 3.x specification saved to: ${CYAN}$OUTPUT_FILE${RESET}"

    # Pretty print the JSON
    if command -v jq >/dev/null 2>&1; then
        echo -e "${BLUE}🎨 Pretty-printing JSON...${RESET}"
        jq '.' "$OUTPUT_FILE" > "${OUTPUT_FILE}.tmp" && mv "${OUTPUT_FILE}.tmp" "$OUTPUT_FILE"
    fi

    # Display some info about the generated spec
    echo ""
    echo -e "${BOLD}${BLUE}📊 OpenAPI Specification Summary:${RESET}"
    echo -e "   File size: ${CYAN}$(wc -c < "$OUTPUT_FILE" | tr -d ' ') bytes${RESET}"

    if command -v jq >/dev/null 2>&1; then
        echo -e "   OpenAPI version: ${CYAN}$(jq -r '.openapi // "unknown"' "$OUTPUT_FILE")${RESET}"
        echo -e "   API title: ${CYAN}$(jq -r '.info.title // "unknown"' "$OUTPUT_FILE")${RESET}"
        echo -e "   API version: ${CYAN}$(jq -r '.info.version // "unknown"' "$OUTPUT_FILE")${RESET}"
        echo -e "   Number of paths: ${CYAN}$(jq -r '.paths | length' "$OUTPUT_FILE")${RESET}"
        echo -e "   Number of schemas: ${CYAN}$(jq -r '.components.schemas | length' "$OUTPUT_FILE")${RESET}"

        # Log the actual paths that were retrieved
        echo ""
        echo -e "${BOLD}${BLUE}🔍 Retrieved API Paths:${RESET}"
        jq -r '.paths | keys[]' "$OUTPUT_FILE" | while read path; do
            echo -e "   ${CYAN}• ${RESET}$path"
        done
    fi

    echo ""
    echo -e "${BOLD}${BLUE}🌐 API Documentation:${RESET}"
    echo -e "   Raw OpenAPI: ${CYAN}$POSTGREST_URL/${RESET}"
    echo -e "   Swagger UI: ${CYAN}make postgrest-swagger-ui${RESET}"

else
    echo -e "${RED}❌ Failed to fetch OpenAPI specification from $POSTGREST_URL${RESET}"
    echo -e "   Make sure PostgREST is running and accessible"
    exit 1
fi

echo ""
echo -e "${GREEN}${BOLD}✨ OpenAPI specification generation completed successfully!${RESET}"
