#!/bin/bash

# ============================================================================
# JWT Token Generator for PostgREST (Shell Script Version)
# ============================================================================
# Generates JWT tokens for PostgREST authentication using shell commands
# Following PostgREST Tutorial: https://docs.postgrest.org/en/v13/tutorials/tut1.html
#
# Dependencies:
#   - openssl (for HMAC-SHA256 signing)
#   - base64 (for encoding)
#   - jq (for JSON processing)
#
# Usage:
#   ./gen-jwt.sh                           # Generate token for 'authenticated' role
#   ./gen-jwt.sh anon                      # Generate token for 'anon' role
#   ./gen-jwt.sh authenticated user@email  # Generate token with specific email

set -e  # Exit on any error

# JWT secret from postgrest.conf (must match exactly)
JWT_SECRET="supersecretjwtsecretforlocaldevelopment123456789"

# Get command line arguments
ROLE="${1:-authenticated}"
EMAIL="${2:-john@doe.com}"

# Calculate expiration (1 hour from now)
EXP=$(date -d '+1 hour' +%s 2>/dev/null || date -v+1H +%s 2>/dev/null || echo $(($(date +%s) + 3600)))

# ============================================================================
# JWT Generation Functions
# ============================================================================

# Base64 URL encoding (removes padding and makes URL-safe)
base64url_encode() {
    base64 -w 0 2>/dev/null | tr '+/' '-_' | tr -d '=' || base64 | tr '+/' '-_' | tr -d '='
}

# Base64 URL decoding (adds padding and decodes)
base64url_decode() {
    local input="$1"
    # Add padding if needed
    case $((${#input} % 4)) in
        2) input="${input}==" ;;
        3) input="${input}=" ;;
    esac
    echo "$input" | tr '_-' '/+' | base64 -d 2>/dev/null || echo "$input" | tr '_-' '/+' | base64 -D 2>/dev/null
}

# Create JWT header
create_header() {
    echo -n '{"alg":"HS256","typ":"JWT"}' | base64url_encode
}

# Create JWT payload
create_payload() {
    local role="$1"
    local email="$2"
    local exp="$3"
    
    # Create JSON payload
    echo -n "{\"role\":\"$role\",\"email\":\"$email\",\"exp\":$exp}" | base64url_encode
}

# Create JWT signature using HMAC-SHA256
create_signature() {
    local data="$1"
    local secret="$2"
    
    echo -n "$data" | openssl dgst -sha256 -hmac "$secret" -binary | base64url_encode
}

# ============================================================================
# Generate JWT Token
# ============================================================================

echo "🔑 Generating JWT Token with Shell Script ✨"
echo "🔐═══════════════════════════════════════════════════🔐"

# Create header and payload
echo "📝 Creating JWT components..."
HEADER=$(create_header)
PAYLOAD=$(create_payload "$ROLE" "$EMAIL" "$EXP")

# Create signature data (header.payload)
SIGNATURE_DATA="$HEADER.$PAYLOAD"

# Generate signature
echo "🔏 Signing token with HMAC-SHA256..."
SIGNATURE=$(create_signature "$SIGNATURE_DATA" "$JWT_SECRET")

# Complete JWT token
JWT_TOKEN="$SIGNATURE_DATA.$SIGNATURE"

# ============================================================================
# Display Results
# ============================================================================

echo "✅ JWT Token Generated Successfully!"
echo ""
echo "👤 Role: $ROLE"
echo "📧 Email: $EMAIL"
echo "⏰ Expires: $(date -d @$EXP 2>/dev/null || date -r $EXP 2>/dev/null || echo "Unix timestamp: $EXP")"
echo ""
echo "🎟️  Token:"
echo "$JWT_TOKEN"
echo ""
echo "🔐═══════════════════════════════════════════════════🔐"
echo "🚀 Usage Example:"
echo "curl http://localhost:3000/rpc/me \\"
echo "  -H \"Authorization: Bearer $JWT_TOKEN\" \\"
echo "  -H \"Content-Type: application/json\""
echo "🔐═══════════════════════════════════════════════════🔐"

# ============================================================================
# Verification and Debugging
# ============================================================================

echo ""
echo "🔍 Token Payload (decoded for verification):"

# Decode the payload using our custom base64url_decode function
DECODED_PAYLOAD=$(base64url_decode "$PAYLOAD")

if command -v jq >/dev/null 2>&1; then
    # Pretty print with jq if available
    echo "$DECODED_PAYLOAD" | jq . 2>/dev/null || echo "❌ Could not parse JSON payload"
else
    # Basic JSON formatting without jq
    echo "$DECODED_PAYLOAD" | sed 's/,/,\n  /g' | sed 's/{/{\n  /' | sed 's/}/\n}/' || echo "❌ Could not decode payload"
    echo ""
    echo "💡 Install 'jq' for better JSON formatting: brew install jq"
fi

# Export token for use by other scripts
export JWT_TOKEN
echo ""
echo "💾 Token exported as \$JWT_TOKEN environment variable for use in other scripts! 🎯"

echo ""
echo "🎉 Happy testing with PostgREST! 🚀"
