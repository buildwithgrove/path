-- ============================================================================
-- Create Portal Application Function
-- ============================================================================
-- This function creates a portal application with all associated data in a single
-- atomic transaction. It validates user membership and generates secure API keys.
--
-- ⚠️  WARNING: Currently stores secret keys in plain text (line 64).
--    TODO_IMPROVE: Implement proper hashing before production use.
--
-- Usage via PostgREST:
--   POST /rpc/create_portal_application
--   {
--     "p_portal_account_id": "account-uuid",
--     "p_portal_user_id": "user-uuid",
--     "p_portal_application_name": "My App",
--     "p_secret_key_required": "true"
--   }
--
-- Reference: https://docs.postgrest.org/en/v13/references/transactions.html

CREATE OR REPLACE FUNCTION public.create_portal_application(
    p_portal_account_id UUID,
    p_portal_user_id VARCHAR(36),
    p_portal_application_name VARCHAR(42) DEFAULT NULL,
    p_emoji VARCHAR(16) DEFAULT NULL,
    p_portal_application_user_limit INT DEFAULT NULL,
    p_portal_application_user_limit_interval plan_interval DEFAULT NULL,
    p_portal_application_user_limit_rps INT DEFAULT NULL,
    p_portal_application_description VARCHAR(255) DEFAULT NULL,
    p_favorite_service_ids VARCHAR[] DEFAULT NULL,
    p_secret_key_required TEXT DEFAULT 'false'
) RETURNS JSON AS $$
DECLARE
    v_new_app_id UUID;
    v_secret_key TEXT;
    v_secret_key_hash VARCHAR(255);
    v_secret_key_required_bool BOOLEAN;
BEGIN
    -- ========================================================================
    -- VALIDATION
    -- ========================================================================

    -- Convert text parameter to boolean
    v_secret_key_required_bool := CASE
        WHEN LOWER(p_secret_key_required) IN ('true', 't', '1', 'yes', 'y') THEN TRUE
        ELSE FALSE
    END;

    -- Verify user is a member of the account
    IF NOT EXISTS (
        SELECT 1 FROM portal_account_rbac
        WHERE portal_account_id = p_portal_account_id::VARCHAR
        AND portal_user_id = p_portal_user_id
        AND user_joined_account = TRUE
    ) THEN
        RAISE EXCEPTION 'User % is not a member of account %', p_portal_user_id, p_portal_account_id
            USING ERRCODE = '42501'; -- insufficient_privilege
    END IF;

    -- ========================================================================
    -- SECRET KEY GENERATION
    -- ========================================================================

    -- Generate cryptographically secure secret key (122 bits of entropy)
    -- TODO_IMPROVE: Implement proper hashing (PBKDF2, Argon2, bcrypt)
    v_secret_key := replace(gen_random_uuid()::text, '-', '');
    v_secret_key_hash := v_secret_key; -- Storing plain text for now

    -- ========================================================================
    -- CREATE APPLICATION
    -- ========================================================================

    v_new_app_id := gen_random_uuid();

    INSERT INTO portal_applications (
        portal_application_id,
        portal_account_id,
        portal_application_name,
        emoji,
        portal_application_user_limit,
        portal_application_user_limit_interval,
        portal_application_user_limit_rps,
        portal_application_description,
        favorite_service_ids,
        secret_key_hash,
        secret_key_required,
        created_at,
        updated_at
    ) VALUES (
        v_new_app_id::VARCHAR,
        p_portal_account_id::VARCHAR,
        p_portal_application_name,
        p_emoji,
        p_portal_application_user_limit,
        p_portal_application_user_limit_interval,
        p_portal_application_user_limit_rps,
        p_portal_application_description,
        p_favorite_service_ids,
        v_secret_key_hash,
        v_secret_key_required_bool,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    );

    -- ========================================================================
    -- GRANT USER ACCESS
    -- ========================================================================

    INSERT INTO portal_application_rbac (
        portal_application_id,
        portal_user_id,
        created_at,
        updated_at
    ) VALUES (
        v_new_app_id::VARCHAR,
        p_portal_user_id,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    );

    -- ========================================================================
    -- RETURN RESPONSE
    -- ========================================================================

    RETURN json_build_object(
        'portal_application_id', v_new_app_id,
        'secret_key', v_secret_key,
        'secret_key_required', v_secret_key_required_bool,
        'message', 'Store the secret key securely - it cannot be retrieved again'
    );

END;
$$ LANGUAGE plpgsql VOLATILE SECURITY DEFINER;

-- ============================================================================
-- PERMISSIONS
-- ============================================================================

GRANT EXECUTE ON FUNCTION public.create_portal_application TO authenticated;

COMMENT ON FUNCTION public.create_portal_application IS
'Creates a portal application with RBAC entries in a single transaction.
Validates user membership before creation. Returns app details including the secret key.
Exposed via PostgREST as: POST /rpc/create_portal_application';
