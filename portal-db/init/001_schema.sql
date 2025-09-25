-- ============================================================================
-- PATH Portal Database Schema
-- ============================================================================

-- ============================================================================
-- CUSTOM TYPES
-- ============================================================================

-- Designates the API Types that we support. We can expand this should new interfaces be introduced.
-- Also used as a mapping for the types of QoS that should be attached to a service.
CREATE TYPE endpoint_type AS ENUM ('cometBFT', 'cosmos', 'REST', 'JSON-RPC', 'WSS', 'gRPC');

-- Creates intervals that plans can be evaluated on, also enables users to set their plan limits
CREATE TYPE plan_interval AS ENUM ('day', 'month', 'year');

-- Enables users to limit their application access
-- Service ID - Allow specified list of onchain services
-- Contract - Allow specific smart contracts
-- Origin - Allow specific IP addresses or URLs
CREATE TYPE allowlist_type AS ENUM ('service_id', 'contract', 'origin');

-- Add support for multiple auth providers offering different types
-- Easily extendable should additional authorization providers become necessary 
-- or requested
-- For legacy support, only adding auth0 and its relevant types
CREATE TYPE portal_auth_provider AS ENUM ('auth0');
CREATE TYPE portal_auth_type AS ENUM ('auth0_github', 'auth0_username', 'auth0_google');

-- ============================================================================
-- CORE ORGANIZATIONAL TABLES
-- ============================================================================

