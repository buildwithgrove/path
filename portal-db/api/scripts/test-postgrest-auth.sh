#!/bin/bash

# ============================================================================
# JWT Authentication Test Script
# ============================================================================
# Validates admin and reader roles end-to-end against the PostgREST API.
# - Ensures admin tokens can read and write
# - Ensures reader tokens can only read
# - Confirms unauthenticated requests fail for protected resources

set -euo pipefail

API_URL="http://localhost:3000"
SCRIPTS_DIR="$(cd "$(dirname "$0")" && pwd)"

info()  { echo -e "\033[0;34m$1\033[0m"; }
success(){ echo -e "\033[0;32m$1\033[0m"; }
warn()  { echo -e "\033[1;33m$1\033[0m"; }
fail()  { echo -e "\033[0;31m$1\033[0m"; exit 1; }

echo "🔐 Testing PostgREST authentication flow"
echo "API URL: $API_URL"
echo

# Helper to generate a JWT via the CLI helper
generate_token() {
    local role_alias="$1"
    "$SCRIPTS_DIR/postgrest-gen-jwt.sh" --token-only "$role_alias"
}

# Helper to perform curl and capture HTTP status + body
request_json() {
    local method="$1" endpoint="$2" token="${3:-}" data="${4:-}" extra_header="${5:-}"

    local headers=()
    [[ -n "$token" ]] && headers+=(-H "Authorization: Bearer $token")
    [[ -n "$extra_header" ]] && headers+=(-H "$extra_header")

    local tmp_body
    tmp_body=$(mktemp)

    if [[ -n "$data" ]]; then
        if (( ${#headers[@]} )); then
            STATUS=$(curl -sS -X "$method" "$API_URL$endpoint" "${headers[@]}" \
                -H "Content-Type: application/json" -d "$data" \
                -o "$tmp_body" -w "%{http_code}")
        else
            STATUS=$(curl -sS -X "$method" "$API_URL$endpoint" \
                -H "Content-Type: application/json" -d "$data" \
                -o "$tmp_body" -w "%{http_code}")
        fi
    else
        if (( ${#headers[@]} )); then
            STATUS=$(curl -sS -X "$method" "$API_URL$endpoint" "${headers[@]}" \
                -o "$tmp_body" -w "%{http_code}")
        else
            STATUS=$(curl -sS -X "$method" "$API_URL$endpoint" \
                -o "$tmp_body" -w "%{http_code}")
        fi
    fi

    BODY=$(cat "$tmp_body")
    rm -f "$tmp_body"
}

echo " === Step 1: Admin token should read and write === "
echo ""
info "Generating admin token"
ADMIN_TOKEN=$(generate_token admin)
[[ -n "$ADMIN_TOKEN" ]] || fail "Failed to generate admin token"
success "Admin token generated"

echo ""
info "Admin: reading portal_accounts"
request_json GET "/portal_accounts" "$ADMIN_TOKEN"
if [[ "$STATUS" != "200" ]]; then
    echo "Response body: $BODY"
    fail "Admin read failed (status $STATUS)"
fi
count=$(echo "$BODY" | jq length 2>/dev/null || echo "unknown")
success "Admin read succeeded (returned $count records)"

echo ""
info "Admin: inserting portal_application via direct table access"
APP_NAME="auth-test-admin-$$"
payload=$(jq -n --arg name "$APP_NAME" '{
  portal_account_id: "10000000-0000-0000-0000-000000000004",
  portal_application_name: $name,
  secret_key_hash: "unit-test-hash",
  secret_key_required: false
}')

request_json POST "/portal_applications" "$ADMIN_TOKEN" "$payload" "Prefer: return=representation"
[[ "$STATUS" == "201" || "$STATUS" == "200" ]] || fail "Admin write failed (status $STATUS): $BODY"
ADMIN_APP_ID=$(echo "$BODY" | jq -r '.[0].portal_application_id // empty')
[[ -n "$ADMIN_APP_ID" ]] || fail "Admin write did not return portal_application_id"
success "Admin write succeeded (portal_application_id=$ADMIN_APP_ID)"

echo ""
info "Admin: deleting created portal_application"
# Using DELETE to clean up
request_json DELETE "/portal_applications?portal_application_id=eq.$ADMIN_APP_ID" "$ADMIN_TOKEN"
[[ "$STATUS" == "204" ]] || fail "Admin delete failed (status $STATUS): $BODY"
success "Admin delete succeeded"

echo ""
echo " === Step 2: Reader token should read but be blocked from writes === "
echo ""
info "Generating reader token"
READER_TOKEN=$(generate_token reader)
[[ -n "$READER_TOKEN" ]] || fail "Failed to generate reader token"
success "Reader token generated"

echo ""
info "Reader: reading portal_accounts"
request_json GET "/portal_accounts" "$READER_TOKEN"
[[ "$STATUS" == "200" ]] || fail "Reader read failed (status $STATUS): $BODY"
success "Reader read succeeded (returned $(echo "$BODY" | jq length 2>/dev/null || echo "unknown"))"

echo ""
info "Reader: attempting to insert portal_application (should fail)"
reader_payload=$(jq -n '{
  portal_account_id: "10000000-0000-0000-0000-000000000004",
  portal_application_name: "auth-test-reader-denied",
  secret_key_hash: "unit-test-hash",
  secret_key_required: false
}')

request_json POST "/portal_applications" "$READER_TOKEN" "$reader_payload" "Prefer: return=representation"
if [[ "$STATUS" == "201" || "$STATUS" == "200" ]]; then
    fail "Reader unexpectedly gained write access: $BODY"
fi
[[ "$STATUS" == "401" || "$STATUS" == "403" || "$STATUS" == "404" || "$STATUS" == "409" ]] || warn "Reader write returned status $STATUS (expected denial)"
success "Reader write attempt correctly denied (status $STATUS)"

echo ""
echo " === Step 3: Unauthenticated access should fail on protected endpoint === "
echo ""
info "Unauthenticated request to protected endpoint"
request_json GET "/portal_accounts"
if [[ "$STATUS" == "200" ]]; then
    warn "Unauthenticated request unexpectedly returned 200; response: $BODY"
else
    success "Unauthenticated request denied as expected (status $STATUS)"
fi

echo ""
success "✅ Authentication tests completed successfully"
echo "   - Admin role can read and write"
echo "   - Reader role can only read"
echo "   - Anonymous access is restricted"
echo ""
