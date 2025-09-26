#!/bin/bash

# 🚀 Test Data Hydration Script for Portal DB
# This script populates the Portal DB with test data for development and testing

set -e

# 🎨 Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 📝 Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 🔍 Function to validate database connection
validate_db_connection() {
    if [ -z "$DB_CONNECTION_STRING" ]; then
        print_status $RED "❌ Error: DB_CONNECTION_STRING environment variable is required"
        print_status $YELLOW "💡 Expected format: postgresql://user:password@host:port/database"
        print_status $YELLOW "💡 For local development: postgresql://postgres:portal_password@localhost:5435/portal_db"
        exit 1
    fi
    
    print_status $BLUE "🔍 Testing database connection..."
    if ! psql "$DB_CONNECTION_STRING" -c "SELECT 1;" > /dev/null 2>&1; then
        print_status $RED "❌ Error: Cannot connect to database"
        print_status $YELLOW "💡 Make sure the database is running: make postgrest-up"
        exit 1
    fi
    
    print_status $GREEN "✅ Database connection successful"
}

# 📊 Function to insert test data
insert_test_data() {
    print_status $BLUE "📊 Inserting test data into Portal DB..."
    
    # Execute the SQL in a single transaction
    psql "$DB_CONNECTION_STRING" <<EOF
BEGIN;

-- Insert test portal plans (has PRIMARY KEY, so ON CONFLICT works)
INSERT INTO portal_plans (portal_plan_type, portal_plan_type_description, plan_usage_limit, plan_usage_limit_interval, plan_rate_limit_rps, plan_application_limit) VALUES
    ('FREE', 'Free tier with basic limits', 1000, 'day', 10, 2),
    ('STARTER', 'Starter plan for small projects', 10000, 'day', 50, 5),
    ('PRO', 'Professional plan for growing businesses', 100000, 'month', 200, 20),
    ('ENTERPRISE', 'Enterprise plan with custom limits', NULL, NULL, 1000, 100)
ON CONFLICT (portal_plan_type) DO NOTHING;

-- Insert test organizations (check if exists first)
INSERT INTO organizations (organization_name) 
SELECT name FROM (VALUES 
    ('Acme Corporation'),
    ('Tech Innovators LLC'),
    ('Blockchain Solutions Inc'),
    ('Web3 Builders Co')
) AS new_orgs(name)
WHERE NOT EXISTS (
    SELECT 1 FROM organizations WHERE organization_name = new_orgs.name
);

-- Insert test services (has PRIMARY KEY, so ON CONFLICT works)
INSERT INTO services (service_id, service_name, compute_units_per_relay, service_domains, network_id, active, quality_fallback_enabled, hard_fallback_enabled) VALUES
    ('ethereum-mainnet', 'Ethereum Mainnet', 1, ARRAY['eth-mainnet.gateway.pokt.network'], 'pocket', true, true, false),
    ('ethereum-sepolia', 'Ethereum Sepolia Testnet', 1, ARRAY['eth-sepolia.gateway.pokt.network'], 'pocket', true, false, false),
    ('polygon-mainnet', 'Polygon Mainnet', 1, ARRAY['poly-mainnet.gateway.pokt.network'], 'pocket', true, true, true),
    ('arbitrum-one', 'Arbitrum One', 2, ARRAY['arbitrum-one.gateway.pokt.network'], 'pocket', true, false, false),
    ('base-mainnet', 'Base Mainnet', 2, ARRAY['base-mainnet.gateway.pokt.network'], 'pocket', false, false, false)
ON CONFLICT (service_id) DO NOTHING;

-- Insert test service endpoints (check if exists first)
INSERT INTO service_endpoints (service_id, endpoint_type) 
SELECT service_id, endpoint_type FROM (VALUES 
    ('ethereum-mainnet', 'JSON-RPC'::endpoint_type),
    ('ethereum-mainnet', 'WSS'::endpoint_type),
    ('ethereum-sepolia', 'JSON-RPC'::endpoint_type),
    ('polygon-mainnet', 'JSON-RPC'::endpoint_type),
    ('polygon-mainnet', 'REST'::endpoint_type),
    ('arbitrum-one', 'JSON-RPC'::endpoint_type),
    ('base-mainnet', 'JSON-RPC'::endpoint_type)
) AS new_endpoints(service_id, endpoint_type)
WHERE NOT EXISTS (
    SELECT 1 FROM service_endpoints se 
    WHERE se.service_id = new_endpoints.service_id 
    AND se.endpoint_type = new_endpoints.endpoint_type
);

-- Insert test service fallbacks (check if exists first)
INSERT INTO service_fallbacks (service_id, fallback_url) 
SELECT service_id, fallback_url FROM (VALUES 
    ('ethereum-mainnet', 'https://eth-mainnet.infura.io/v3/fallback'),
    ('ethereum-mainnet', 'https://mainnet.infura.io/v3/backup'),
    ('polygon-mainnet', 'https://polygon-mainnet.infura.io/v3/fallback'),
    ('arbitrum-one', 'https://arbitrum-mainnet.infura.io/v3/fallback')
) AS new_fallbacks(service_id, fallback_url)
WHERE NOT EXISTS (
    SELECT 1 FROM service_fallbacks sf 
    WHERE sf.service_id = new_fallbacks.service_id 
    AND sf.fallback_url = new_fallbacks.fallback_url
);

-- Insert test portal users (has UNIQUE constraint, so ON CONFLICT works)
-- Using deterministic UUIDs for reliable testing
INSERT INTO portal_users (portal_user_id, portal_user_email, signed_up, portal_admin) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin@grove.city', true, true),
    ('00000000-0000-0000-0000-000000000002', 'alice@acme.com', true, false),
    ('00000000-0000-0000-0000-000000000003', 'bob@techinnovators.com', true, false),
    ('00000000-0000-0000-0000-000000000004', 'charlie@blockchain.com', false, false)
ON CONFLICT (portal_user_email) DO NOTHING;

-- Insert test portal accounts with deterministic UUIDs
INSERT INTO portal_accounts (portal_account_id, organization_id, portal_plan_type, user_account_name, internal_account_name, billing_type) 
SELECT 
    CASE 
        WHEN org.organization_name = 'Acme Corporation' THEN '10000000-0000-0000-0000-000000000001'
        WHEN org.organization_name = 'Tech Innovators LLC' THEN '10000000-0000-0000-0000-000000000002'
        WHEN org.organization_name = 'Blockchain Solutions Inc' THEN '10000000-0000-0000-0000-000000000003'
        ELSE '10000000-0000-0000-0000-000000000004'
    END,
    org.organization_id,
    CASE 
        WHEN org.organization_name = 'Acme Corporation' THEN 'ENTERPRISE'
        WHEN org.organization_name = 'Tech Innovators LLC' THEN 'PRO'
        WHEN org.organization_name = 'Blockchain Solutions Inc' THEN 'STARTER'
        ELSE 'FREE'
    END,
    CASE 
        WHEN org.organization_name = 'Acme Corporation' THEN 'acme-corp'
        WHEN org.organization_name = 'Tech Innovators LLC' THEN 'tech-innovators'
        WHEN org.organization_name = 'Blockchain Solutions Inc' THEN 'blockchain-solutions'
        ELSE 'web3-builders'
    END,
    CASE 
        WHEN org.organization_name = 'Acme Corporation' THEN 'internal-acme'
        WHEN org.organization_name = 'Tech Innovators LLC' THEN 'internal-tech'
        WHEN org.organization_name = 'Blockchain Solutions Inc' THEN 'internal-blockchain'
        ELSE 'internal-web3'
    END,
    'stripe'
FROM organizations org
WHERE NOT EXISTS (
    SELECT 1 FROM portal_accounts pa 
    WHERE pa.organization_id = org.organization_id
);

-- Insert test portal account RBAC (link users to accounts)
INSERT INTO portal_account_rbac (portal_account_id, portal_user_id, role_name, user_joined_account)
SELECT 
    pa.portal_account_id,
    pu.portal_user_id,
    'OWNER',
    true
FROM portal_accounts pa
CROSS JOIN portal_users pu
WHERE (
    (pa.user_account_name = 'acme-corp' AND pu.portal_user_email = 'alice@acme.com') OR
    (pa.user_account_name = 'tech-innovators' AND pu.portal_user_email = 'bob@techinnovators.com') OR  
    (pa.user_account_name = 'blockchain-solutions' AND pu.portal_user_email = 'charlie@blockchain.com') OR
    (pa.user_account_name = 'web3-builders' AND pu.portal_user_email = 'admin@grove.city')
)
AND NOT EXISTS (
    SELECT 1 FROM portal_account_rbac rbac 
    WHERE rbac.portal_account_id = pa.portal_account_id 
    AND rbac.portal_user_id = pu.portal_user_id
);

-- Insert test portal applications with deterministic UUIDs
INSERT INTO portal_applications (
    portal_application_id,
    portal_account_id, 
    portal_application_name, 
    emoji, 
    portal_application_description,
    favorite_service_ids,
    secret_key_required
) 
SELECT 
    CASE 
        WHEN pa.user_account_name = 'acme-corp' THEN '20000000-0000-0000-0000-000000000001'
        WHEN pa.user_account_name = 'tech-innovators' THEN '20000000-0000-0000-0000-000000000002'
        WHEN pa.user_account_name = 'blockchain-solutions' THEN '20000000-0000-0000-0000-000000000003'
        ELSE '20000000-0000-0000-0000-000000000004'
    END,
    pa.portal_account_id,
    CASE 
        WHEN pa.user_account_name = 'acme-corp' THEN 'DeFi Dashboard'
        WHEN pa.user_account_name = 'tech-innovators' THEN 'NFT Marketplace'
        WHEN pa.user_account_name = 'blockchain-solutions' THEN 'Analytics Platform'
        ELSE 'Test Application'
    END,
    CASE 
        WHEN pa.user_account_name = 'acme-corp' THEN '💰'
        WHEN pa.user_account_name = 'tech-innovators' THEN '🖼️'
        WHEN pa.user_account_name = 'blockchain-solutions' THEN '📊'
        ELSE '🧪'
    END,
    CASE 
        WHEN pa.user_account_name = 'acme-corp' THEN 'Real-time DeFi protocol dashboard'
        WHEN pa.user_account_name = 'tech-innovators' THEN 'Multi-chain NFT marketplace application'
        WHEN pa.user_account_name = 'blockchain-solutions' THEN 'Cross-chain analytics and monitoring'
        ELSE 'General purpose test application'
    END,
    ARRAY['ethereum-mainnet', 'polygon-mainnet'],
    true
FROM portal_accounts pa
WHERE NOT EXISTS (
    SELECT 1 FROM portal_applications papp 
    WHERE papp.portal_account_id = pa.portal_account_id
);

COMMIT;
EOF
}