-- Organizations table (referenced by contacts and portal_accounts)
-- Organizations are Companies or Customer Groups that can be attached to Portal Accounts
CREATE TABLE organizations (
    organization_id SERIAL PRIMARY KEY,
    organization_name VARCHAR(69) NOT NULL,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE organizations IS 'Companies or customer groups that can be attached to Portal Accounts';
COMMENT ON COLUMN organizations.organization_name IS 'Name of the organization';
COMMENT ON COLUMN organizations.deleted_at IS 'Soft delete timestamp';

-- Organization tags table
-- For categorizing and analyzing organizations
CREATE TABLE organization_tags (
    id SERIAL PRIMARY KEY,
    organization_id INT NOT NULL,
    tag VARCHAR(42),
    FOREIGN KEY (organization_id) REFERENCES organizations(organization_id) ON DELETE CASCADE
);

COMMENT ON TABLE organization_tags IS 'Tags for categorizing and analyzing organizations';

-- ============================================================================
-- PORTAL PLANS AND ACCOUNTS
-- ============================================================================

-- Portal plans table
-- Set of plans that can be assigned to Portal Accounts. i.e. PLAN_FREE, PLAN_UNLIMITED
CREATE TABLE portal_plans (
    portal_plan_type VARCHAR(42) PRIMARY KEY,
    portal_plan_type_description VARCHAR(420),
    plan_usage_limit INT CHECK (plan_usage_limit >= 0),
    plan_usage_limit_interval plan_interval,
    plan_rate_limit_rps INT CHECK (plan_rate_limit_rps >= 0),
    plan_application_limit INT CHECK (plan_application_limit >= 0)
);

COMMENT ON TABLE portal_plans IS 'Available subscription plans for Portal Accounts';
COMMENT ON COLUMN portal_plans.plan_usage_limit IS 'Maximum usage allowed within the interval';
COMMENT ON COLUMN portal_plans.plan_rate_limit_rps IS 'Rate limit in requests per second';

-- Portal accounts table
-- Portal Accounts can have many applications and many users. Only 1 user can be the OWNER.
-- When a new user signs up in the Portal, they automatically generate a personal account.
CREATE TABLE portal_accounts (
    portal_account_id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id INT,
    portal_plan_type VARCHAR(42) NOT NULL,
    user_account_name VARCHAR(42),
    internal_account_name VARCHAR(42),
    portal_account_user_limit INT CHECK (portal_account_user_limit >= 0),
    portal_account_user_limit_interval plan_interval,
    portal_account_user_limit_rps INT CHECK (portal_account_user_limit_rps >= 0),
    billing_type VARCHAR(20),
    stripe_subscription_id VARCHAR(255),
    gcp_account_id VARCHAR(255),
    gcp_entitlement_id VARCHAR(255),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(organization_id),
    FOREIGN KEY (portal_plan_type) REFERENCES portal_plans(portal_plan_type)
);

COMMENT ON TABLE portal_accounts IS 'Multi-tenant accounts with plans and billing integration';
COMMENT ON COLUMN portal_accounts.portal_account_id IS 'Unique identifier for the portal account';
COMMENT ON COLUMN portal_accounts.stripe_subscription_id IS 'Stripe subscription identifier for billing';

-- ============================================================================
-- USER MANAGEMENT
-- ============================================================================

-- Portal users table
-- Users can belong to multiple Accounts
CREATE TABLE portal_users (
    portal_user_id VARCHAR(36) PRIMARY KEY,
    portal_user_email VARCHAR(255) NOT NULL UNIQUE,
    signed_up BOOLEAN DEFAULT FALSE,
    portal_admin BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE portal_users IS 'Users who can access the portal and belong to multiple accounts';
COMMENT ON COLUMN portal_users.portal_user_email IS 'Unique email address for the user';
COMMENT ON COLUMN portal_users.portal_admin IS 'Whether user has admin privileges across the portal';

-- TODO_IMPROVE: Add user_authentication table for password management
-- TODO_CONSIDERATION: Add support for MFA/2FA
-- TODO_CONSIDERATION: Consider session management table

-- Portal User Auth Table
-- Determines which Auth Provider (portal_auth_provider) and which Auth Type 
-- (portal_auth_type) a user is authenticated into the Portal by
CREATE TABLE portal_user_auth (
    portal_user_auth_id SERIAL PRIMARY KEY,
    portal_user_id VARCHAR(36),
    portal_auth_provider portal_auth_provider,
    portal_auth_type portal_auth_type,
    auth_provider_user_id VARCHAR(69),
    federated BOOL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portal_user_id) REFERENCES portal_users(portal_user_id) ON DELETE CASCADE
);

COMMENT ON TABLE portal_user_auth IS 'Authorization provider and type for each user. Determines how to authenticate a user into the Portal.';

-- Contacts table
-- Contacts are individuals that are members of an Organization. Can be attached to Portal Users
CREATE TABLE contacts (
    contact_id SERIAL PRIMARY KEY,
    organization_id INT,
    portal_user_id VARCHAR(36),
    contact_telegram_handle VARCHAR(32),
    contact_twitter_handle VARCHAR(15),
    contact_linkedin_handle VARCHAR(30),
    contact_initial_meeting_location VARCHAR(100),
    contact_initial_meeting_event VARCHAR(100),
    contact_initial_meeting_datetime TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(organization_id) ON DELETE CASCADE,
    FOREIGN KEY (portal_user_id) REFERENCES portal_users(portal_user_id)
);

COMMENT ON TABLE contacts IS 'Contact information for individuals associated with organizations';

-- ============================================================================
-- RBAC (ROLE-BASED ACCESS CONTROL)
-- ============================================================================

-- RBAC table
-- Sets the roles and permissions associated with roles across the Portal
CREATE TABLE rbac (
    role_id SERIAL PRIMARY KEY,
    role_name VARCHAR(20),
    permissions VARCHAR[]
);

COMMENT ON TABLE rbac IS 'Role definitions and their associated permissions';

-- Portal account RBAC table
-- Sets the role and access controls for a user on a particular account.
CREATE TABLE portal_account_rbac (
    id SERIAL PRIMARY KEY,
    portal_account_id VARCHAR(36) NOT NULL,
    portal_user_id VARCHAR(36) NOT NULL,
    role_name VARCHAR(20) NOT NULL,
    user_joined_account BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (portal_account_id) REFERENCES portal_accounts(portal_account_id) ON DELETE CASCADE,
    FOREIGN KEY (portal_user_id) REFERENCES portal_users(portal_user_id),
    UNIQUE (portal_account_id, portal_user_id)
);

COMMENT ON TABLE portal_account_rbac IS 'User roles and permissions for specific portal accounts';

-- ============================================================================
-- APPLICATIONS
-- ============================================================================

-- Portal applications table
-- Portal Accounts can have many Portal Applications that have associated settings.
CREATE TABLE portal_applications (
    portal_application_id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid(),
    portal_account_id VARCHAR(36) NOT NULL,
    portal_application_name VARCHAR(42),
    emoji VARCHAR(16),
    portal_application_user_limit INT CHECK (portal_application_user_limit >= 0),
    portal_application_user_limit_interval plan_interval,
    portal_application_user_limit_rps INT CHECK (portal_application_user_limit_rps >= 0),
    portal_application_description VARCHAR(255),
    favorite_service_ids VARCHAR[],
    secret_key_hash VARCHAR(255), -- TODO_IMPROVE: Never store plain text secrets - use proper hashing
    secret_key_required BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portal_account_id) REFERENCES portal_accounts(portal_account_id) ON DELETE CASCADE
);

COMMENT ON TABLE portal_applications IS 'Applications created within portal accounts with their own rate limits and settings';
COMMENT ON COLUMN portal_applications.secret_key_hash IS 'Hashed secret key for application authentication';

-- TODO_IMPROVE: Add API key rotation history table
-- TODO_CONSIDERATION: Add webhook configurations table

