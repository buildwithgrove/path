-- ============================================================================
-- PostgREST Authentication + Authorization Bootstrap for Portal DB
-- ============================================================================
--
-- This migration follows the PostgREST security model:
-- - Keep authorization in PostgreSQL; PostgREST only authenticates requests.
-- - Use three role classes:
--   1. Authenticator role: Chameleon connection role to set application role from JWT.
--   2. Application roles: JWT impersonation targets.
--   3. Anon role: fallback for unauthenticated requests.
-- - Rely on user impersonation via `SET ROLE` after verifying the JWT `role`
--   claim. Note that the authenticator must be granted each impersonated role.
-- - Default the anonymous role to minimal privileges so unauthenticated
--   sessions cannot access data.
--
-- Further reading: https://postgrest.org/en/stable/explanations/auth.html

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================================
-- ROLES
-- ============================================================================

-- PostgREST connection role; configure its password outside of migrations.
CREATE ROLE authenticator WITH NOINHERIT LOGIN;
COMMENT ON ROLE authenticator IS 'PostgREST connection role; password managed outside migrations.';

-- Role assumed for unauthenticated requests (no privileges granted).
CREATE ROLE anon NOLOGIN;
COMMENT ON ROLE anon IS 'Unauthenticated PostgREST role with no privileges.';

-- Read/write application role.
CREATE ROLE portal_db_admin NOLOGIN;
COMMENT ON ROLE portal_db_admin IS 'PostgREST role for administrative clients (read/write).';

-- Read-only application role.
CREATE ROLE portal_db_reader NOLOGIN;
COMMENT ON ROLE portal_db_reader IS 'PostgREST role for read-only clients.';

-- Allow PostgREST to impersonate application roles.
GRANT anon TO authenticator;
GRANT portal_db_admin TO authenticator;
GRANT portal_db_reader TO authenticator;

-- ============================================================================
-- SCHEMAS & DEFAULT PRIVILEGES
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS api;

-- Remove implicit PUBLIC access and grant only what the API roles need.
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA api FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO portal_db_admin, portal_db_reader;
GRANT USAGE ON SCHEMA api TO portal_db_admin, portal_db_reader;

-- Ensure new tables/sequences created under `public` remain private by default.
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON TABLES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON SEQUENCES FROM PUBLIC;

-- Keep the runtime search_path deterministic for the connection role.
ALTER ROLE authenticator SET search_path = 'public, api';

-- ============================================================================
-- TABLE & SEQUENCE PRIVILEGES
-- ============================================================================

REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM PUBLIC;

-- Shared read access to public reference data.
GRANT SELECT ON TABLE
    networks,
    services,
    service_endpoints,
    service_fallbacks,
    portal_plans
TO portal_db_admin, portal_db_reader;

-- Administrative access to mutable business data.
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE
    organizations,
    portal_accounts,
    portal_account_rbac,
    portal_applications,
    portal_application_rbac,
    portal_users
TO portal_db_admin;

-- Read-only access to business data for reader role.
GRANT SELECT ON TABLE
    organizations,
    portal_accounts,
    portal_account_rbac,
    portal_applications,
    portal_application_rbac
TO portal_db_reader;

-- Sequence usage for administrators; readers can observe values if needed.
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO portal_db_admin;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO portal_db_reader;

-- ============================================================================
-- ROW LEVEL SECURITY POLICIES
-- ============================================================================

ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE portal_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE portal_account_rbac ENABLE ROW LEVEL SECURITY;
ALTER TABLE portal_applications ENABLE ROW LEVEL SECURITY;
ALTER TABLE portal_application_rbac ENABLE ROW LEVEL SECURITY;
ALTER TABLE portal_users ENABLE ROW LEVEL SECURITY;

-- Organizations
CREATE POLICY organizations_admin_all ON organizations
    FOR ALL
    TO portal_db_admin
    USING (TRUE)
    WITH CHECK (TRUE);

CREATE POLICY organizations_reader_select ON organizations
    FOR SELECT
    TO portal_db_reader
    USING (TRUE);

-- Portal accounts
CREATE POLICY portal_accounts_admin_all ON portal_accounts
    FOR ALL
    TO portal_db_admin
    USING (TRUE)
    WITH CHECK (TRUE);

CREATE POLICY portal_accounts_reader_select ON portal_accounts
    FOR SELECT
    TO portal_db_reader
    USING (TRUE);

-- Portal account RBAC memberships
CREATE POLICY portal_account_rbac_admin_all ON portal_account_rbac
    FOR ALL
    TO portal_db_admin
    USING (TRUE)
    WITH CHECK (TRUE);

CREATE POLICY portal_account_rbac_reader_select ON portal_account_rbac
    FOR SELECT
    TO portal_db_reader
    USING (TRUE);

-- Portal applications
CREATE POLICY portal_applications_admin_all ON portal_applications
    FOR ALL
    TO portal_db_admin
    USING (TRUE)
    WITH CHECK (TRUE);

CREATE POLICY portal_applications_reader_select ON portal_applications
    FOR SELECT
    TO portal_db_reader
    USING (TRUE);

-- Portal application RBAC memberships
CREATE POLICY portal_application_rbac_admin_all ON portal_application_rbac
    FOR ALL
    TO portal_db_admin
    USING (TRUE)
    WITH CHECK (TRUE);

CREATE POLICY portal_application_rbac_reader_select ON portal_application_rbac
    FOR SELECT
    TO portal_db_reader
    USING (TRUE);

-- Portal users
CREATE POLICY portal_users_admin_all ON portal_users
    FOR ALL
    TO portal_db_admin
    USING (TRUE)
    WITH CHECK (TRUE);