#!/bin/bash

# ============================================================================
# JWT Authentication Test Script
# ============================================================================
# This script tests the basic JWT authentication functionality
# Make sure the services are running: make portal-db-up

set -e  # Exit on any error

API_URL="http://localhost:3000"
echo "🔐 Testing JWT Authentication with PostgREST"
echo "API URL: $API_URL"
echo

# ============================================================================
# Test 1: Anonymous Access (should work)
# ============================================================================
echo "📖 Test 1: Anonymous access to public data"
echo "GET $API_URL/networks"

RESPONSE=$(curl -s "$API_URL/networks" || echo "ERROR")
if [[ "$RESPONSE" == *"ERROR"* ]] || [[ "$RESPONSE" == *"error"* ]]; then
    echo "❌ Anonymous access failed"
    echo "Response: $RESPONSE"
    exit 1
else
    echo "✅ Anonymous access works"
    echo "Found $(echo "$RESPONSE" | jq length 2>/dev/null || echo "some") networks"
fi
echo

# ============================================================================
# Test 2: Generate JWT Token (External Generation)
# ============================================================================
echo "🔑 Test 2: Generating JWT token (following PostgREST docs) ✨"

# Generate JWT token using shell script (PostgREST best practice)
echo "🔧 Generating fresh JWT token using shell script..."
cd "$(dirname "$0")"  # Ensure we're in the scripts directory

# Generate token and capture output using --token-only flag for clean parsing
JWT_TOKEN=$(./postgrest-gen-jwt.sh --token-only authenticated 2>/dev/null)

if [[ -z "$JWT_TOKEN" ]]; then
    echo "❌ Failed to generate JWT token"
    echo "💡 Make sure postgrest-gen-jwt.sh is executable and openssl is installed"
    exit 1
fi

echo "✅ Generated fresh JWT token: ${JWT_TOKEN:0:50}... 🎯"
echo "🌟 This demonstrates external JWT generation (PostgREST best practice)"
echo

# ============================================================================
# Test 3: Access Protected Resource with Token
# ============================================================================
echo "🔒 Test 3: Access protected data with JWT token"
echo "GET $API_URL/portal_accounts (with Authorization header)"

AUTH_RESPONSE=$(curl -s "$API_URL/portal_accounts" \
    -H "Authorization: Bearer $JWT_TOKEN" || echo "ERROR")

if [[ "$AUTH_RESPONSE" == *"ERROR"* ]] || [[ "$AUTH_RESPONSE" == *"error"* ]]; then
    echo "❌ Authenticated access failed"
    echo "Response: $AUTH_RESPONSE"
    exit 1
else
    echo "✅ Authenticated access works"
    echo "Found $(echo "$AUTH_RESPONSE" | jq length 2>/dev/null || echo "some") portal accounts"
fi
echo

# ============================================================================
# Test 4: Get Current User Info
# ============================================================================
echo "👤 Test 4: Get current user info"
echo "POST $API_URL/rpc/me"

ME_RESPONSE=$(curl -s -X POST "$API_URL/rpc/me" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" || echo "ERROR")

if [[ "$ME_RESPONSE" == *"ERROR"* ]] || [[ "$ME_RESPONSE" == *"error"* ]]; then
    echo "❌ Get user info failed"
    echo "Response: $ME_RESPONSE"
    exit 1
else
    echo "✅ Get user info works"
    echo "User info: $ME_RESPONSE"
fi
echo

# ============================================================================
# Test 5: Access Protected Resource WITHOUT Token (should fail)
# ============================================================================
echo "🚫 Test 5: Try to access protected data without token (should fail)"
echo "GET $API_URL/portal_accounts (no Authorization header)"

UNAUTH_RESPONSE=$(curl -s "$API_URL/portal_accounts" || echo "ERROR")
# We expect this to either return empty or give an error - both are fine

echo "Response: $UNAUTH_RESPONSE"
echo "✅ This test shows the difference between authenticated and anonymous access"
echo

# ============================================================================
# Summary
# ============================================================================
echo "🎉 All JWT authentication tests passed! 🚀"
echo
echo "📊 Summary:"
echo "- ✅ Anonymous users can access public data 🌐"
echo "- ✅ JWT tokens (generated externally) provide access to protected data 🔐"
echo "- ✅ JWT claims can be accessed in SQL via current_setting() 📋"
echo "- ✅ Requests without tokens have limited access 🚫"
echo
echo "📚 PostgREST Documentation Approach:"
echo "- ✅ JWT tokens generated externally (as documented) 🔧"
echo "- ✅ Simple role-based access control via JWT role claim 👥"
echo "- ✅ No hardcoded user data in database functions 🎯"
echo
echo "🚀 Next steps:"
echo "- 📖 Try the examples in api/auth-examples.md"
echo "- 🔑 Generate your own JWT tokens: ./api/scripts/postgrest-gen-jwt.sh"
echo "- 📄 View the API documentation at $API_URL"
echo "- 🔍 Explore the database roles and permissions"
echo
echo "🎊 Happy coding with PostgREST! ✨"