-- Portal application RBAC table
-- Sets the role and access controls for a user on a particular application.
-- Users must be members of the parent Account in order to have access to a particular application
CREATE TABLE portal_application_rbac (
    id SERIAL PRIMARY KEY,
    portal_application_id VARCHAR(36) NOT NULL,
    portal_user_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portal_application_id) REFERENCES portal_applications(portal_application_id) ON DELETE CASCADE,
    FOREIGN KEY (portal_user_id) REFERENCES portal_users(portal_user_id),
    UNIQUE (portal_application_id, portal_user_id)
);

COMMENT ON TABLE portal_application_rbac IS 'User access controls for specific applications';

-- ============================================================================
-- NETWORK AND INFRASTRUCTURE
-- ============================================================================

-- Networks table
-- Future proofing table, but allows the Portal/PATH to send traffic on both Pocket Beta and Pocket Mainnet
CREATE TABLE networks (
    network_id VARCHAR(42) PRIMARY KEY
);

COMMENT ON TABLE networks IS 'Supported blockchain networks (Pocket mainnet, testnet, etc.)';

-- Pavers table
CREATE TABLE pavers (
    paver_id SERIAL PRIMARY KEY,
    paver_url VARCHAR(69) NOT NULL,
    network_id VARCHAR(42) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (network_id) REFERENCES networks(network_id)
);

COMMENT ON TABLE pavers IS 'Paver infrastructure endpoints for different networks';

-- Gateways table
-- Stores the relevant information for the onchain Gateway
CREATE TABLE gateways (
    gateway_address VARCHAR(50) PRIMARY KEY,
    stake_amount BIGINT NOT NULL,
    stake_denom VARCHAR(15) NOT NULL,
    network_id VARCHAR(42) NOT NULL,
    gateway_private_key_hex VARCHAR(64), -- TODO_CONSIDERATION: Store private keys only in encrypted manner and not in plain text
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (network_id) REFERENCES networks(network_id)
);

COMMENT ON TABLE gateways IS 'Onchain gateway information including stake and network details';
COMMENT ON COLUMN gateways.gateway_address IS 'Blockchain address of the gateway';
COMMENT ON COLUMN gateways.stake_amount IS 'Amount of tokens staked by the gateway';

-- ============================================================================
-- SERVICES
-- ============================================================================

-- Services table
-- Stores the set of supported Services from the Pocket Chain
CREATE TABLE services (
    service_id VARCHAR(42) PRIMARY KEY,
    service_name VARCHAR(169) NOT NULL,
    compute_units_per_relay INT,
    service_domains VARCHAR[] NOT NULL,
    service_owner_address VARCHAR(50),
    network_id VARCHAR(42),
    active BOOLEAN DEFAULT FALSE,
    beta BOOLEAN DEFAULT FALSE,
    coming_soon BOOLEAN DEFAULT FALSE,
    quality_fallback_enabled BOOLEAN DEFAULT FALSE,
    hard_fallback_enabled BOOLEAN DEFAULT FALSE,
    svg_icon TEXT, 
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (network_id) REFERENCES networks(network_id)
);

COMMENT ON TABLE services IS 'Supported blockchain services from the Pocket Network';
COMMENT ON COLUMN services.compute_units_per_relay IS 'Cost in compute units for each relay';
COMMENT ON COLUMN services.service_domains IS 'Valid domains for this service';

-- TODO_ITERATE: Consider adding service versioning for API version management

-- Service fallbacks table
-- Defines the set of fallbacks per service for offchain processing.
CREATE TABLE service_fallbacks (
    service_fallback_id SERIAL PRIMARY KEY,
    service_id VARCHAR(42) NOT NULL,
    fallback_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES services(service_id) ON DELETE CASCADE
);

COMMENT ON TABLE service_fallbacks IS 'Fallback URLs for services when primary endpoints fail';

-- Service endpoints table
-- Defines the active list of endpoints per Service. See: endpoint_type for more information
CREATE TABLE service_endpoints (
    endpoint_id SERIAL PRIMARY KEY,
    service_id VARCHAR(42) NOT NULL,
    endpoint_type endpoint_type,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES services(service_id) ON DELETE CASCADE
);

COMMENT ON TABLE service_endpoints IS 'Available endpoint types for each service';

-- ============================================================================
-- ONCHAIN APPLICATIONS
-- ============================================================================

-- Applications table
-- Stores the onchain applications for relay processing. Includes Secret Keys
CREATE TABLE applications (
    application_address VARCHAR(50) PRIMARY KEY,
    gateway_address VARCHAR(50) NOT NULL,
    service_id VARCHAR(42) NOT NULL,
    stake_amount BIGINT,
    stake_denom VARCHAR(15),
    application_private_key_hex VARCHAR(64), -- TODO_IMPROVE: Store private keys in encrypted manner and not in plain text.
    network_id VARCHAR(42) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (gateway_address) REFERENCES gateways(gateway_address),
    FOREIGN KEY (service_id) REFERENCES services(service_id),
    FOREIGN KEY (network_id) REFERENCES networks(network_id)
);

