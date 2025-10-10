-- ============================================================================
-- AUTH0 JWT INTEGRATION FOR FRONTEND USER ACCESS
-- ============================================================================
-- This migration adds user-scoped access control for frontend users
-- authenticating via Auth0. Existing roles (portal_db_admin, portal_db_reader)
-- remain unchanged and are used by backend services.
--
-- Overview:
-- - Creates 'authenticated_user' role for Auth0-authenticated frontend users
-- - Implements Row-Level Security (RLS) policies based on portal_account_rbac
-- - Users can only access portal accounts/applications they have RBAC permissions for
-- - Permission levels: 'legacy_read' (SELECT) and 'legacy_write' (INSERT/UPDATE/DELETE)
-- ============================================================================

-- ============================================================================
-- ROLES
-- ============================================================================

-- New role for Auth0-authenticated frontend users
CREATE ROLE authenticated_user NOLOGIN;
COMMENT ON ROLE authenticated_user IS 'Role for Auth0-authenticated frontend users with user-scoped RLS policies';

-- Allow PostgREST authenticator to impersonate this role
GRANT authenticated_user TO authenticator;

-- ============================================================================
-- SCHEMA ACCESS
-- ============================================================================

-- Grant schema access
GRANT USAGE ON SCHEMA public, api TO authenticated_user;

-- ============================================================================
-- TABLE PERMISSIONS
-- ============================================================================

-- Grant table access (users will only be able to see/modify their own data via RLS)
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE
    portal_accounts,
    portal_account_rbac,
    portal_applications,
    portal_users
TO authenticated_user;

-- Grant access to public tables used by the UI (read-only, contains no sensitive data)
GRANT SELECT ON TABLE
    services,
    portal_plans,
TO authenticated_user;

-- Grant access to portal_user_auth for JWT sub claim lookup
GRANT SELECT ON TABLE portal_user_auth TO authenticated_user;

-- Grant sequence access for inserts
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO authenticated_user;

-- ============================================================================
-- HELPER FUNCTION: Extract portal_user_id from Auth0 JWT sub claim
-- ============================================================================

CREATE OR REPLACE FUNCTION api.current_portal_user_id()
RETURNS VARCHAR(36) AS $$
DECLARE
  auth_sub TEXT;
  user_id VARCHAR(36);
BEGIN
  -- Extract 'sub' claim from JWT (e.g., "auth0|6536b4a897072b320a2d41ea")
  auth_sub := current_setting('request.jwt.claims', true)::json->>'sub';
  
  IF auth_sub IS NULL THEN
    RETURN NULL;
  END IF;
  
  -- Look up portal_user_id from portal_user_auth
  SELECT pua.portal_user_id INTO user_id
  FROM portal_user_auth pua
  WHERE pua.auth_provider_user_id = auth_sub
  LIMIT 1;
  
  RETURN user_id;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION api.current_portal_user_id() IS 'Extracts portal_user_id from Auth0 JWT sub claim by looking up auth_provider_user_id';

-- ============================================================================
-- RLS POLICIES: portal_accounts (user-scoped access)
-- ============================================================================

-- Users can SELECT accounts they have RBAC access to
CREATE POLICY portal_accounts_user_select ON portal_accounts
    FOR SELECT
    TO authenticated_user
    USING (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            WHERE par.portal_user_id = api.current_portal_user_id()
        )
    );

-- Users can INSERT new accounts (WITH CHECK will verify they have write permission)
CREATE POLICY portal_accounts_user_insert ON portal_accounts
    FOR INSERT
    TO authenticated_user
    WITH CHECK (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            JOIN rbac r ON r.role_name = par.role_name
            WHERE par.portal_user_id = api.current_portal_user_id()
              AND 'legacy_write' = ANY(r.permissions)
        )
    );

-- Users can UPDATE accounts where they have 'legacy_write' permission
CREATE POLICY portal_accounts_user_update ON portal_accounts
    FOR UPDATE
    TO authenticated_user
    USING (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            JOIN rbac r ON r.role_name = par.role_name
            WHERE par.portal_user_id = api.current_portal_user_id()
              AND 'legacy_write' = ANY(r.permissions)
        )
    )
    WITH CHECK (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            JOIN rbac r ON r.role_name = par.role_name
            WHERE par.portal_user_id = api.current_portal_user_id()
              AND 'legacy_write' = ANY(r.permissions)
        )
    );

-- Users can DELETE accounts where they have 'legacy_write' permission
CREATE POLICY portal_accounts_user_delete ON portal_accounts
    FOR DELETE
    TO authenticated_user
    USING (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            JOIN rbac r ON r.role_name = par.role_name
            WHERE par.portal_user_id = api.current_portal_user_id()
              AND 'legacy_write' = ANY(r.permissions)
        )
    );

