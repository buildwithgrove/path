-- ============================================================================
-- PostgREST API Setup for Portal DB
-- ============================================================================
-- This file sets up the necessary roles and permissions for PostgREST API access
-- Run this after the main schema (001_schema.sql) has been loaded

-- ============================================================================
-- CREATE API ROLES
-- ============================================================================

-- Anonymous role - for public API access (read-only by default)
DROP ROLE IF EXISTS web_anon;
CREATE ROLE web_anon NOLOGIN;

-- Authenticated role - for authenticated API access
DROP ROLE IF EXISTS web_user;
CREATE ROLE web_user NOLOGIN;

-- Admin role - for administrative API access
DROP ROLE IF EXISTS web_admin;
CREATE ROLE web_admin NOLOGIN;

-- ============================================================================
-- GRANT BASIC PERMISSIONS
-- ============================================================================

-- Grant usage on schema
GRANT USAGE ON SCHEMA public TO web_anon, web_user, web_admin;

-- Grant sequence usage for auto-incrementing IDs
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO web_user, web_admin;

-- ============================================================================
-- READ-ONLY PERMISSIONS (web_anon)
-- ============================================================================

-- Grant SELECT on most tables for anonymous users (public data)
GRANT SELECT ON TABLE 
    networks,
    services,
    service_endpoints,
    service_fallbacks,
    portal_plans
TO web_anon;

-- ============================================================================
-- AUTHENTICATED USER PERMISSIONS (web_user)
-- ============================================================================

-- Inherit anonymous permissions
GRANT web_anon TO web_user;

-- Read access to more tables for authenticated users
GRANT SELECT ON TABLE
    organizations,
    portal_accounts,
    portal_applications,
    applications,
    gateways
TO web_user;

-- Limited write access for user-owned resources
-- Users can only modify their own data (enforced by RLS policies)
GRANT INSERT, UPDATE ON TABLE
    portal_users,
    contacts,
    portal_accounts,
    portal_applications,
    portal_application_allowlists
TO web_user;

-- ============================================================================
-- ADMIN PERMISSIONS (web_admin)
-- ============================================================================

-- Inherit user permissions
GRANT web_user TO web_admin;

-- Full access to all tables for admin users
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO web_admin;

-- ============================================================================
-- ROW LEVEL SECURITY (RLS) POLICIES
-- ============================================================================

-- Enable RLS on sensitive tables
ALTER TABLE portal_users ENABLE ROW LEVEL SECURITY;
ALTER TABLE portal_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE portal_applications ENABLE ROW LEVEL SECURITY;
ALTER TABLE contacts ENABLE ROW LEVEL SECURITY;
ALTER TABLE portal_account_rbac ENABLE ROW LEVEL SECURITY;
ALTER TABLE portal_application_rbac ENABLE ROW LEVEL SECURITY;
ALTER TABLE portal_application_allowlists ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- RLS POLICIES FOR PORTAL_USERS
-- ============================================================================

-- Users can view their own profile
CREATE POLICY "Users can view own profile" ON portal_users
    FOR SELECT
    USING (portal_user_email = current_setting('request.jwt.claims', true)::json->>'email');

-- Users can update their own profile
CREATE POLICY "Users can update own profile" ON portal_users
    FOR UPDATE
    USING (portal_user_email = current_setting('request.jwt.claims', true)::json->>'email');

-- Admins can view all users
CREATE POLICY "Admins can view all users" ON portal_users
    FOR ALL
    TO web_admin
    USING (true);

-- ============================================================================
-- RLS POLICIES FOR PORTAL_ACCOUNTS
-- ============================================================================

-- Users can view accounts they have access to
CREATE POLICY "Users can view accessible accounts" ON portal_accounts
    FOR SELECT
    USING (
        portal_account_id IN (
            SELECT pa.portal_account_id 
            FROM portal_account_rbac pa
            JOIN portal_users pu ON pa.portal_user_id = pu.portal_user_id
            WHERE pu.portal_user_email = current_setting('request.jwt.claims', true)::json->>'email'
        )
    );

-- Account owners can update their accounts
CREATE POLICY "Account owners can update accounts" ON portal_accounts
    FOR UPDATE
    USING (
        portal_account_id IN (
            SELECT pa.portal_account_id 
            FROM portal_account_rbac pa
            JOIN portal_users pu ON pa.portal_user_id = pu.portal_user_id
            WHERE pu.portal_user_email = current_setting('request.jwt.claims', true)::json->>'email'
            AND pa.role_name = 'owner'
        )
    );

