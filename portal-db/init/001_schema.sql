-- Initial schema for PATH Portal DB

-- TODO: Add more comments to tables so the business logic is more clear

-- Create custom enum types

-- Designates the API Types that we support. We can expand this should new interfaces be introduced.
CREATE TYPE endpoint_type AS ENUM ('cometBFT', 'cosmos', 'REST', 'JSON-RPC', 'WSS', 'gRPC');
-- Creates intervals that plans can be evaluated on, also enables users to set their plan limits
CREATE TYPE plan_interval AS ENUM ('day', 'month', 'year');
-- Enables users to limit their application access
-- Service ID - Allow specified list of onchain services
-- Contract - Allow specific smart contracts
-- Origin - Allow specific IP addresses or URLs
CREATE TYPE whitelist_type AS ENUM ('service_id', 'contract', 'origin');

-- Organizations table (referenced by contacts and portal_accounts)
-- Organizations are Companies or Customer Groups that can be attached to Portal Accounts
CREATE TABLE organizations (
    organization_id SERIAL PRIMARY KEY,
    organization_name VARCHAR(69),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Organization tags table
-- For categorizing and analyzing organizations
CREATE TABLE organization_tags (
    id SERIAL PRIMARY KEY,
    organization_id INT NOT NULL,
    tag VARCHAR(42),
    FOREIGN KEY (organization_id) REFERENCES organizations(organization_id)
);

-- Portal plans table
-- Set of plans that can be assigned to Portal Accounts. i.e. PLAN_FREE, PLAN_UNLIMITED
CREATE TABLE portal_plans (
    portal_plan_type VARCHAR(42) PRIMARY KEY,
    portal_plan_type_description VARCHAR(420),
    plan_usage_limit INT,
    plan_usage_limit_interval plan_interval,
    plan_rate_limit_rps INT,
    plan_application_limit INT
);

-- Portal accounts table
-- Portal Accounts can have many applications and many users. Only 1 user can be the OWNER.
-- When a new user signs up in the Portal, they automatically generate a personal account.
CREATE TABLE portal_accounts (
    portal_account_id VARCHAR(24) PRIMARY KEY,
    organization_id INT,
    portal_plan_type VARCHAR(42) NOT NULL,
    user_account_name VARCHAR(42),
    internal_account_name VARCHAR(42),
    portal_account_user_limit INT,
    portal_account_user_limit_interval plan_interval,
    portal_account_user_limit_rps INT,
    billing_type VARCHAR(20),
    stripe_subscription_id VARCHAR(255),
    gcp_account_id VARCHAR(255),
    gcp_entitlement_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(organization_id),
    FOREIGN KEY (portal_plan_type) REFERENCES portal_plans(portal_plan_type)
);

-- Portal users table
-- Users can belong to multiple Accounts
CREATE TABLE portal_users (
    portal_user_id SERIAL PRIMARY KEY,
    portal_user_email VARCHAR(255),
    signed_up BOOLEAN DEFAULT FALSE,
    portal_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Contacts table
-- Contacts are individuals that are members of an Organization. Can be attached to Portal Users
CREATE TABLE contacts (
    contact_id SERIAL PRIMARY KEY,
    organization_id INT,
    portal_user_id INT,
    contact_telegram_handle VARCHAR(32),
    contact_twitter_handle VARCHAR(15),
    contact_linkedin_handle VARCHAR(30),
    contact_initial_meeting_location VARCHAR(100),
    contact_initial_meeting_event VARCHAR(100),
    contact_initial_meeting_datetime TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(organization_id),
    FOREIGN KEY (portal_user_id) REFERENCES portal_users(portal_user_id)
);

-- RBAC table
-- Sets the roles and permissions associated with roles across the Portal
CREATE TABLE rbac (
    role_id SERIAL PRIMARY KEY,
    role_name VARCHAR(20),
    permissions VARCHAR[]
);

-- Portal account RBAC table
-- Sets the role and access controls for a user on a particular account.
CREATE TABLE portal_account_rbac (
    id SERIAL PRIMARY KEY,
    portal_account_id VARCHAR(24) NOT NULL,
    portal_user_id INT NOT NULL,
    role_name VARCHAR(20) NOT NULL,
    user_joined_account BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (portal_account_id) REFERENCES portal_accounts(portal_account_id),
    FOREIGN KEY (portal_user_id) REFERENCES portal_users(portal_user_id)
);

-- Portal applications table
-- Portal Accounts can have many Portal Applications that have associated settings.
CREATE TABLE portal_applications (
    portal_application_id VARCHAR(42) PRIMARY KEY,
    portal_account_id VARCHAR(24) NOT NULL,
    portal_application_name VARCHAR(42),
    emoji VARCHAR(16),
    portal_application_user_limit INT,
    portal_application_user_limit_interval plan_interval,
    portal_application_user_limit_rps INT,
    portal_application_description VARCHAR(255),
    favorite_service_ids VARCHAR[],
    secret_key VARCHAR(64),
    secret_key_required BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portal_account_id) REFERENCES portal_accounts(portal_account_id)
);

-- Portal application RBAC table
-- Sets the role and access controls for a user on a particular application. 
-- Users must be members of the parent Account in order to have access to a particular application
CREATE TABLE portal_application_rbac (
    id SERIAL PRIMARY KEY,
    portal_application_id VARCHAR(42) NOT NULL,
    portal_user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portal_application_id) REFERENCES portal_applications(portal_application_id),
    FOREIGN KEY (portal_user_id) REFERENCES portal_users(portal_user_id)
);

