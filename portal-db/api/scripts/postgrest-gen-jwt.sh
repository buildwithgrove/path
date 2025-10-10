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
#   ./postgrest-gen-jwt.sh                                    # Use defaults (reads secret from .env)
#   ./postgrest-gen-jwt.sh --role reader --email user@email   # Custom role and email
#   ./postgrest-gen-jwt.sh --expires 24h                      # 24 hour expiry
#   ./postgrest-gen-jwt.sh --expires never                    # Never expires
#   ./postgrest-gen-jwt.sh --token-only                       # Output token only (for scripting)
#   ./postgrest-gen-jwt.sh --secret YOUR_SECRET               # Provide custom JWT secret
#   ./postgrest-gen-jwt.sh --help                             # Display usage information

set -e

# Color codes
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
BOLD='\033[1m'
RESET='\033[0m'

# Determine the script directory and portal-db root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PORTAL_DB_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Try to load JWT secret from .env file
if [[ -f "$PORTAL_DB_ROOT/.env" ]]; then
    # Source the .env file to get POSTGREST_JWT_SECRET
    set -a
    source "$PORTAL_DB_ROOT/.env"
    set +a
    
    if [[ -n "$POSTGREST_JWT_SECRET" ]]; then
        JWT_SECRET="$POSTGREST_JWT_SECRET"
    fi
fi

# Default values
ROLE="portal_db_admin"
EMAIL="john@doe.com"
EXPIRES="1h"
TOKEN_ONLY=false

# Show help function (defined early so it can be called during argument parsing)
print_help() {
    cat <<EOF
${BOLD}${CYAN}JWT Token Generator for PostgREST${RESET}

${BOLD}DESCRIPTION${RESET}
    Generates JWT tokens for PostgREST authentication using HMAC-SHA256 signing.
    The JWT secret is required and can be provided via .env file, environment 
    variable, or command line flag.

${BOLD}USAGE${RESET}
    postgrest-gen-jwt.sh [OPTIONS]

${BOLD}OPTIONS${RESET}
    ${CYAN}-h, --help${RESET}
        Show this help message and exit

    ${CYAN}--role ROLE${RESET}
        Database role to embed in the JWT
        Default: portal_db_admin
        Aliases: admin (portal_db_admin), reader (portal_db_reader)

    ${CYAN}--email EMAIL${RESET}
        Email claim to embed in the JWT
        Default: john@doe.com

    ${CYAN}--expires DURATION${RESET}
        Set token expiration time
        Default: 1h
        Formats:
          - Hours: 1h, 2h, 24h, etc.
          - Days:  1d, 7d, 30d, etc.
          - Never: never (token never expires)

    ${CYAN}--secret SECRET${RESET}
        JWT secret for signing the token
        Overrides .env file and environment variable
        Note: Must match the secret in postgrest.conf

    ${CYAN}--token-only${RESET}
        Output only the JWT token (for scripting)
        Suppresses all other output

${BOLD}JWT SECRET (REQUIRED)${RESET}
    The script requires a JWT secret to sign tokens. It looks for the secret 
    in the following order (first found wins):
    
    ${CYAN}1.${RESET} --secret command line flag
    ${CYAN}2.${RESET} POSTGREST_JWT_SECRET in portal-db/.env file
    ${CYAN}3.${RESET} JWT_SECRET environment variable

    ${YELLOW}Important: The secret must match the one configured in postgrest.conf${RESET}

${BOLD}EXAMPLES${RESET}
    ${CYAN}# Generate token with defaults (reads secret from .env)${RESET}
    ./postgrest-gen-jwt.sh

    ${CYAN}# Generate token with custom role and email${RESET}
    ./postgrest-gen-jwt.sh --role reader --email user@example.com

    ${CYAN}# Generate token using role alias${RESET}
    ./postgrest-gen-jwt.sh --role admin --email admin@example.com

    ${CYAN}# Generate token that expires in 24 hours${RESET}
    ./postgrest-gen-jwt.sh --expires 24h

    ${CYAN}# Generate token that expires in 7 days${RESET}
    ./postgrest-gen-jwt.sh --expires 7d --role admin --email user@example.com

    ${CYAN}# Generate token that never expires${RESET}
    ./postgrest-gen-jwt.sh --expires never

    ${CYAN}# Generate token with custom secret${RESET}
    ./postgrest-gen-jwt.sh --secret your_jwt_secret_here --role admin

    ${CYAN}# Generate token for scripting (token only output)${RESET}
    ./postgrest-gen-jwt.sh --token-only --expires never --role admin

    ${CYAN}# Use token in curl request${RESET}
    export POSTGREST_JWT_TOKEN=\$(./postgrest-gen-jwt.sh --token-only)
    curl -H "Authorization: Bearer \$POSTGREST_JWT_TOKEN" http://localhost:3000/organizations

${BOLD}ROLE ALIASES${RESET}
    ${CYAN}admin${RESET}   -> portal_db_admin
    ${CYAN}reader${RESET}  -> portal_db_reader

${BOLD}DEPENDENCIES${RESET}
    - openssl (for HMAC-SHA256 signing)
    - base64 (for encoding)
    - date (for expiration calculation)

${BOLD}REFERENCE${RESET}
    https://docs.postgrest.org/en/v13/tutorials/tut1.html

EOF
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --help|-h)
            print_help
            exit 0
            ;;
        --token-only)
            TOKEN_ONLY=true
            shift
            ;;
        --role)
            if [[ -z "$2" || "$2" == --* ]]; then
                echo -e "${RED}${BOLD}Error: --role requires a value${RESET}" >&2
                echo "Run with --help for usage information" >&2
                exit 1
            fi
            ROLE="$2"
            shift 2
            ;;
        --email)
            if [[ -z "$2" || "$2" == --* ]]; then
                echo -e "${RED}${BOLD}Error: --email requires a value${RESET}" >&2
                echo "Run with --help for usage information" >&2
                exit 1
            fi
            EMAIL="$2"
            shift 2
            ;;
        --expires)
            if [[ -z "$2" || "$2" == --* ]]; then
                echo -e "${RED}${BOLD}Error: --expires requires a value${RESET}" >&2
                echo "Run with --help for usage information" >&2
                exit 1
            fi
            EXPIRES="$2"
            shift 2
            ;;
        --secret)
            if [[ -z "$2" || "$2" == --* ]]; then
                echo -e "${RED}${BOLD}Error: --secret requires a value${RESET}" >&2
                echo "Run with --help for usage information" >&2
                exit 1
            fi
            JWT_SECRET="$2"
            shift 2
            ;;
        *)
            echo -e "${RED}${BOLD}Error: Unknown option '$1'${RESET}" >&2
            echo "Run with --help for usage information" >&2
            exit 1
            ;;
    esac