-- ============================================================================
-- RLS POLICIES FOR PORTAL_APPLICATIONS
-- ============================================================================

-- Users can view applications they have access to
CREATE POLICY "Users can view accessible applications" ON portal_applications
    FOR SELECT
    USING (
        portal_account_id IN (
            SELECT pa.portal_account_id 
            FROM portal_account_rbac pa
            JOIN portal_users pu ON pa.portal_user_id = pu.portal_user_id
            WHERE pu.portal_user_email = current_setting('request.jwt.claims', true)::json->>'email'
        )
    );

-- Users can create applications in accounts they have access to
CREATE POLICY "Users can create applications in accessible accounts" ON portal_applications
    FOR INSERT
    WITH CHECK (
        portal_account_id IN (
            SELECT pa.portal_account_id 
            FROM portal_account_rbac pa
            JOIN portal_users pu ON pa.portal_user_id = pu.portal_user_id
            WHERE pu.portal_user_email = current_setting('request.jwt.claims', true)::json->>'email'
        )
    );

-- Users can update applications they have access to
CREATE POLICY "Users can update accessible applications" ON portal_applications
    FOR UPDATE
    USING (
        portal_account_id IN (
            SELECT pa.portal_account_id 
            FROM portal_account_rbac pa
            JOIN portal_users pu ON pa.portal_user_id = pu.portal_user_id
            WHERE pu.portal_user_email = current_setting('request.jwt.claims', true)::json->>'email'
        )
    );

-- ============================================================================
-- API HELPER FUNCTIONS
-- ============================================================================

-- Function to get current user info
CREATE OR REPLACE FUNCTION api.current_user_info()
RETURNS TABLE (
    user_id INTEGER,
    email VARCHAR(255),
    is_admin BOOLEAN
)
LANGUAGE sql STABLE SECURITY DEFINER
AS $$
    SELECT 
        pu.portal_user_id,
        pu.portal_user_email,
        pu.portal_admin
    FROM portal_users pu
    WHERE pu.portal_user_email = current_setting('request.jwt.claims', true)::json->>'email';
$$;

-- Function to get user's accessible accounts
CREATE OR REPLACE FUNCTION api.user_accounts()
RETURNS TABLE (
    account_id UUID,
    account_name VARCHAR(42),
    role_name VARCHAR(20),
    plan_type VARCHAR(42)
)
LANGUAGE sql STABLE SECURITY DEFINER
AS $$
    SELECT 
        pa.portal_account_id,
        pa.user_account_name,
        par.role_name,
        pa.portal_plan_type
    FROM portal_accounts pa
    JOIN portal_account_rbac par ON pa.portal_account_id = par.portal_account_id
    JOIN portal_users pu ON par.portal_user_id = pu.portal_user_id
    WHERE pu.portal_user_email = current_setting('request.jwt.claims', true)::json->>'email';
$$;

-- Function to get user's accessible applications
CREATE OR REPLACE FUNCTION api.user_applications()
RETURNS TABLE (
    application_id UUID,
    application_name VARCHAR(42),
    account_id UUID,
    account_name VARCHAR(42)
)
LANGUAGE sql STABLE SECURITY DEFINER
AS $$
    SELECT 
        pa.portal_application_id,
        pa.portal_application_name,
        pac.portal_account_id,
        pac.user_account_name
    FROM portal_applications pa
    JOIN portal_accounts pac ON pa.portal_account_id = pac.portal_account_id
    JOIN portal_account_rbac par ON pac.portal_account_id = par.portal_account_id
    JOIN portal_users pu ON par.portal_user_id = pu.portal_user_id
    WHERE pu.portal_user_email = current_setting('request.jwt.claims', true)::json->>'email'
    AND pa.deleted_at IS NULL;
$$;

-- ============================================================================
-- CREATE API SCHEMA FOR FUNCTIONS
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS api;
GRANT USAGE ON SCHEMA api TO web_anon, web_user, web_admin;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA api TO web_anon, web_user, web_admin;

-- ============================================================================
-- GRANTS FOR API SCHEMA
-- ============================================================================

-- Ensure the authenticator role can switch to our API roles
-- Note: This will be handled by your JWT authentication setup
GRANT web_anon TO current_user;
GRANT web_user TO current_user;
GRANT web_admin TO current_user;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON ROLE web_anon IS 'Anonymous role for public API access (read-only)';
COMMENT ON ROLE web_user IS 'Authenticated role for standard API access';
COMMENT ON ROLE web_admin IS 'Administrative role for full API access';
COMMENT ON SCHEMA api IS 'API functions and procedures for PostgREST';
