-- ============================================================================
-- GROVE PORTAL FUNCTIONS
-- ============================================================================
-- This file contains PostgreSQL functions that implement business logic for
-- the Grove Portal, replicating functionality from the portal-ui-backend
-- Go codebase for direct database access via PostgREST.
-- ============================================================================

-- ============================================================================
-- USER QUERY FUNCTIONS
-- ============================================================================

-- ----------------------------------------------------------------------------
-- get_portal_user
-- ----------------------------------------------------------------------------
-- Returns the portal user information for the authenticated user based on
-- their Auth0 JWT token. Uses the api.current_portal_user_id() helper from
-- 004_grove_portal_access.sql to extract the portal_user_id from the JWT.
--
-- Returns a single row containing:
--   portal_user_id: The user's portal ID
--   portal_user_email: The user's email address
--   signed_up: Whether the user has completed signup
--   account_permissions: JSONB object where keys are account IDs and values
--                        are objects containing role_name, user_joined_account,
--                        account_name, and permissions array
--
-- Example account_permissions structure:
-- {
--   "account-uuid-1": {
--     "role_name": "OWNER",
--     "user_joined_account": true,
--     "account_name": "My Account",
--     "permissions": ["legacy_read", "legacy_write"]
--   },
--   "account-uuid-2": {
--     "role_name": "MEMBER",
--     "user_joined_account": false,
--     "account_name": "Team Account",
--     "permissions": ["legacy_read"]
--   }
-- }
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.get_portal_user()
RETURNS TABLE (
    portal_user_id VARCHAR(36),
    portal_user_email VARCHAR(255),
    signed_up BOOLEAN,
    account_permissions JSONB
)
AS $$
DECLARE
    v_portal_user_id VARCHAR(36);
BEGIN
    -- Get the portal_user_id from the JWT using the helper function
    v_portal_user_id := api.current_portal_user_id();
    
    IF v_portal_user_id IS NULL THEN
        RAISE EXCEPTION 'No authenticated user found in JWT';
    END IF;
    
    -- Return the user data with aggregated account permissions as JSONB
    RETURN QUERY
    SELECT 
        pu.portal_user_id,
        pu.portal_user_email,
        pu.signed_up,
        COALESCE(
            (
                SELECT jsonb_object_agg(
                    par.portal_account_id,
                    jsonb_build_object(
                        'role_name', par.role_name,
                        'user_joined_account', par.user_joined_account,
                        'account_name', pa.user_account_name,
                        'permissions', COALESCE(r.permissions, ARRAY[]::VARCHAR[])
                    )
                )
                FROM portal_account_rbac par
                LEFT JOIN portal_accounts pa ON pa.portal_account_id = par.portal_account_id
                LEFT JOIN rbac r ON r.role_name = par.role_name
                WHERE par.portal_user_id = pu.portal_user_id
            ),
            '{}'::jsonb
        ) AS account_permissions
    FROM portal_users pu
    WHERE pu.portal_user_id = v_portal_user_id;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION public.get_portal_user IS 
'Returns the current authenticated user information from JWT along with their account permissions. 
Uses api.current_portal_user_id() to extract the portal_user_id from the Auth0 JWT token.
Returns a single row with account_permissions as a JSONB object keyed by account_id.
Each account permission includes role_name, user_joined_account, account_name, and permissions array from the rbac table.';

-- Grant execute permission to authenticated users (they can only see their own data)
GRANT EXECUTE ON FUNCTION public.get_portal_user() TO authenticated_user;

-- Grant execute permission to admin role for testing/debugging
GRANT EXECUTE ON FUNCTION public.get_portal_user() TO portal_db_admin;

-- ============================================================================
-- ADMIN FUNCTIONS
-- ============================================================================

