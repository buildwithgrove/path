#!/bin/bash

# Generate Go SDK from OpenAPI specification using oapi-codegen
# This script generates a Go client SDK for the Portal DB API

set -e

# Configuration
OPENAPI_DIR="../openapi"
OPENAPI_V2_FILE="$OPENAPI_DIR/openapi-v2.json"
OPENAPI_V3_FILE="$OPENAPI_DIR/openapi.json"
GO_OUTPUT_DIR="../../sdk/go"
CONFIG_MODELS="./codegen-models.yaml"
CONFIG_CLIENT="./codegen-client.yaml"
POSTGREST_URL="http://localhost:3000"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "🔧 Generating Go SDK from OpenAPI specification using oapi-codegen..."

# ============================================================================
# PHASE 1: ENVIRONMENT VALIDATION
# ============================================================================

echo ""
echo -e "${BLUE}📋 Phase 1: Environment Validation${NC}"

# Check if Go is installed
if ! command -v go >/dev/null 2>&1; then
    echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
    echo "   - Mac: brew install go"
    echo "   - Or download from: https://golang.org/"
    exit 1
fi

echo -e "${GREEN}✅ Go is installed: $(go version)${NC}"

# Check if oapi-codegen is installed
if ! command -v oapi-codegen >/dev/null 2>&1; then
    echo "📦 Installing oapi-codegen..."
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
    
    # Verify installation
    if ! command -v oapi-codegen >/dev/null 2>&1; then
        echo -e "${RED}❌ Failed to install oapi-codegen. Please check your Go installation.${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}✅ oapi-codegen is available: $(oapi-codegen -version 2>/dev/null || echo 'installed')${NC}"

# Check if PostgREST is running
echo "🌐 Checking PostgREST availability..."
if ! curl -s --connect-timeout 5 "$POSTGREST_URL" >/dev/null 2>&1; then
    echo -e "${RED}❌ PostgREST is not accessible at $POSTGREST_URL${NC}"
    echo "   Please ensure PostgREST is running:"
    echo "   cd .. && docker compose up -d"
    echo "   cd api && docker compose up -d"
    exit 1
fi

echo -e "${GREEN}✅ PostgREST is accessible at $POSTGREST_URL${NC}"

# Check if configuration files exist
for config_file in "$CONFIG_MODELS" "$CONFIG_CLIENT"; do
    if [ ! -f "$config_file" ]; then
        echo -e "${RED}❌ Configuration file not found: $config_file${NC}"
        echo "   This should have been created as a permanent file."
        exit 1
    fi
done

echo -e "${GREEN}✅ Configuration files found: models, client${NC}"

# ============================================================================
# PHASE 2: SPEC RETRIEVAL & CONVERSION
# ============================================================================

echo ""
echo -e "${BLUE}📋 Phase 2: Spec Retrieval & Conversion${NC}"

# Create openapi directory if it doesn't exist
mkdir -p "$OPENAPI_DIR"

# Clean any existing files to start fresh
echo "🧹 Cleaning previous OpenAPI files..."
rm -f "$OPENAPI_V2_FILE" "$OPENAPI_V3_FILE"

# Fetch OpenAPI spec from PostgREST (Swagger 2.0 format)
echo "📥 Fetching OpenAPI specification from PostgREST..."
if ! curl -s "$POSTGREST_URL" -H "Accept: application/json" > "$OPENAPI_V2_FILE"; then
    echo -e "${RED}❌ Failed to fetch OpenAPI specification from $POSTGREST_URL${NC}"
    exit 1
fi

# Verify the file was created and has content
if [ ! -f "$OPENAPI_V2_FILE" ] || [ ! -s "$OPENAPI_V2_FILE" ]; then
    echo -e "${RED}❌ OpenAPI specification file is empty or missing${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Swagger 2.0 specification saved to: $OPENAPI_V2_FILE${NC}"

# Convert Swagger 2.0 to OpenAPI 3.x
echo "🔄 Converting Swagger 2.0 to OpenAPI 3.x..."

# Check if swagger2openapi is available
if ! command -v swagger2openapi >/dev/null 2>&1; then
    echo "📦 Installing swagger2openapi converter..."
    if command -v npm >/dev/null 2>&1; then
        npm install -g swagger2openapi
    else
        echo -e "${RED}❌ npm not found. Please install Node.js and npm first.${NC}"
        echo "   - Mac: brew install node"
        echo "   - Or download from: https://nodejs.org/"
        exit 1
    fi
fi