# 📈 Function to show summary of inserted data
show_summary() {
    print_status $BLUE "📈 Generating test data summary..."
    
    psql "$DB_CONNECTION_STRING" <<EOF
\echo
\echo '🏢 ORGANIZATIONS:'
SELECT organization_id, organization_name FROM organizations ORDER BY organization_id;

\echo
\echo '📋 PORTAL PLANS:'
SELECT portal_plan_type, portal_plan_type_description, plan_usage_limit FROM portal_plans ORDER BY portal_plan_type;

\echo
\echo '🌐 SERVICES:'
SELECT service_id, service_name, network_id, active FROM services ORDER BY service_id;

\echo
\echo '👥 PORTAL USERS:'
SELECT portal_user_email, signed_up, portal_admin FROM portal_users ORDER BY portal_user_email;

\echo
\echo '💳 PORTAL ACCOUNTS:'
SELECT portal_account_id, user_account_name, portal_plan_type FROM portal_accounts ORDER BY portal_account_id;

\echo
\echo '📱 PORTAL APPLICATIONS:'
SELECT portal_application_id, portal_application_name, emoji FROM portal_applications ORDER BY portal_application_id;
EOF
}

# 🎯 Main execution
main() {
    print_status $PURPLE "🚀 Starting Portal DB Test Data Hydration"
    print_status $PURPLE "============================================"
    
    # Validate database connection
    validate_db_connection
    
    # Insert test data
    insert_test_data
    
    # Show summary
    show_summary
    
    print_status $GREEN "✅ Test data hydration completed successfully!"
    print_status $CYAN "💡 You can now test the API with:"
    print_status $CYAN "   curl http://localhost:3000/networks"
    print_status $CYAN "   curl http://localhost:3000/services"
    print_status $CYAN "   curl http://localhost:3000/portal_plans"
}

# 🏁 Execute main function
main "$@"
