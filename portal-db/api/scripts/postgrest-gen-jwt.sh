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
#   ./postgrest-gen-jwt.sh                           # authenticated role, default email
#   ./postgrest-gen-jwt.sh anon                      # anon role
#   ./postgrest-gen-jwt.sh authenticated user@email  # custom email

set -e

# Color codes
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
RESET='\033[0m'

# JWT secret from postgrest.conf (must match exactly)
JWT_SECRET="supersecretjwtsecretforlocaldevelopment123456789"

# Parse arguments
ROLE="${1:-authenticated}"
EMAIL="${2:-john@doe.com}"
TOKEN_ONLY=false

# Check for --token-only flag
if [[ "$1" == "--token-only" ]]; then
    TOKEN_ONLY=true
    ROLE="${2:-authenticated}"
    EMAIL="${3:-john@doe.com}"
fi

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

if [[ "$TOKEN_ONLY" == true ]]; then
    # Just output the token for scripting
    echo "$JWT_TOKEN"
else
    # Full colorized output
    echo -e "${GREEN}${BOLD}🔑 JWT Token Generated${RESET}"
    echo -e "${GREEN}${BOLD}=======================${RESET}"
    echo -e "${BOLD}Role:${RESET}    ${BLUE}$ROLE${RESET}"
    echo -e "${BOLD}Email:${RESET}   ${BLUE}$EMAIL${RESET}"
    echo -e "${BOLD}Expires:${RESET} ${BLUE}$(date -r $EXP 2>/dev/null || date -d @$EXP 2>/dev/null)${RESET}"
    echo ""
    echo -e "${BOLD}Token:${RESET}"
    echo -e "${YELLOW}$JWT_TOKEN${RESET}"
    echo ""
    echo -e "${BOLD}Export to shell:${RESET}"
    echo -e "${CYAN}export JWT_TOKEN=\"$JWT_TOKEN\"${RESET}"
    echo ""
    echo -e "${BOLD}Usage:${RESET}"
    echo -e "${CYAN}curl http://localhost:3000/rpc/me -H \"Authorization: Bearer \$JWT_TOKEN\"${RESET}"
    echo ""
fi
