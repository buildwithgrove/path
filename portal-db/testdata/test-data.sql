-- ============================================================================
-- TEST DATA FOR PORTAL DATABASE
-- ============================================================================

-- Insert test portal plans
INSERT INTO portal_plans (portal_plan_type, portal_plan_type_description, plan_usage_limit, plan_usage_limit_interval, plan_rate_limit_rps, plan_application_limit) VALUES
    ('FREE', 'Free tier with basic limits', 1000, 'day', 10, 2),
    ('STARTER', 'Starter plan for small projects', 10000, 'day', 50, 5),
    ('PRO', 'Professional plan for growing businesses', 100000, 'month', 200, 20),
    ('ENTERPRISE', 'Enterprise plan with custom limits', NULL, NULL, 1000, 100);

-- Insert test organizations
INSERT INTO organizations (organization_name) VALUES
    ('Acme Corporation'),
    ('Tech Innovators LLC'),
    ('Blockchain Solutions Inc'),
    ('Web3 Builders Co');

-- Insert test services
INSERT INTO services (service_id, service_name, compute_units_per_relay, service_domains, network_id, active, quality_fallback_enabled, hard_fallback_enabled) VALUES
    ('ethereum-mainnet', 'Ethereum Mainnet', 1, ARRAY['eth-mainnet.gateway.pokt.network'], 'pocket', true, true, false),
    ('ethereum-sepolia', 'Ethereum Sepolia Testnet', 1, ARRAY['eth-sepolia.gateway.pokt.network'], 'pocket', true, false, false),
    ('polygon-mainnet', 'Polygon Mainnet', 1, ARRAY['poly-mainnet.gateway.pokt.network'], 'pocket', true, true, true),
    ('arbitrum-one', 'Arbitrum One', 2, ARRAY['arbitrum-one.gateway.pokt.network'], 'pocket', true, false, false),
    ('base-mainnet', 'Base Mainnet', 2, ARRAY['base-mainnet.gateway.pokt.network'], 'pocket', false, false, false);

-- Insert test service endpoints
INSERT INTO service_endpoints (service_id, endpoint_type) VALUES
    ('ethereum-mainnet', 'JSON-RPC'),
    ('ethereum-mainnet', 'WSS'),
    ('ethereum-sepolia', 'JSON-RPC'),
    ('polygon-mainnet', 'JSON-RPC'),
    ('polygon-mainnet', 'REST'),
    ('arbitrum-one', 'JSON-RPC'),
    ('base-mainnet', 'JSON-RPC');

-- Insert test service fallbacks
INSERT INTO service_fallbacks (service_id, fallback_url) VALUES
    ('ethereum-mainnet', 'https://eth-mainnet.infura.io/v3/fallback'),
    ('ethereum-mainnet', 'https://mainnet.infura.io/v3/backup'),
    ('polygon-mainnet', 'https://polygon-mainnet.infura.io/v3/fallback'),
    ('arbitrum-one', 'https://arbitrum-mainnet.infura.io/v3/fallback');

-- Insert test portal users
INSERT INTO portal_users (portal_user_email, signed_up, portal_admin) VALUES
    ('admin@grove.city', true, true),
    ('alice@acme.com', true, false),
    ('bob@techinnovators.com', true, false),
    ('charlie@blockchain.com', false, false);

-- Insert test portal accounts
INSERT INTO portal_accounts (organization_id, portal_plan_type, user_account_name, internal_account_name, billing_type) VALUES
    (1, 'ENTERPRISE', 'acme-corp', 'internal-acme', 'stripe'),
    (2, 'PRO', 'tech-innovators', 'internal-tech', 'stripe'),
    (3, 'STARTER', 'blockchain-solutions', 'internal-blockchain', 'gcp'),
    (4, 'FREE', 'web3-builders', 'internal-web3', 'stripe');

-- Insert test portal applications
INSERT INTO portal_applications (
    portal_account_id, 
    portal_application_name, 
    emoji, 
    portal_application_description,
    favorite_service_ids,
    secret_key_required
) 
SELECT 
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
FROM portal_accounts pa;
