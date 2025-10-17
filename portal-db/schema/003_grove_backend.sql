-- ============================================================================
-- Auxiliary Services Queries
-- ============================================================================
-- This file contains custom views to support auxiliary services
-- like portal-workers, notifications, billing, etc.

-- ============================================================================
-- PORTAL WORKERS ACCOUNT DATA VIEW
-- ============================================================================

-- View to get account data needed by portal-workers for billing operations
--
-- Returns accounts with owner email - generates proper types in SDK
--
-- Use WHERE clauses to filter by portal_plan_type and/or billing_type
CREATE VIEW portal_workers_account_data AS
SELECT
    pa.portal_account_id,
    pa.user_account_name,
    pa.portal_plan_type,
    pa.billing_type,
    pa.portal_account_user_limit,
    pa.gcp_entitlement_id,
    pu.portal_user_email AS owner_email,
    pu.portal_user_id AS owner_user_id
FROM portal_accounts pa
INNER JOIN portal_account_rbac par
    ON pa.portal_account_id = par.portal_account_id
INNER JOIN portal_users pu
    ON par.portal_user_id = pu.portal_user_id
WHERE
    pa.deleted_at IS NULL
    AND pu.deleted_at IS NULL
    AND par.role_name = 'LEGACY_OWNER';

-- Grant select permissions to admin role only (admin has access to portal_users)
GRANT SELECT ON portal_workers_account_data TO portal_db_admin;

COMMENT ON VIEW portal_workers_account_data IS 'Account data for portal-workers billing operations with owner email. Filter using WHERE portal_plan_type = ''PLAN_UNLIMITED'' AND billing_type = ''stripe''';
