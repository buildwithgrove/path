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
#   ./postgrest-gen-jwt.sh                              # portal_db_admin role, sample email, 1h expiry
#   ./postgrest-gen-jwt.sh portal_db_reader user@email  # custom role + email
#   ./postgrest-gen-jwt.sh --expires 24h                # 24 hour expiry
#   ./postgrest-gen-jwt.sh --expires never              # Never expires
#   ./postgrest-gen-jwt.sh --token-only portal_db_admin user@email
#   ./postgrest-gen-jwt.sh --help                       # display usage information

set -e

# Color codes
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
RESET='\033[0m'

# JWT secret from postgrest.conf (must match exactly)
# TODO_PRODUCTION: Extract JWT_SECRET from postgrest.conf automatically to maintain single source of truth and avoid drift between files
JWT_SECRET="${JWT_SECRET:-supersecretjwtsecretforlocaldevelopment123456789}"

# Default values
ROLE="portal_db_admin"
EMAIL="john@doe.com"
EXPIRES="1h"
TOKEN_ONLY=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --token-only)
            TOKEN_ONLY=true
            shift
            ;;
        --expires)
            EXPIRES="$2"
            shift 2
            ;;
        --help|-h)
            print_help
            exit 0
            ;;
        *)
            if [[ -z "$ROLE" || "$ROLE" == "portal_db_admin" ]]; then
                ROLE="$1"
            elif [[ "$EMAIL" == "john@doe.com" ]]; then
                EMAIL="$1"
            fi
            shift
            ;;
    esac
done

# Show help
print_help() {
    cat <<'EOF'
Usage: postgrest-gen-jwt.sh [OPTIONS] [ROLE] [EMAIL]

Options:
  --help, -h              Show this help message and exit
  --token-only            Print only the JWT for scripting
  --expires DURATION      Set token expiration (default: 1h)
                          Examples: 1h, 24h, 7d, 30d, never

Role aliases:
  admin                   Shortcut for portal_db_admin
  reader                  Shortcut for portal_db_reader

Positional arguments:
  ROLE                    Database role to embed in the JWT (default: portal_db_admin)
  EMAIL                   Email claim to embed in the JWT (default: john@doe.com)

Expiration formats:
  1h, 2h, etc.           Hours (e.g., 1h = 1 hour from now)
  1d, 7d, 30d            Days (e.g., 7d = 7 days from now)
  never                  Token never expires (no exp claim)

Examples:
  # Generate token with default 1 hour expiry
  ./postgrest-gen-jwt.sh

  # Generate token with custom role and email
  ./postgrest-gen-jwt.sh reader user@example.com

  # Generate token that expires in 24 hours
  ./postgrest-gen-jwt.sh --expires 24h

  # Generate token that expires in 7 days
  ./postgrest-gen-jwt.sh --expires 7d admin user@example.com

  # Generate token that never expires
  ./postgrest-gen-jwt.sh --expires never

  # Generate token for scripting (token only output)
  ./postgrest-gen-jwt.sh --token-only --expires never admin
EOF
}

# Allow short aliases for role names
translate_role() {
    case "$1" in
        admin)
            echo "portal_db_admin"
            ;;
        reader)
            echo "portal_db_reader"
            ;;
        *)
            echo "$1"
            ;;
    esac
}

# Map alias after parsing the arguments
ROLE=$(translate_role "$ROLE")

# Calculate expiration based on EXPIRES value
calculate_expiration() {
    local duration="$1"
    
    if [[ "$duration" == "never" ]]; then
        echo "never"
        return
    fi
    
    # Extract number and unit (e.g., "24h" -> 24 and h)
    local num="${duration//[^0-9]/}"
    local unit="${duration//[0-9]/}"
    
    # Default to hours if no unit specified
    if [[ -z "$unit" ]]; then
        unit="h"
    fi
    
    case "$unit" in
        h)
            # Hours
            date -v+${num}H +%s 2>/dev/null || date -d "+${num} hours" +%s 2>/dev/null
            ;;
        d)
            # Days
            date -v+${num}d +%s 2>/dev/null || date -d "+${num} days" +%s 2>/dev/null
            ;;
        *)
            echo "Error: Invalid expiration format '$duration'. Use format like: 1h, 24h, 7d, or 'never'" >&2
            exit 1
            ;;
    esac
}

EXP=$(calculate_expiration "$EXPIRES")

# ============================================================================
# JWT Generation
# ============================================================================

# Base64 URL encoding (removes padding and makes URL-safe)
base64url() {
    base64 2>/dev/null | tr '+/' '-_' | tr -d '=' || base64 -w 0 | tr '+/' '-_' | tr -d '='
}

# Create JWT components
HEADER=$(echo -n '{"alg":"HS256","typ":"JWT"}' | base64url)

# Create payload with or without exp claim
if [[ "$EXP" == "never" ]]; then
    PAYLOAD=$(echo -n "{\"role\":\"$ROLE\",\"email\":\"$EMAIL\",\"aud\":\"postgrest\"}" | base64url)
else
    PAYLOAD=$(echo -n "{\"role\":\"$ROLE\",\"email\":\"$EMAIL\",\"aud\":\"postgrest\",\"exp\":$EXP}" | base64url)
fi

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
    
    if [[ "$EXP" == "never" ]]; then
        echo -e "${BOLD}Expires:${RESET} ${YELLOW}Never (no expiration)${RESET}"
    else
        echo -e "${BOLD}Expires:${RESET} ${BLUE}$(date -r $EXP 2>/dev/null || date -d @$EXP 2>/dev/null)${RESET} ${CYAN}($EXPIRES)${RESET}"
    fi
    
    echo ""
    echo -e "${BOLD}Token:${RESET}"
    echo -e "${YELLOW}$JWT_TOKEN${RESET}"
    echo ""
    echo -e "${BOLD}Export to shell:${RESET}"
    echo -e "${CYAN}export JWT_TOKEN=\"$JWT_TOKEN\"${RESET}"
    echo ""
    echo -e "${BOLD}Usage:${RESET}"
    echo -e "${CYAN}curl http://localhost:3000/organizations -H \"Authorization: Bearer \$JWT_TOKEN\"${RESET}"
    echo ""
fi
