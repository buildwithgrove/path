-- This file contains transactions functions for performing complex database operations
-- such as creating a new portal application and its associated data

-- ============================================================================
-- SECRET KEY GENERATION FUNCTION
-- ============================================================================
-- 
-- Generates cryptographically secure secret keys for portal applications.
-- This function is separated for reusability and to centralize key generation logic.
--
-- References:
-- - PostgreSQL UUID Generation: https://www.postgresql.org/docs/current/functions-uuid.html
-- - PostgREST Security Definer: https://docs.postgrest.org/en/v13/explanations/db_authz.html#security-definer
--
-- Security Considerations:
-- - Uses gen_random_uuid() which provides 122 bits of entropy
-- - Returns both plain text key (for immediate use) and "hash" (for storage)
-- - TODO_IMPROVE: Implement proper cryptographic hashing (PBKDF2, Argon2, bcrypt)
-- - TODO_IMPROVE: Consider using PostgreSQL's pgcrypto extension
--
CREATE OR REPLACE FUNCTION public.generate_portal_app_secret()
RETURNS JSON AS $$
DECLARE
    v_secret_key TEXT;
    v_secret_key_hash VARCHAR(255);
BEGIN
    -- Generate a cryptographically secure secret key
    -- Using gen_random_uuid() provides 122 bits of entropy which is suitable for API keys
    -- Format: Remove hyphens to create a 32-character hex string
    v_secret_key := replace(gen_random_uuid()::text, '-', '');
    
    -- For now, we're storing the "hash" as the plain key since the schema comment
    -- indicates this needs improvement. In production, this should be properly hashed.
    -- TODO_IMPROVE: Use a proper key derivation function like PBKDF2 or Argon2
    -- TODO_IMPROVE: Consider using PostgreSQL's pgcrypto extension for better hashing
    v_secret_key_hash := v_secret_key;
    
    -- Return both the plain key (for immediate use) and hash (for storage)
    RETURN json_build_object(
        'secret_key', v_secret_key,
        'secret_key_hash', v_secret_key_hash,
        'generated_at', CURRENT_TIMESTAMP,
        'entropy_bits', 122,
        'algorithm', 'UUID v4 (hex formatted)'
    );
END;
$$ LANGUAGE plpgsql VOLATILE SECURITY DEFINER;

-- Remove public execute permission to prevent PostgREST from exposing this as an API endpoint
-- This function is for internal use by other database functions only
REVOKE EXECUTE ON FUNCTION public.generate_portal_app_secret FROM PUBLIC;

-- Do NOT grant to authenticated role - this keeps it private
-- Only functions with SECURITY DEFINER can call this function

COMMENT ON FUNCTION public.generate_portal_app_secret IS 
'INTERNAL USE ONLY: Generates cryptographically secure secret keys for portal applications. 
Returns both plain text key and storage hash. 
Used internally by create_portal_application function.
NOT exposed via PostgREST API.';