-- ----------------------------------------------------------------------------
-- admin_create_portal_user
-- ----------------------------------------------------------------------------
-- Creates a new portal user with appropriate auth provider and account setup.
-- This function handles 4 different user signup scenarios:
--
-- 1. NEW USER: Brand new user who has never signed up before
--    - Creates new user record
--    - Creates new auth provider entry
--    - Creates default free-tier account
--    - Assigns user as OWNER of new account
--
-- 2. EXISTING USER: User signing up with a different auth provider
--    - Reuses existing user record
--    - Creates new auth provider entry
--    - Does NOT create new account (user keeps existing accounts)
--
-- 3. INVITED USER: User who was invited to a team but hasn't signed up yet
--    - Reuses existing user record (created during invitation)
--    - Creates new auth provider entry
--    - Creates default free-tier account
--    - Assigns user as OWNER of new account
--
-- 4. GCP MARKETPLACE USER: User signing up via GCP Marketplace redirect
--    - Creates new user record
--    - Creates new auth provider entry
--    - Adds user as OWNER to existing GCP account (created by pub/sub)
--    - Does NOT create new account (account already exists from GCP)
--
-- Parameters:
--   p_email: User's email address
--   p_auth_provider_user_id: Auth provider's user ID (e.g., "auth0|abc123")
--   p_gcp_account_id: Optional GCP account ID for marketplace signups
--
-- Returns:
--   portal_user_id: The created or existing user's ID
--   portal_user_email: The user's email address
--   portal_account_id: The account ID (newly created or assigned)
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.admin_create_portal_user(
    p_email VARCHAR(255),
    p_auth_provider_user_id VARCHAR(69),
    p_gcp_account_id VARCHAR(255) DEFAULT NULL
)
RETURNS TABLE (
    out_portal_user_id VARCHAR(36),
    out_portal_user_email VARCHAR(255),
    out_portal_account_id VARCHAR(36)
)
AS $$
DECLARE
    v_user_type TEXT;
    v_portal_user_id VARCHAR(36);
    v_portal_account_id VARCHAR(36);
    v_auth_type portal_auth_type;
    v_auth_provider portal_auth_provider;
    v_is_federated BOOLEAN;