-- ============================================================================
-- RLS POLICIES: portal_applications (user-scoped access)
-- ============================================================================
-- Note: Application access inherits from parent account permissions

-- Users can SELECT applications if they have access to the parent account
CREATE POLICY portal_applications_user_select ON portal_applications
    FOR SELECT
    TO authenticated_user
    USING (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            WHERE par.portal_user_id = api.current_portal_user_id()
        )
    );

-- Users can INSERT applications if they have 'legacy_write' on parent account
CREATE POLICY portal_applications_user_insert ON portal_applications
    FOR INSERT
    TO authenticated_user
    WITH CHECK (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            JOIN rbac r ON r.role_name = par.role_name
            WHERE par.portal_user_id = api.current_portal_user_id()
              AND 'legacy_write' = ANY(r.permissions)
        )
    );

-- Users can UPDATE applications if they have 'legacy_write' on parent account
CREATE POLICY portal_applications_user_update ON portal_applications
    FOR UPDATE
    TO authenticated_user
    USING (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            JOIN rbac r ON r.role_name = par.role_name
            WHERE par.portal_user_id = api.current_portal_user_id()
              AND 'legacy_write' = ANY(r.permissions)
        )
    )
    WITH CHECK (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            JOIN rbac r ON r.role_name = par.role_name
            WHERE par.portal_user_id = api.current_portal_user_id()
              AND 'legacy_write' = ANY(r.permissions)
        )
    );

-- Users can DELETE applications if they have 'legacy_write' on parent account
CREATE POLICY portal_applications_user_delete ON portal_applications
    FOR DELETE
    TO authenticated_user
    USING (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            JOIN rbac r ON r.role_name = par.role_name
            WHERE par.portal_user_id = api.current_portal_user_id()
              AND 'legacy_write' = ANY(r.permissions)
        )
    );

-- ============================================================================
-- RLS POLICIES: portal_account_rbac (user-scoped access)
-- ============================================================================

-- Users can view their own RBAC entries
-- IMPORTANT: This policy must be simple to avoid infinite recursion when other
-- tables query portal_account_rbac. Users can only see RBAC rows that directly
-- reference their portal_user_id.
CREATE POLICY portal_account_rbac_user_select ON portal_account_rbac
    FOR SELECT
    TO authenticated_user
    USING (portal_user_id = api.current_portal_user_id());

-- Users can INSERT RBAC entries if they have 'legacy_write' on the account
CREATE POLICY portal_account_rbac_user_insert ON portal_account_rbac
    FOR INSERT
    TO authenticated_user
    WITH CHECK (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            JOIN rbac r ON r.role_name = par.role_name
            WHERE par.portal_user_id = api.current_portal_user_id()
              AND 'legacy_write' = ANY(r.permissions)
        )
    );

-- Users can UPDATE RBAC entries if they have 'legacy_write' on the account
CREATE POLICY portal_account_rbac_user_update ON portal_account_rbac
    FOR UPDATE
    TO authenticated_user
    USING (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            JOIN rbac r ON r.role_name = par.role_name
            WHERE par.portal_user_id = api.current_portal_user_id()
              AND 'legacy_write' = ANY(r.permissions)
        )
    )
    WITH CHECK (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            JOIN rbac r ON r.role_name = par.role_name
            WHERE par.portal_user_id = api.current_portal_user_id()
              AND 'legacy_write' = ANY(r.permissions)
        )
    );

-- Users can DELETE RBAC entries if they have 'legacy_write' on the account
CREATE POLICY portal_account_rbac_user_delete ON portal_account_rbac
    FOR DELETE
    TO authenticated_user
    USING (
        portal_account_id IN (
            SELECT par.portal_account_id
            FROM portal_account_rbac par
            JOIN rbac r ON r.role_name = par.role_name
            WHERE par.portal_user_id = api.current_portal_user_id()
              AND 'legacy_write' = ANY(r.permissions)
        )
    );

-- ============================================================================
-- RLS POLICIES: portal_users (self-access only)
-- ============================================================================

-- Users can view their own user record
CREATE POLICY portal_users_self_select ON portal_users
    FOR SELECT
    TO authenticated_user
    USING (portal_user_id = api.current_portal_user_id());

-- Users can update their own record
CREATE POLICY portal_users_self_update ON portal_users
    FOR UPDATE
    TO authenticated_user
    USING (portal_user_id = api.current_portal_user_id())
    WITH CHECK (portal_user_id = api.current_portal_user_id());
