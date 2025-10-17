-- ============================================================================
-- GROVE PORTAL VIEWS
-- ============================================================================
-- This file contains views that aggregate and structure portal data for
-- frontend consumption, matching GraphQL query structures.
-- ============================================================================

-- ============================================================================
-- VIEW: user_accounts_view
-- ============================================================================
-- Aggregates portal account data with related users, applications, and plans
-- for authenticated users. Matches the structure of getUserAccount/getUserAccounts
-- GraphQL queries.
--
-- Access Control: Only returns accounts where the authenticated user
-- (via JWT) has RBAC access through portal_account_rbac.
-- ============================================================================

CREATE OR REPLACE VIEW api.user_accounts_view AS
SELECT
    pa.portal_account_id,
    pa.user_account_name,
    pa.portal_plan_type,
    pa.portal_account_user_limit,
    pa.portal_account_user_limit_interval,
    pa.portal_account_user_limit_rps,
    pa.stripe_subscription_id,
    pa.billing_type,
    pa.gcp_account_id,
    pa.gcp_entitlement_id,
    pa.created_at,
    pa.updated_at,
    -- Aggregated notification thresholds as JSONB array
    COALESCE(
        jsonb_agg(DISTINCT threshold) FILTER (WHERE threshold IS NOT NULL),
        '[]'::jsonb
    ) AS portal_account_user_limit_notification_thresholds,
    -- Users array: aggregate all users with RBAC access to this account
    COALESCE(
        (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'portal_user_id', pu.portal_user_id,
                    'portal_user_email', pu.portal_user_email,
                    'role_name', par_inner.role_name,
                    'user_joined_account', par_inner.user_joined_account
                )
                ORDER BY pu.portal_user_email
            )
            FROM portal_account_rbac par_inner
            JOIN portal_users pu ON par_inner.portal_user_id = pu.portal_user_id
            WHERE par_inner.portal_account_id = pa.portal_account_id
              AND pu.deleted_at IS NULL
        ),
        '[]'::jsonb
    ) AS users,
    -- Portal Apps array: aggregate all applications for this account
    COALESCE(
        (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'portal_application_id', papp.portal_application_id,
                    'portal_application_name', papp.portal_application_name,
                    'emoji', papp.emoji
                )
                ORDER BY papp.portal_application_name
            )
            FROM portal_applications papp
            WHERE papp.portal_account_id = pa.portal_account_id
              AND papp.deleted_at IS NULL
        ),
        '[]'::jsonb
    ) AS portal_apps,
    -- Plan object: full portal_plans row as JSONB
    COALESCE(
        (
            SELECT jsonb_build_object(
                'portal_plan_type', pp.portal_plan_type,
                'portal_plan_type_description', pp.portal_plan_type_description,
                'plan_usage_limit', pp.plan_usage_limit,
                'plan_usage_limit_interval', pp.plan_usage_limit_interval,
                'plan_rate_limit_rps', pp.plan_rate_limit_rps,
                'plan_application_limit', pp.plan_application_limit
            )
            FROM portal_plans pp
            WHERE pp.portal_plan_type = pa.portal_plan_type
        ),
        '{}'::jsonb
    ) AS plan
FROM portal_accounts pa
LEFT JOIN LATERAL unnest(pa.portal_account_user_limit_notification_thresholds) AS threshold ON true
WHERE pa.deleted_at IS NULL
  -- RLS: Only include accounts the authenticated user has access to
  AND EXISTS (
      SELECT 1
      FROM portal_account_rbac par
      WHERE par.portal_account_id = pa.portal_account_id
        AND par.portal_user_id = api.current_portal_user_id()
  )
GROUP BY 
    pa.portal_account_id,
    pa.user_account_name,
    pa.portal_plan_type,
    pa.portal_account_user_limit,
    pa.portal_account_user_limit_interval,
    pa.portal_account_user_limit_rps,
    pa.stripe_subscription_id,
    pa.billing_type,
    pa.gcp_account_id,
    pa.gcp_entitlement_id,
    pa.created_at,
    pa.updated_at;

COMMENT ON VIEW api.user_accounts_view IS 'Aggregated view of portal accounts with users, applications, and plans. Filtered by authenticated user RBAC access.';

-- ============================================================================
-- PERMISSIONS
-- ============================================================================

-- Grant SELECT access to authenticated_user role only
GRANT SELECT ON api.user_accounts_view TO authenticated_user;

-- Revoke access from other roles to ensure only authenticated users can query
REVOKE ALL ON api.user_accounts_view FROM PUBLIC;
REVOKE ALL ON api.user_accounts_view FROM anon;