BEGIN
    -- ========================================================================
    -- STEP 1: Determine User Type
    -- ========================================================================
    -- Determine which of the 4 user scenarios applies based on the input
    -- parameters and existing database state.
    
    IF p_gcp_account_id IS NOT NULL THEN
        -- Case: GCP Marketplace user (has gcp_account_id parameter)
        v_user_type := 'gcp';
        
    ELSIF NOT EXISTS (
        SELECT 1 FROM portal_users pu WHERE pu.portal_user_email = p_email
    ) THEN
        -- Case: New user (email doesn't exist in portal_users)
        v_user_type := 'new';
        
    ELSIF EXISTS (
        SELECT 1 FROM portal_users pu
        WHERE pu.portal_user_email = p_email AND pu.signed_up = true
    ) THEN
        -- Case: Existing user (email exists AND signed_up = true)
        v_user_type := 'existing';
        
    ELSIF EXISTS (
        SELECT 1 FROM portal_users pu
        WHERE pu.portal_user_email = p_email AND pu.signed_up = false
    ) THEN
        -- Case: Invited user (email exists AND signed_up = false)
        v_user_type := 'invited';
        
    ELSE
        -- This should never happen, but handle it gracefully
        RAISE EXCEPTION 'Unable to determine user type for email: %', p_email;
    END IF;

    -- ========================================================================
    -- STEP 2: Extract Auth Provider Type
    -- ========================================================================
    -- Parse the auth provider type from the provider_user_id prefix.
    -- Auth0 uses prefixes like "auth0|", "github|", "google-oauth2|", etc.
    
    IF p_auth_provider_user_id LIKE 'auth0|%' THEN
        v_auth_type := 'auth0_username';
        v_auth_provider := 'auth0';
        v_is_federated := false;
        
    ELSIF p_auth_provider_user_id LIKE 'github|%' THEN
        v_auth_type := 'auth0_github';
        v_auth_provider := 'auth0';
        v_is_federated := true;
        
    ELSIF p_auth_provider_user_id LIKE 'google-oauth2|%' THEN
        v_auth_type := 'auth0_google';
        v_auth_provider := 'auth0';
        v_is_federated := true;
        
    ELSE
        RAISE EXCEPTION 'Invalid auth provider user ID format: %', p_auth_provider_user_id;
    END IF;

    -- ========================================================================
    -- STEP 3: Generate or Retrieve User ID
    -- ========================================================================
    -- New and GCP users get a fresh UUID.
    -- Existing and invited users reuse their existing ID.
    
    IF v_user_type IN ('new', 'gcp') THEN
        -- Generate new UUID for new users
        v_portal_user_id := gen_random_uuid()::VARCHAR(36);
        
    ELSE
        -- Reuse existing user ID for existing/invited users
        SELECT pu.portal_user_id INTO v_portal_user_id
        FROM portal_users pu
        WHERE pu.portal_user_email = p_email
        LIMIT 1;
        
        IF v_portal_user_id IS NULL THEN
            RAISE EXCEPTION 'User not found for email: %', p_email;
        END IF;
    END IF;

    -- ========================================================================
    -- STEP 4: Create or Update User Record
    -- ========================================================================
    -- Insert new user or update existing user to signed_up = true.
    -- Uses ON CONFLICT to handle both new and existing users gracefully.
    
    INSERT INTO portal_users (
        portal_user_id, 
        portal_user_email, 
        signed_up, 
        created_at, 
        updated_at
    )
    VALUES (
        v_portal_user_id,
        p_email,
        true,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    )
    ON CONFLICT (portal_user_email) DO UPDATE
    SET signed_up = true,
        updated_at = CURRENT_TIMESTAMP;

    -- ========================================================================
    -- STEP 5: Create Auth Provider Entry
    -- ========================================================================
    -- Insert the auth provider record that links this user to their
    -- authentication provider (Auth0) and method (username, GitHub, Google).
    
    INSERT INTO portal_user_auth (
        portal_user_id,
        portal_auth_provider,
        portal_auth_type,
        auth_provider_user_id,
        federated,
        created_at,
        updated_at
    )
    VALUES (
        v_portal_user_id,
        v_auth_provider,
        v_auth_type,
        p_auth_provider_user_id,
        v_is_federated,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    );

    -- ========================================================================
    -- STEP 6: Handle Account Creation or Assignment
    -- ========================================================================
    -- Based on user type, either create a new account or assign user to
    -- an existing account.
    
    IF v_user_type IN ('new', 'invited') THEN
        -- ====================================================================
        -- Case A: New/Invited Users - Create Default Account
        -- ====================================================================
        -- Check if user already has an account as OWNER. This prevents
        -- creating duplicate accounts for invited users who might already
        -- have been made an owner of another account.
        
        IF NOT EXISTS (
            SELECT 1 FROM portal_account_rbac 
            WHERE portal_user_id = v_portal_user_id 
            AND role_name = 'OWNER'
        ) THEN
            -- Generate new account ID
            v_portal_account_id := gen_random_uuid()::VARCHAR(36);
            
            -- Create new account with free plan
            INSERT INTO portal_accounts (
                portal_account_id,
                portal_plan_type,
                created_at,
                updated_at
            )
            VALUES (
                v_portal_account_id,
                'PLAN_FREE',
                CURRENT_TIMESTAMP,
                CURRENT_TIMESTAMP
            );
            
            -- Add user as OWNER of the new account
            INSERT INTO portal_account_rbac (
                portal_account_id,
                portal_user_id,
                role_name,
                user_joined_account
            )
            VALUES (
                v_portal_account_id,
                v_portal_user_id,
                'OWNER',
                true
            );
        ELSE
            -- User already has an account, retrieve it
            SELECT par.portal_account_id INTO v_portal_account_id
            FROM portal_account_rbac par
            WHERE par.portal_user_id = v_portal_user_id 
            AND par.role_name = 'OWNER'
            LIMIT 1;
        END IF;
        
    ELSIF v_user_type = 'gcp' THEN
        -- ====================================================================
        -- Case B: GCP Marketplace User - Assign to Existing Account
        -- ====================================================================
        -- Look up the account that was created by the GCP pub/sub message
        -- and add this user as the OWNER.
        
        SELECT pa.portal_account_id INTO v_portal_account_id
        FROM portal_accounts pa
        WHERE pa.gcp_account_id = p_gcp_account_id
        LIMIT 1;
        
        IF v_portal_account_id IS NULL THEN
            RAISE EXCEPTION 'GCP account not found for GCP account ID: %', p_gcp_account_id;
        END IF;
        
        -- Add user as OWNER to the GCP account
        INSERT INTO portal_account_rbac (
            portal_account_id,
            portal_user_id,
            role_name,
            user_joined_account
        )
        VALUES (
            v_portal_account_id,
            v_portal_user_id,
            'OWNER',
            true
        );
        
    ELSE
        -- ====================================================================
        -- Case C: Existing User - No New Account
        -- ====================================================================
        -- Existing users don't get a new account. They keep their existing
        -- accounts and are just adding a new authentication method.
        -- Set account_id to NULL to indicate no new account was created.
        v_portal_account_id := NULL;
    END IF;

    -- ========================================================================
    -- STEP 7: Return Results
    -- ========================================================================
    -- Return the user ID, email, and account ID to the caller.
    
    RETURN QUERY
    SELECT 
        v_portal_user_id,
        p_email::VARCHAR(255),
        v_portal_account_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Add descriptive comment for the function
COMMENT ON FUNCTION public.admin_create_portal_user IS 
'Admin function to create a new portal user with auth provider and account setup.
Handles 4 user types: new (creates account), existing (no new account), 
invited (creates account), and GCP marketplace (assigns to existing account).
Returns portal_user_id, portal_user_email, and portal_account_id.';

-- ============================================================================
-- FUNCTION PERMISSIONS
-- ============================================================================
-- Grant execute permissions to PostgREST roles so functions are exposed as
-- RPC endpoints in the API

-- Grant admin role permission to execute this function
GRANT EXECUTE ON FUNCTION public.admin_create_portal_user(VARCHAR, VARCHAR, VARCHAR) TO portal_db_admin;

-- Revoke from public to ensure only authorized roles can execute
REVOKE ALL ON FUNCTION public.admin_create_portal_user(VARCHAR, VARCHAR, VARCHAR) FROM PUBLIC;
