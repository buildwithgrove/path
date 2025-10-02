-- ============================================================================
-- PostgREST API Authentication Setup for Portal DB
-- ============================================================================
-- This file sets up the database roles and permissions for PostgREST JWT authentication
--
-- THREE-ROLE AUTHENTICATION SYSTEM:
-- 1. AUTHENTICATOR: The role PostgREST uses to connect to the database
-- 2. ANON: Role for anonymous/unauthenticated requests (no JWT token)
-- 3. AUTHENTICATED: Role for authenticated requests (valid JWT token)
--
-- JWT AUTHENTICATION FLOW:
-- 1. Client sends request with "Authorization: Bearer <JWT_TOKEN>"
-- 2. PostgREST verifies JWT signature using jwt-secret from postgrest.conf
-- 3. PostgREST extracts 'role' claim from JWT payload
-- 4. PostgREST executes "SET ROLE <extracted_role>;" in database
-- 5. Database query runs with that role's permissions
-- 6. JWT claims available via: current_setting('request.jwt.claims', true)::json
--
-- Reference: https://postgrest.org/en/stable/explanations/db_authz.html

-- ============================================================================
-- CREATE ROLES
-- ============================================================================

-- Authenticator role (PostgREST connection role)
-- NOINHERIT means it starts with no permissions and must switch roles
CREATE ROLE authenticator NOINHERIT LOGIN PASSWORD 'authenticator_password';

-- Anonymous role (public API access)
CREATE ROLE anon NOLOGIN;

-- Authenticated role (logged-in users)
CREATE ROLE authenticated NOLOGIN;

-- ============================================================================
-- SCHEMA PERMISSIONS
-- ============================================================================

-- Grant schema access to both roles
GRANT USAGE ON SCHEMA public, api TO anon, authenticated;

-- Create API schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS api;

-- ============================================================================
-- TABLE PERMISSIONS
-- ============================================================================

-- Anonymous users: read-only access to public data
GRANT SELECT ON TABLE
    networks,
    services,
    service_endpoints,
    service_fallbacks,
    portal_plans
TO anon;

-- Authenticated users: inherit anon permissions + additional tables
GRANT anon TO authenticated;
GRANT SELECT ON TABLE
    organizations,
    portal_accounts,
    portal_applications,
    applications,
    gateways
TO authenticated;

-- ============================================================================
-- JWT UTILITY FUNCTIONS
-- ============================================================================

-- Example function demonstrating JWT claims access
CREATE OR REPLACE FUNCTION public.me()
RETURNS JSON AS $$
BEGIN
    RETURN json_build_object(
        'role', current_setting('request.jwt.claims', true)::json->>'role',
        'email', current_setting('request.jwt.claims', true)::json->>'email'
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

GRANT EXECUTE ON FUNCTION public.me() TO authenticated;

COMMENT ON FUNCTION public.me() IS
'Returns current user info from JWT claims. Demonstrates how to access JWT data in functions.';

-- ============================================================================
-- ROLE SWITCHING GRANTS
-- ============================================================================

-- Allow authenticator to switch to API roles (required for JWT authentication)
GRANT anon TO authenticator;
GRANT authenticated TO authenticator;