done

# Validate that JWT_SECRET is available
if [[ -z "$JWT_SECRET" ]]; then
    echo -e "${RED}${BOLD}❌ Error: JWT secret not found${RESET}" >&2
    echo "" >&2
    echo -e "${BOLD}The JWT secret must be provided via one of the following methods:${RESET}" >&2
    echo "" >&2
    echo -e "  ${CYAN}1.${RESET} Use the ${CYAN}--secret${RESET} flag:" >&2
    echo -e "     ${CYAN}./postgrest-gen-jwt.sh --secret YOUR_SECRET${RESET}" >&2
    echo "" >&2
    echo -e "  ${CYAN}2.${RESET} Create a ${CYAN}.env${RESET} file at ${YELLOW}$PORTAL_DB_ROOT/.env${RESET} with:" >&2
    echo -e "     ${CYAN}POSTGREST_JWT_SECRET=your_secret_here${RESET}" >&2
    echo "" >&2
    echo -e "  ${CYAN}3.${RESET} Set the ${CYAN}JWT_SECRET${RESET} environment variable:" >&2
    echo -e "     ${CYAN}JWT_SECRET=your_secret ./postgrest-gen-jwt.sh${RESET}" >&2
    echo "" >&2
    echo -e "${YELLOW}Note: The secret must match the one configured in postgrest.conf${RESET}" >&2
    echo -e "${YELLOW}Run with --help for more information${RESET}" >&2
    exit 1
fi

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
POSTGREST_JWT_TOKEN="$HEADER.$PAYLOAD.$SIGNATURE"

# ============================================================================
# Output
# ============================================================================

if [[ "$TOKEN_ONLY" == true ]]; then
    # Just output the token for scripting
    echo "$POSTGREST_JWT_TOKEN"
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
    echo -e "${YELLOW}$POSTGREST_JWT_TOKEN${RESET}"
    echo ""
    echo -e "${BOLD}Export to shell:${RESET}"
    echo -e "${CYAN}export POSTGREST_JWT_TOKEN=\"$POSTGREST_JWT_TOKEN\"${RESET}"
    echo ""
    echo -e "${BOLD}Usage:${RESET}"
    echo -e "${CYAN}curl http://localhost:3000/organizations -H \"Authorization: Bearer \$POSTGREST_JWT_TOKEN\"${RESET}"
    echo ""
fi
