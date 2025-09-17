-- ============================================================================
-- PostgREST API Setup for Portal DB
-- ============================================================================
-- Minimal setup for PostgREST API access

-- ============================================================================
-- CREATE ESSENTIAL POSTGREST ROLES
-- ============================================================================

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
-- GRANTS FOR AUTHENTICATOR
-- ============================================================================

-- Grant the authenticator role the ability to switch to API roles
GRANT anon TO authenticator;
GRANT authenticated TO authenticator;