COMMENT ON TABLE applications IS 'Onchain applications for processing relays through the network';
COMMENT ON COLUMN applications.application_address IS 'Blockchain address of the application';

-- ============================================================================
-- ACCESS CONTROL AND SECURITY
-- ============================================================================

-- Portal application allowlists table
-- Sets access controls to Portal Applications based on allowlist_type
CREATE TABLE portal_application_allowlists (
    id SERIAL PRIMARY KEY,
    portal_application_id VARCHAR(36) NOT NULL,
    type allowlist_type,
    value VARCHAR(255),
    service_id VARCHAR(42),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portal_application_id) REFERENCES portal_applications(portal_application_id) ON DELETE CASCADE,
    FOREIGN KEY (service_id) REFERENCES services(service_id)
);

COMMENT ON TABLE portal_application_allowlists IS 'Access control lists for portal applications';

-- Supplier blocklist table
-- Permanently block specific onchain suppliers from processing traffic
CREATE TABLE supplier_blocklist (
    id SERIAL PRIMARY KEY,
    supplier_address VARCHAR(50),
    network_id VARCHAR(42) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (network_id) REFERENCES networks(network_id)
);

COMMENT ON TABLE supplier_blocklist IS 'Blocked supplier addresses to prevent processing';

-- Domain blocklist table
-- Permanently block traffic from being processed by certain domains
CREATE TABLE domain_blocklist (
    id SERIAL PRIMARY KEY,
    domain VARCHAR(169),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE domain_blocklist IS 'Blocked domains to prevent traffic processing';

-- Crypto address blocklist table
-- Permanently block traffic that contains specific addresses.
-- !!! IMPORTANT FOR COMPLIANCE !!!
CREATE TABLE crypto_address_blocklist (
    id SERIAL PRIMARY KEY,
    address VARCHAR(169),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE crypto_address_blocklist IS 'Blocked cryptocurrency addresses for compliance requirements';

-- TODO_IMPROVE: Add audit_logs table for compliance tracking
-- TODO_NEXT_UP: Add rate limit buckets table for global multi-region quota rate limiting
-- TODO_CONSIDERATION: Add usage metrics tables

-- ============================================================================
-- INITIAL DATA
-- ============================================================================

-- Insert default network IDs (Pocket and associated testnets)
INSERT INTO networks (network_id) VALUES
    ('pocket'),
    ('pocket-beta'),
    ('pocket-alpha');

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Core table indexes
CREATE INDEX idx_gateways_network_id ON gateways(network_id);
CREATE INDEX idx_services_network_id ON services(network_id);
CREATE INDEX idx_services_owner ON services(service_owner_address);
CREATE INDEX idx_service_fallbacks_service_id ON service_fallbacks(service_id);
CREATE INDEX idx_service_endpoints_service_id ON service_endpoints(service_id);
CREATE INDEX idx_applications_gateway ON applications(gateway_address);
CREATE INDEX idx_applications_service ON applications(service_id);
CREATE INDEX idx_applications_network ON applications(network_id);
CREATE INDEX idx_portal_applications_account ON portal_applications(portal_account_id);
CREATE INDEX idx_portal_application_allowlists_app ON portal_application_allowlists(portal_application_id);
CREATE INDEX idx_portal_account_rbac_account ON portal_account_rbac(portal_account_id);
CREATE INDEX idx_portal_account_rbac_user ON portal_account_rbac(portal_user_id);
CREATE INDEX idx_portal_application_rbac_app ON portal_application_rbac(portal_application_id);
CREATE INDEX idx_portal_application_rbac_user ON portal_application_rbac(portal_user_id);
CREATE INDEX idx_supplier_blocklist_network ON supplier_blocklist(network_id);
CREATE INDEX idx_contacts_organization ON contacts(organization_id);
CREATE INDEX idx_organization_tags_org ON organization_tags(organization_id);

-- Additional performance indexes
CREATE INDEX idx_portal_users_email ON portal_users(portal_user_email) WHERE deleted_at IS NULL;
CREATE INDEX idx_portal_accounts_org ON portal_accounts(organization_id);
CREATE INDEX idx_portal_accounts_stripe ON portal_accounts(stripe_subscription_id)
  WHERE stripe_subscription_id IS NOT NULL;
CREATE INDEX idx_services_active ON services(service_id) WHERE active = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_portal_applications_active ON portal_applications(portal_account_id)
  WHERE deleted_at IS NULL;
