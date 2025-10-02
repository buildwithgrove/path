-- ============================================================================
-- PostgREST API Authentication Setup for Portal DB
-- ============================================================================
-- This file sets up the database roles and permissions for PostgREST JWT authentication

-- ============================================================================
-- CREATE ESSENTIAL POSTGREST ROLES
-- ============================================================================
--
-- THREE-ROLE AUTHENTICATION SYSTEM:
-- 1. AUTHENTICATOR: The role PostgREST uses to connect to the database
--    - Has LOGIN permission to connect
--    - Has NOINHERIT so it starts with no permissions
--    - Can switch to other roles via "SET ROLE" command
--
-- 2. ANON: Role for anonymous/unauthenticated requests
--    - Has NOLOGIN (cannot connect directly)
--    - Limited permissions for public data only
--    - Used when no JWT token is provided
--
-- 3. AUTHENTICATED: Role for authenticated requests
--    - Has NOLOGIN (cannot connect directly)
--    - Extended permissions for user data
--    - Used when JWT contains "role": "authenticated"

-- Create the authenticator role (used by PostgREST to connect)
-- This role can switch to other roles but has no direct permissions
CREATE ROLE authenticator NOINHERIT LOGIN PASSWORD 'authenticator_password';

-- Anonymous role - for public API access (read-only by default)
CREATE ROLE anon NOLOGIN;

-- Authenticated role - for authenticated API access
CREATE ROLE authenticated NOLOGIN;

-- ============================================================================
-- GRANT BASIC PERMISSIONS
-- ============================================================================

-- Grant usage on public schema
GRANT USAGE ON SCHEMA public TO anon, authenticated;

-- Grant basic SELECT permissions for anonymous users (public data only)
GRANT SELECT ON TABLE
    networks,
    services,
    service_endpoints,
    service_fallbacks,
    portal_plans
TO anon;

-- Grant authenticated users access to anon role permissions plus more tables
GRANT anon TO authenticated;
GRANT SELECT ON TABLE
    organizations,
    portal_accounts,
    portal_applications,
    applications,
    gateways
TO authenticated;

-- Create API schema for functions
CREATE SCHEMA IF NOT EXISTS api;
GRANT USAGE ON SCHEMA api TO anon, authenticated;

-- ============================================================================
-- JWT CLAIMS ACCESS EXAMPLE
-- ============================================================================
-- Example function showing how to access JWT claims in SQL
-- Based on PostgREST documentation: https://postgrest.org/en/stable/explanations/db_authz.html
--
-- NOTE: RPC functions must be in 'public' schema for PostgREST to find them
-- PostgREST looks for RPC functions in the public schema by default
--
-- JWT AUTHENTICATION FLOW:
-- 1. Client generates JWT externally
-- 2. Client sends request with "Authorization: Bearer <JWT_TOKEN>"
-- 3. PostgREST verifies JWT signature using jwt-secret from postgrest.conf
-- 4. PostgREST extracts 'role' claim from JWT payload
-- 5. PostgREST executes "SET ROLE <extracted_role>;" in database
-- 6. Database query runs with that role's permissions
-- 7. PostgREST sets JWT claims as transaction-scoped settings for SQL access

-- Function to get current user info from JWT claims
CREATE OR REPLACE FUNCTION public.me()
RETURNS JSON AS $$
BEGIN
    -- Access JWT claims as shown in PostgREST docs
    -- PostgREST automatically sets 'request.jwt.claims' with the JWT payload
    -- current_setting('request.jwt.claims', true)::json->>'claim_name'
    RETURN json_build_object(
        'role', current_setting('request.jwt.claims', true)::json->>'role',
        'email', current_setting('request.jwt.claims', true)::json->>'email'
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Grant execute permission to authenticated users only
GRANT EXECUTE ON FUNCTION public.me TO authenticated;

-- ============================================================================
-- GRANTS FOR AUTHENTICATOR
-- ============================================================================
--
-- CRITICAL: Allow authenticator to "become" other roles
-- When PostgREST receives a JWT with "role": "authenticated", it executes:
--   SET ROLE authenticated;
-- When PostgREST receives no JWT (or invalid JWT), it executes:
--   SET ROLE anon;
--
-- These GRANT statements make the role switching possible:

-- Grant the authenticator role the ability to switch to API roles
GRANT anon TO authenticator;        -- Allows: SET ROLE anon;
GRANT authenticated TO authenticator; -- Allows: SET ROLE authenticated;
