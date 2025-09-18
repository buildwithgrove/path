-- legacy-transform_down.sql
-- Rollback script to truncate all data inserted by legacy-transform_up.sql
-- This allows for clean re-testing of the migration

BEGIN;

-- Truncate tables in reverse dependency order to avoid foreign key constraint issues

-- 1. Remove portal application allowlists (depends on portal_applications)
TRUNCATE TABLE public.portal_application_allowlists RESTART IDENTITY CASCADE;

-- 2. Remove portal applications (depends on portal_accounts)
TRUNCATE TABLE public.portal_applications RESTART IDENTITY CASCADE;

-- 3. Remove portal account RBAC (depends on portal_accounts and portal_users)
TRUNCATE TABLE public.portal_account_rbac RESTART IDENTITY CASCADE;

-- 4. Remove RBAC roles
TRUNCATE TABLE public.rbac RESTART IDENTITY CASCADE;

-- 5. Remove portal users
TRUNCATE TABLE public.portal_users RESTART IDENTITY CASCADE;

-- 6. Remove portal accounts (depends on portal_plans)
TRUNCATE TABLE public.portal_accounts RESTART IDENTITY CASCADE;

-- 7. Remove portal plans (base table)
TRUNCATE TABLE public.portal_plans RESTART IDENTITY CASCADE;

COMMIT;
