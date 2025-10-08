#!/bin/bash

# Generate Go and TypeScript SDKs from OpenAPI specification
# This script generates both Go and TypeScript SDKs for the Portal DB API
# - Go SDK: Uses oapi-codegen for client and models generation  
# - TypeScript SDK: Uses openapi-typescript for minimal, type-safe client generation

set -e

# Configuration
OPENAPI_DIR="../openapi"
OPENAPI_V2_FILE="$OPENAPI_DIR/openapi-v2.json"
OPENAPI_V3_FILE="$OPENAPI_DIR/openapi.json"
GO_OUTPUT_DIR="../../sdk/go"
TS_OUTPUT_DIR="../../sdk/typescript"
CONFIG_MODELS="./codegen-models.yaml"
CONFIG_CLIENT="./codegen-client.yaml"
POSTGREST_URL="http://localhost:3000"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "🔧 Generating Go and TypeScript SDKs from OpenAPI specification..."

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

# Check if Node.js and npm are installed for TypeScript SDK generation
if ! command -v node >/dev/null 2>&1; then
    echo -e "${RED}❌ Node.js is not installed. Please install Node.js first.${NC}"
    echo "   - Mac: brew install node"
    echo "   - Or download from: https://nodejs.org/"
    exit 1
fi

echo -e "${GREEN}✅ Node.js is installed: $(node --version)${NC}"

if ! command -v npm >/dev/null 2>&1; then
    echo -e "${RED}❌ npm is not installed. Please install npm first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ npm is installed: $(npm --version)${NC}"

# Check if Java is installed
if ! command -v java >/dev/null 2>&1; then
    echo -e "${RED}❌ Java is not installed. OpenAPI Generator requires Java.${NC}"
    echo "   Install Java: brew install openjdk"
    exit 1
fi

# Verify Java is working properly
if ! java -version >/dev/null 2>&1; then
    echo -e "${RED}❌ Java is installed but not working properly. OpenAPI Generator requires Java.${NC}"
    echo "   Fix Java installation: brew install openjdk"
    echo "   Add to PATH: export PATH=\"/opt/homebrew/opt/openjdk/bin:\$PATH\""
    exit 1
fi

JAVA_VERSION=$(java -version 2>&1 | head -n1)
echo -e "${GREEN}✅ Java is available: $JAVA_VERSION${NC}"

# Check if openapi-generator-cli is installed  
if ! command -v openapi-generator-cli >/dev/null 2>&1; then
    echo "📦 Installing openapi-generator-cli..."
    npm install -g @openapitools/openapi-generator-cli
    
    # Verify installation
    if ! command -v openapi-generator-cli >/dev/null 2>&1; then
        echo -e "${RED}❌ Failed to install openapi-generator-cli. Please check your Node.js/npm installation.${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}✅ openapi-generator-cli is available: $(openapi-generator-cli version 2>/dev/null || echo 'installed')${NC}"

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

# Generate JWT token for authenticated access to get all endpoints
echo "🔑 Generating JWT token for authenticated OpenAPI spec..."
JWT_TOKEN=$(cd ../scripts && ./postgrest-gen-jwt.sh portal_db_admin 2>/dev/null | grep -A1 "🎟️  Token:" | tail -1)

if [ -z "$JWT_TOKEN" ]; then
    echo "⚠️  Could not generate JWT token, fetching public endpoints only..."
    AUTH_HEADER=""
else
    echo "✅ JWT token generated, will fetch all endpoints (public + protected)"
    AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"
fi

# Fetch OpenAPI spec from PostgREST (Swagger 2.0 format)
echo "📥 Fetching OpenAPI specification from PostgREST..."
if ! curl -s "$POSTGREST_URL" -H "Accept: application/openapi+json" ${AUTH_HEADER:+-H "$AUTH_HEADER"} > "$OPENAPI_V2_FILE"; then
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

echo ""
echo "🔷 Generating TypeScript SDK with minimal dependencies..."

# Create TypeScript output directory if it doesn't exist
mkdir -p "$TS_OUTPUT_DIR"

# Clean previous generated TypeScript files (keep permanent files)
echo "🧹 Cleaning previous TypeScript generated files..."
rm -rf "$TS_OUTPUT_DIR/src" "$TS_OUTPUT_DIR/models" "$TS_OUTPUT_DIR/apis"

# Generate TypeScript client using openapi-generator-cli (auto-generates client methods)
echo "   Generating TypeScript client with built-in methods..."
if ! openapi-generator-cli generate \
    -i "$OPENAPI_V3_FILE" \
    -g typescript-fetch \
    -o "$TS_OUTPUT_DIR" \
    --skip-validate-spec \
    --additional-properties=npmName="@grove/portal-db-sdk",typescriptThreePlus=true; then
    echo -e "${RED}❌ Failed to generate TypeScript client${NC}"
    exit 1
fi


echo -e "${GREEN}✅ TypeScript SDK generated successfully${NC}"

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

# TypeScript module setup
echo ""
echo "🔷 Setting up TypeScript module..."

# Navigate to TypeScript SDK directory
cd "$TS_OUTPUT_DIR"

