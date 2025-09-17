-- legacy-transform_down.sql

-- Undoes only the changes made by legacy-transform_up.sql
-- This preserves any other data that might exist in the portal database

BEGIN;

-- Disable foreign key checks temporarily to avoid constraint issues
SET session_replication_role = replica;

-- Delete only the data we inserted (in reverse dependency order)
-- portal_applications (depends on portal_accounts)
DELETE FROM portal.portal_applications;

-- portal_account_rbac (depends on portal_accounts and portal_users)
DELETE FROM portal.portal_account_rbac;

-- portal_accounts (depends on portal_plans)
DELETE FROM portal.portal_accounts;

-- portal_users
DELETE FROM portal.portal_users;

-- rbac (only the legacy roles we inserted)
DELETE FROM portal.rbac WHERE role_name IN ('LEGACY_ADMIN', 'LEGACY_OWNER', 'LEGACY_MEMBER');

-- portal_plans (only the specific plans we inserted)
DELETE FROM portal.portal_plans WHERE portal_plan_type IN ('PLAN_FREE', 'PLAN_UNLIMITED', 'PLAN_ENTERPRISE', 'PLAN_INTERNAL');

-- Re-enable foreign key checks
SET session_replication_role = DEFAULT;

-- Reset sequences to start from 1 (only for tables we populated)
ALTER SEQUENCE IF EXISTS portal_users_portal_user_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS rbac_role_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS portal_account_rbac_id_seq RESTART WITH 1;

COMMIT;