if ! swagger2openapi "$OPENAPI_V2_FILE" -o "$OPENAPI_V3_FILE"; then
    echo -e "${RED}❌ Failed to convert Swagger 2.0 to OpenAPI 3.x${NC}"
    exit 1
fi

# Fix boolean format issues in the converted spec (in place)
echo "🔧 Fixing boolean format issues..."
sed -i.bak 's/"format": "boolean",//g' "$OPENAPI_V3_FILE"
rm -f "${OPENAPI_V3_FILE}.bak"

# Remove the temporary Swagger 2.0 file since we only need the OpenAPI 3.x version
echo "🧹 Cleaning temporary Swagger 2.0 file..."
rm -f "$OPENAPI_V2_FILE"

echo -e "${GREEN}✅ OpenAPI 3.x specification ready: $OPENAPI_V3_FILE${NC}"

# ============================================================================
# PHASE 3: SDK GENERATION
# ============================================================================

echo ""
echo -e "${BLUE}📋 Phase 3: SDK Generation${NC}"

echo "🐹 Generating Go SDK in separate files for better readability..."

# Clean previous generated files
echo "🧹 Cleaning previous generated files..."
rm -f "$GO_OUTPUT_DIR/models.go" "$GO_OUTPUT_DIR/client.go"

# Generate models file (data types and structures)
echo "   Generating models.go..."
if ! oapi-codegen -config "$CONFIG_MODELS" "$OPENAPI_V3_FILE"; then
    echo -e "${RED}❌ Failed to generate models${NC}"
    exit 1
fi

# Generate client file (API client and methods)
echo "   Generating client.go..."
if ! oapi-codegen -config "$CONFIG_CLIENT" "$OPENAPI_V3_FILE"; then
    echo -e "${RED}❌ Failed to generate client${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Go SDK generated successfully in separate files${NC}"

# ============================================================================
# PHASE 4: MODULE SETUP
# ============================================================================

echo ""
echo -e "${BLUE}📋 Phase 4: Module Setup${NC}"

# Navigate to SDK directory for module setup
cd "$GO_OUTPUT_DIR"

# Run go mod tidy to resolve dependencies
echo "🔧 Resolving dependencies..."
if ! go mod tidy; then
    echo -e "${RED}❌ Failed to resolve Go dependencies${NC}"
    cd - >/dev/null
    exit 1
fi

echo -e "${GREEN}✅ Go dependencies resolved${NC}"

# Test compilation
echo "🔍 Validating generated code compilation..."
if ! go build ./...; then
    echo -e "${RED}❌ Generated code does not compile${NC}"
    cd - >/dev/null
    exit 1
fi

echo -e "${GREEN}✅ Generated code compiles successfully${NC}"

# Return to scripts directory
cd - >/dev/null

# ============================================================================
# SUCCESS SUMMARY
# ============================================================================

echo ""
echo -e "${GREEN}🎉 SDK generation completed successfully!${NC}"
echo ""
echo -e "${BLUE}📁 Generated Files:${NC}"
echo "   API Docs: $OPENAPI_V3_FILE"
echo "   SDK:      $GO_OUTPUT_DIR"
echo "   Module:   github.com/grove/path/portal-db/sdk/go"
echo "   Package:  portaldb"
echo ""
echo -e "${BLUE}📚 SDK Files:${NC}"
echo "   • models.go       - Generated data models and types (updated)"
echo "   • client.go       - Generated SDK client and methods (updated)"
echo "   • go.mod          - Go module definition (permanent)"
echo "   • README.md       - Documentation (permanent)"
echo ""
echo -e "${BLUE}📚 API Documentation:${NC}"
echo "   • openapi.json    - OpenAPI 3.x specification (updated)"
echo ""
echo -e "${BLUE}🚀 Next steps:${NC}"
echo "   1. Review the generated models: cat $GO_OUTPUT_DIR/models.go | head -50"
echo "   2. Review the generated client: cat $GO_OUTPUT_DIR/client.go | head -50"
echo "   3. Import in your project: go get github.com/grove/path/portal-db/sdk/go"
echo "   4. Check the README: cat $GO_OUTPUT_DIR/README.md"
echo ""
echo -e "${BLUE}💡 Tips:${NC}"
echo "   • Generated files: models.go (data types), client.go (API methods)"
echo "   • Permanent files: go.mod, README.md"
echo "   • Better readability: types separated from client logic"
echo "   • Run this script after database schema changes"
echo ""
echo -e "${GREEN}✨ Happy coding!${NC}"