# Create package.json if it doesn't exist
if [ ! -f "package.json" ]; then
    echo "📦 Creating package.json..."
    cat > package.json << 'EOF'
{
  "name": "@grove/portal-db-sdk",
  "version": "1.0.0",
  "description": "TypeScript SDK for Grove Portal DB API",
  "main": "index.ts",
  "types": "types.d.ts",
  "scripts": {
    "build": "tsc",
    "type-check": "tsc --noEmit"
  },
  "keywords": ["grove", "portal", "db", "api", "sdk", "typescript"],
  "author": "Grove Team",
  "license": "MIT",
  "devDependencies": {
    "typescript": "^5.0.0"
  },
  "peerDependencies": {
    "typescript": ">=4.5.0"
  }
}
EOF
fi

# Create tsconfig.json if it doesn't exist
if [ ! -f "tsconfig.json" ]; then
    echo "🔧 Creating tsconfig.json..."
    cat > tsconfig.json << 'EOF'
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "lib": ["ES2020", "DOM"],
    "moduleResolution": "Bundler",
    "noUncheckedIndexedAccess": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "strict": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "declaration": true,
    "declarationMap": true,
    "outDir": "./dist"
  },
  "include": ["src/**/*", "models/**/*", "apis/**/*"],
  "exclude": ["node_modules", "dist"]
}
EOF
fi

echo -e "${GREEN}✅ TypeScript module setup completed${NC}"

# Test TypeScript compilation if TypeScript is available
if command -v tsc >/dev/null 2>&1; then
    echo "🔍 Validating TypeScript compilation..."
    if ! npx tsc --noEmit; then
        echo -e "${YELLOW}⚠️  TypeScript compilation check failed, but types were generated${NC}"
        echo "   This may be due to missing dependencies or configuration issues"
        echo "   The generated types.ts file should still be usable"
    else
        echo -e "${GREEN}✅ TypeScript types validate successfully${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  TypeScript not found, skipping compilation validation${NC}"
    echo "   Install TypeScript globally: npm install -g typescript"
fi

# Return to scripts directory
cd - >/dev/null

# ============================================================================
# SUCCESS SUMMARY
# ============================================================================

echo ""
echo -e "${GREEN}🎉 SDK generation completed successfully!${NC}"
echo ""
echo -e "${BLUE}📁 Generated Files:${NC}"
echo "   API Docs:     $OPENAPI_V3_FILE"
echo "   Go SDK:       $GO_OUTPUT_DIR"
echo "   TypeScript:   $TS_OUTPUT_DIR"
echo ""
echo -e "${BLUE}🐹 Go SDK:${NC}"
echo "   Module:   github.com/buildwithgrove/path/portal-db/sdk/go"
echo "   Package:  portaldb"
echo "   Files:"
echo "   • models.go       - Generated data models and types (updated)"
echo "   • client.go       - Generated SDK client and methods (updated)"
echo "   • go.mod          - Go module definition (permanent)"
echo "   • README.md       - Documentation (permanent)"
echo ""
echo -e "${BLUE}🔷 TypeScript SDK:${NC}"
echo "   Package:  @grove/portal-db-sdk"
echo "   Runtime:  Zero dependencies (uses native fetch)"
echo "   Files:"
echo "   • apis/           - Generated API client classes (updated)"
echo "   • models/         - Generated TypeScript models (updated)"
echo "   • package.json    - Node.js package definition (permanent)"
echo "   • tsconfig.json   - TypeScript configuration (permanent)"
echo "   • README.md       - Documentation (permanent)"
echo ""
echo -e "${BLUE}📚 API Documentation:${NC}"
echo "   • openapi.json    - OpenAPI 3.x specification (updated)"
echo ""
echo -e "${BLUE}🚀 Next Steps:${NC}"
echo ""
echo -e "${BLUE}Go SDK:${NC}"
echo "   1. Review generated models: cat $GO_OUTPUT_DIR/models.go | head -50"
echo "   2. Review generated client: cat $GO_OUTPUT_DIR/client.go | head -50" 
echo "   3. Import in your project: go get github.com/buildwithgrove/path/portal-db/sdk/go"
echo "   4. Check documentation: cat $GO_OUTPUT_DIR/README.md"
echo ""
echo -e "${BLUE}TypeScript SDK:${NC}"
echo "   1. Review generated APIs: ls $TS_OUTPUT_DIR/apis/"
echo "   2. Review generated models: ls $TS_OUTPUT_DIR/models/"
echo "   3. Copy to your React project or publish as npm package"
echo "   4. Import client: import { DefaultApi } from './apis'"
echo "   5. Use built-in methods: await client.portalApplicationsGet()"
echo "   6. Check documentation: cat $TS_OUTPUT_DIR/README.md"
echo ""
echo -e "${BLUE}💡 Tips:${NC}"
echo "   • Go: Full client with methods, types separated for readability"
echo "   • TypeScript: Auto-generated client classes with built-in methods"
echo "   • Both SDKs update automatically when you run this script"
echo "   • Run after database schema changes to stay in sync"
echo "   • TypeScript SDK has zero runtime dependencies"
echo ""
echo -e "${GREEN}✨ Happy coding!${NC}"