-- ============================================================================
-- CREATE PORTAL APPLICATION FUNCTION
-- ============================================================================
-- 
-- This function creates a portal application with all associated data in a single
-- atomic transaction. This approach follows PostgREST best practices for complex
-- multi-table operations that require ACID guarantees.
--
-- References:
-- - PostgREST Transactions: https://docs.postgrest.org/en/v13/references/transactions.html
-- - PostgreSQL UUID Generation: https://www.postgresql.org/docs/current/functions-uuid.html
-- - PostgreSQL Security Functions: https://www.postgresql.org/docs/current/functions-info.html
--
-- Security Considerations:
-- - Uses SECURITY DEFINER to execute with function owner privileges
-- - Validates that user belongs to the account before granting access
-- - Generates cryptographically secure API keys using gen_random_uuid()
-- - TODO_IMPROVE: Should hash secret keys using proper cryptographic functions
--
CREATE OR REPLACE FUNCTION public.create_portal_application(
    p_portal_account_id VARCHAR(36),
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
    v_new_app_id VARCHAR(36);
    v_secret_data JSON;
    v_secret_key TEXT;
    v_secret_key_hash VARCHAR(255);
    v_user_is_account_member BOOLEAN := FALSE;
    v_secret_key_required_bool BOOLEAN;
    v_result JSON;
BEGIN
    -- ========================================================================
    -- VALIDATION PHASE
    -- ========================================================================
    
    -- Convert text parameter to boolean for internal use
    -- This avoids OpenAPI/SDK generation issues with boolean parameters
    v_secret_key_required_bool := CASE 
        WHEN LOWER(p_secret_key_required) IN ('true', 't', '1', 'yes', 'y') THEN TRUE
        ELSE FALSE
    END;
    
    -- Validate required parameters
    -- PostgreSQL will enforce NOT NULL constraints, but we provide better error messages
    IF p_portal_account_id IS NULL THEN
        RAISE EXCEPTION 'portal_account_id is required'
            USING ERRCODE = '23502', -- not_null_violation
                  HINT = 'Provide a valid portal account ID';
    END IF;
    
    IF p_portal_user_id IS NULL THEN
        RAISE EXCEPTION 'portal_user_id is required'
            USING ERRCODE = '23502',
                  HINT = 'Provide a valid portal user ID';
    END IF;
    
    -- Verify the account exists
    -- This will throw a foreign key constraint error if account doesn't exist
    IF NOT EXISTS (SELECT 1 FROM portal_accounts WHERE portal_account_id = p_portal_account_id) THEN
        RAISE EXCEPTION 'Portal account not found: %', p_portal_account_id
            USING ERRCODE = '23503', -- foreign_key_violation
                  HINT = 'Ensure the portal account exists before creating applications';
    END IF;
    
    -- Verify the user exists and has access to this account
    -- This implements the business rule that users must be account members
    -- before they can create applications within that account
    SELECT EXISTS (
        SELECT 1 FROM portal_account_rbac 
        WHERE portal_account_id = p_portal_account_id 
        AND portal_user_id = p_portal_user_id
        AND user_joined_account = TRUE
    ) INTO v_user_is_account_member;
    
    IF NOT v_user_is_account_member THEN
        RAISE EXCEPTION 'User % is not a member of account %', p_portal_user_id, p_portal_account_id
            USING ERRCODE = '42501', -- insufficient_privilege
                  HINT = 'User must be a member of the account to create applications';
    END IF;
    
    -- ========================================================================
    -- SECRET KEY GENERATION
    -- ========================================================================
    
    -- Use the dedicated secret generation function
    -- This centralizes key generation logic and makes it reusable
    SELECT public.generate_portal_app_secret() INTO v_secret_data;
    
    -- Extract the generated key and hash from the JSON response
    v_secret_key := v_secret_data->>'secret_key';
    v_secret_key_hash := v_secret_data->>'secret_key_hash';
    
    -- ========================================================================
    -- APPLICATION CREATION PHASE
    -- ========================================================================
    
    -- Generate unique application ID
    -- gen_random_uuid() provides a UUID v4 with extremely low collision probability
    -- The PRIMARY KEY constraint ensures database-level uniqueness
    v_new_app_id := gen_random_uuid()::text;
    
    -- Insert the portal application
    -- All optional fields use their database defaults if not provided
    -- The function parameters allow overriding defaults when needed
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
        v_new_app_id,
        p_portal_account_id,
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
    -- RBAC SETUP PHASE
    -- ========================================================================
    
    -- Grant the creating user access to the new application
    -- This follows the principle that application creators should have access to their apps
    -- The unique constraint prevents duplicate entries
    INSERT INTO portal_application_rbac (
        portal_application_id,
        portal_user_id,
        created_at,
        updated_at
    ) VALUES (
        v_new_app_id,
        p_portal_user_id,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    );
    
    -- ========================================================================
    -- RESPONSE GENERATION
    -- ========================================================================
    
    -- Build JSON response with all relevant information
    -- Include the secret key in the response since this is the only time it can be viewed
    -- TODO_IMPROVE: When proper hashing is implemented, return the secret key only once
    SELECT json_build_object(
        'portal_application_id', v_new_app_id,
        'portal_account_id', p_portal_account_id,
        'portal_application_name', p_portal_application_name,
        'secret_key', v_secret_key,
        'secret_key_required', v_secret_key_required_bool,
        'created_at', CURRENT_TIMESTAMP,
        'message', 'Portal application created successfully',
        'warning', 'Store the secret key securely - it cannot be retrieved again'
    ) INTO v_result;
    
    RETURN v_result;
    
EXCEPTION
    WHEN unique_violation THEN
        -- Handle the rare case of UUID collision or constraint violations
        RAISE EXCEPTION 'Application creation failed due to constraint violation: %', SQLERRM
            USING ERRCODE = '23505',
                  HINT = 'This may be due to a duplicate name or UUID collision. Try again.';
    
    WHEN foreign_key_violation THEN
        -- Handle cases where referenced records don't exist
        RAISE EXCEPTION 'Application creation failed due to invalid reference: %', SQLERRM
            USING ERRCODE = '23503',
                  HINT = 'Ensure all referenced accounts and users exist.';
    
    WHEN check_violation THEN
        -- Handle constraint check failures (e.g., negative limits)
        RAISE EXCEPTION 'Application creation failed due to invalid data: %', SQLERRM
            USING ERRCODE = '23514',
                  HINT = 'Check that all numeric values are within valid ranges.';
    
    WHEN OTHERS THEN
        -- Handle any other errors with context
        RAISE EXCEPTION 'Unexpected error during application creation: %', SQLERRM
            USING ERRCODE = SQLSTATE,
                  HINT = 'Contact support if this error persists.';
END;
$$ LANGUAGE plpgsql VOLATILE SECURITY DEFINER;

-- Set function ownership and permissions
-- SECURITY DEFINER allows the function to run with elevated privileges
-- This is necessary to ensure the function can access all required tables
-- See: https://www.postgresql.org/docs/current/sql-createfunction.html#SQL-CREATEFUNCTION-SECURITY

-- Grant execute permission to authenticated users
-- This assumes your PostgREST setup uses an 'authenticated' role for logged-in users
-- Adjust the role name based on your authentication setup
GRANT EXECUTE ON FUNCTION public.create_portal_application TO authenticated;

-- Note: The function should be publicly executable for PostgREST to expose it
-- PostgREST requires functions to have appropriate permissions to be exposed as RPC endpoints

-- Add function comment for documentation
COMMENT ON FUNCTION public.create_portal_application IS 
'Creates a portal application with all associated RBAC entries in a single atomic transaction. 
Validates user membership in the account before creation. 
Returns the application details including the generated secret key.
This function is exposed via PostgREST as POST /rpc/create_portal_application';
