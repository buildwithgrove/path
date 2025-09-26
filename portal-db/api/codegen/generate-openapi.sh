#!/bin/bash

# Generate OpenAPI specification from PostgREST
# This script fetches the OpenAPI spec and saves it to openapi/

set -e

# Configuration
POSTGREST_URL="${POSTGREST_URL:-http://localhost:3000}"
OUTPUT_DIR="../openapi"
OUTPUT_FILE="$OUTPUT_DIR/openapi.json"

echo "🔍 Generating OpenAPI specification from PostgREST..."

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Wait for PostgREST to be ready
echo "⏳ Waiting for PostgREST to be ready at $POSTGREST_URL..."
max_attempts=30
attempt=1

while [ $attempt -le $max_attempts ]; do
    if curl -s -f "$POSTGREST_URL/" > /dev/null 2>&1; then
        echo "✅ PostgREST is ready!"
        break
    fi
    
    if [ $attempt -eq $max_attempts ]; then
        echo "❌ PostgREST is not responding after $max_attempts attempts"
        echo "   Make sure PostgREST is running: docker compose up postgrest"
        exit 1
    fi
    
    echo "   Attempt $attempt/$max_attempts - waiting 2 seconds..."
    sleep 2
    ((attempt++))
done

# Generate JWT token for authenticated access to get all endpoints
echo "🔑 Generating JWT token for authenticated OpenAPI spec..."
JWT_TOKEN=$(cd ../scripts && ./gen-jwt.sh authenticated 2>/dev/null | grep -A1 "🎟️  Token:" | tail -1)

if [ -z "$JWT_TOKEN" ]; then
    echo "⚠️  Could not generate JWT token, fetching public endpoints only..."
    AUTH_HEADER=""
else
    echo "✅ JWT token generated, will fetch all endpoints (public + protected)"
    AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"
fi

# Fetch OpenAPI specification
echo "📥 Fetching OpenAPI specification..."
if curl -s -f -H "Accept: application/openapi+json" ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$POSTGREST_URL/" > "$OUTPUT_FILE"; then
    echo "✅ OpenAPI specification saved to: $OUTPUT_FILE"
    
    # Pretty print the JSON
    if command -v jq >/dev/null 2>&1; then
        echo "🎨 Pretty-printing JSON..."
        jq '.' "$OUTPUT_FILE" > "${OUTPUT_FILE}.tmp" && mv "${OUTPUT_FILE}.tmp" "$OUTPUT_FILE"
    fi
    
    # Display some info about the generated spec
    echo ""
    echo "📊 OpenAPI Specification Summary:"
    echo "   File size: $(wc -c < "$OUTPUT_FILE" | tr -d ' ') bytes"
    
    if command -v jq >/dev/null 2>&1; then
        echo "   OpenAPI version: $(jq -r '.openapi // "unknown"' "$OUTPUT_FILE")"
        echo "   API title: $(jq -r '.info.title // "unknown"' "$OUTPUT_FILE")"
        echo "   API version: $(jq -r '.info.version // "unknown"' "$OUTPUT_FILE")"
        echo "   Number of paths: $(jq -r '.paths | length' "$OUTPUT_FILE")"
        echo "   Number of schemas: $(jq -r '.components.schemas | length' "$OUTPUT_FILE")"
        
        # Log the actual paths that were retrieved
        echo ""
        echo "🔍 Retrieved API Paths:"
        jq -r '.paths | keys[]' "$OUTPUT_FILE" | sed 's/^/   • /'
    fi
    
    echo ""
    echo "🌐 You can view the API documentation at:"
    echo "   Swagger UI: http://localhost:8080"
    echo "   Raw OpenAPI: $POSTGREST_URL/"
    
else
    echo "❌ Failed to fetch OpenAPI specification from $POSTGREST_URL"
    echo "   Make sure PostgREST is running and accessible"
    exit 1
fi

echo ""
echo "✨ OpenAPI specification generation completed successfully!"