-- Networks table
-- Future proofing table, but allows the Portal/PATH to send traffic on both Pocket Beta and Pocket Mainnet
CREATE TABLE networks (
    network_id VARCHAR(42) PRIMARY KEY
);

CREATE TABLE pavers (
    paver_id SERIAL PRIMARY KEY,
    paver_url VARCHAR(69) NOT NULL,
    network_id VARCHAR(42) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (network_id) REFERENCES networks(network_id)
);

-- Gateways table
-- Stores the relevant information for the onchain Gateway
CREATE TABLE gateways (
    gateway_address VARCHAR(50) PRIMARY KEY,
    stake_amount INT NOT NULL,
    stake_denom VARCHAR(15) NOT NULL,
    network_id VARCHAR(42) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (network_id) REFERENCES networks(network_id)
);

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
    quality_fallback_enabled BOOLEAN DEFAULT FALSE,
    hard_fallback_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (network_id) REFERENCES networks(network_id),
    FOREIGN KEY (service_owner_address) REFERENCES gateways(gateway_address)
);

-- Service fallbacks table
-- Defines the set of fallbacks per service for offchain processing.
CREATE TABLE service_fallbacks (
    service_fallback_id SERIAL PRIMARY KEY,
    service_id VARCHAR(42) NOT NULL,
    fallback_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES services(service_id)
);

-- Service endpoints table
-- Defines the active list of endpoints per Service. See: endpoint_type for more information
CREATE TABLE service_endpoints (
    endpoint_id SERIAL PRIMARY KEY,
    service_id VARCHAR(42) NOT NULL,
    endpoint_type endpoint_type,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES services(service_id)
);

-- Applications table
-- Stores the onchain applications for relay processing. Includes Secret Keys
CREATE TABLE applications (
    application_address VARCHAR(50) PRIMARY KEY,
    gateway_address VARCHAR(50) NOT NULL,
    service_id VARCHAR(42) NOT NULL,
    stake_amount INT,
    stake_denom VARCHAR(15),
    application_private_key_hex VARCHAR(64) NOT NULL,
    network_id VARCHAR(42) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (gateway_address) REFERENCES gateways(gateway_address),
    FOREIGN KEY (service_id) REFERENCES services(service_id),
    FOREIGN KEY (network_id) REFERENCES networks(network_id)
);

-- Portal application whitelists table
-- Sets access controls to Portal Applications based on whitelist_type
CREATE TABLE portal_application_whitelists (
    id SERIAL PRIMARY KEY,
    portal_application_id VARCHAR(42) NOT NULL,
    type whitelist_type,
    value VARCHAR(255),
    service_id VARCHAR(42),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portal_application_id) REFERENCES portal_applications(portal_application_id),
    FOREIGN KEY (service_id) REFERENCES services(service_id)
);

-- Supplier blacklist table
-- Permanently block specific onchain suppliers from processing traffic
CREATE TABLE supplier_blacklist (
    id SERIAL PRIMARY KEY,
    supplier_address VARCHAR(50),
    network_id VARCHAR(42) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (network_id) REFERENCES networks(network_id)
);

-- Domain blocklist table
-- Permanently block traffic from being processed by certain domains
CREATE TABLE domain_blacklist (
    id SERIAL PRIMARY KEY,
    domain VARCHAR(169),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Crypto address blocklist table
-- Permanently block traffic that contains specific addresses. 
-- !!! IMPORTANT FOR COMPLIANCE !!!
CREATE TABLE crypto_address_blacklist (
    id SERIAL PRIMARY KEY,
    address VARCHAR(169),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indeces for better performance
CREATE INDEX idx_gateways_network_id ON gateways(network_id);
CREATE INDEX idx_services_network_id ON services(network_id);
CREATE INDEX idx_services_owner ON services(service_owner_address);
CREATE INDEX idx_service_fallbacks_service_id ON service_fallbacks(service_id);
CREATE INDEX idx_service_endpoints_service_id ON service_endpoints(service_id);
CREATE INDEX idx_applications_gateway ON applications(gateway_address);
CREATE INDEX idx_applications_service ON applications(service_id);
CREATE INDEX idx_applications_network ON applications(network_id);
CREATE INDEX idx_portal_applications_account ON portal_applications(portal_account_id);
CREATE INDEX idx_portal_application_whitelists_app ON portal_application_whitelists(portal_application_id);
CREATE INDEX idx_portal_account_rbac_account ON portal_account_rbac(portal_account_id);
CREATE INDEX idx_portal_account_rbac_user ON portal_account_rbac(portal_user_id);
CREATE INDEX idx_portal_application_rbac_app ON portal_application_rbac(portal_application_id);
CREATE INDEX idx_portal_application_rbac_user ON portal_application_rbac(portal_user_id);
CREATE INDEX idx_supplier_blacklist_network ON supplier_blacklist(network_id);
CREATE INDEX idx_contacts_organization ON contacts(organization_id);
CREATE INDEX idx_organization_tags_org ON organization_tags(organization_id);
