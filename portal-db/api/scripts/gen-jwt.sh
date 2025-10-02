#!/bin/bash

# ============================================================================
# JWT Token Generator for PostgREST
# ============================================================================
# Generates JWT tokens for PostgREST authentication using HMAC-SHA256 signing.
#
# Dependencies: openssl, base64
# Reference: https://docs.postgrest.org/en/v13/tutorials/tut1.html
#
# Usage:
#   ./gen-jwt.sh                           # authenticated role, default email
#   ./gen-jwt.sh anon                      # anon role
#   ./gen-jwt.sh authenticated user@email  # custom email

set -e

# JWT secret from postgrest.conf (must match exactly)
JWT_SECRET="supersecretjwtsecretforlocaldevelopment123456789"

# Parse arguments
ROLE="${1:-authenticated}"
EMAIL="${2:-john@doe.com}"

# Show help
if [[ "$ROLE" == "--help" || "$ROLE" == "-h" ]]; then
    sed -n '3,15p' "$0" | sed 's/^# //'
    exit 0
fi

# Calculate expiration (1 hour from now)
EXP=$(date -v+1H +%s 2>/dev/null || date -d '+1 hour' +%s 2>/dev/null)

# ============================================================================
# JWT Generation
# ============================================================================

# Base64 URL encoding (removes padding and makes URL-safe)
base64url() {
    base64 2>/dev/null | tr '+/' '-_' | tr -d '=' || base64 -w 0 | tr '+/' '-_' | tr -d '='
}

# Create JWT components
HEADER=$(echo -n '{"alg":"HS256","typ":"JWT"}' | base64url)
PAYLOAD=$(echo -n "{\"role\":\"$ROLE\",\"email\":\"$EMAIL\",\"aud\":\"postgrest\",\"exp\":$EXP}" | base64url)
SIGNATURE=$(echo -n "$HEADER.$PAYLOAD" | openssl dgst -sha256 -hmac "$JWT_SECRET" -binary | base64url)

# Complete JWT token
JWT_TOKEN="$HEADER.$PAYLOAD.$SIGNATURE"

# ============================================================================
# Output
# ============================================================================

echo "JWT Token Generated"
echo "===================="
echo "Role:    $ROLE"
echo "Email:   $EMAIL"
echo "Expires: $(date -r $EXP 2>/dev/null || date -d @$EXP 2>/dev/null)"
echo ""
echo "Token:"
echo "$JWT_TOKEN"
echo ""
echo "Usage:"
echo "curl http://localhost:3000/rpc/me \\"
echo "  -H \"Authorization: Bearer \$JWT_TOKEN\""
echo ""

# Export for use in other scripts
export JWT_TOKEN
echo "Exported as \$JWT_TOKEN environment variable"
