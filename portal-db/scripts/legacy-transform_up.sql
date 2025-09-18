-- legacy-transform_up.sql


-- Transform from Legacy Portal DB Data to New Portal DB
-- Order of operations mimics the init script in PATH repo
-- See: https://github.com/buildwithgrove/path
BEGIN;

-- We omit the organizations tables as these are net new and not a hard requirement for any subsequent tables

-- First populate the Portal Plans
INSERT INTO "portal".portal_plans (
    portal_plan_type,
    portal_plan_type_description,
    plan_usage_limit,
    plan_usage_limit_interval,
    plan_rate_limit_rps,
    plan_application_limit
) 
VALUES
    ('PLAN_FREE', 'Free tier with limited usage', 1000000, 'month', 0, 2),
    ('PLAN_UNLIMITED', 'Unlimited Relays. Unlimited RPS.', 0, NULL, 0, 0),
    -- Note that we rename ENTERPRISE -> PLAN_ENTERPRISE for consistency
    ('PLAN_ENTERPRISE', 'Special case for Legacy Enterprise customers', 0, NULL, 0, 0),
    -- Introduce the new plan type so we can separate out Grove-owned services
    ('PLAN_INTERNAL', 'Plan for internal accounts', 0, NULL, 0, 0);

-- Transform the legacy accounts to the new accounts structure
-- If the plan doesn't match one of our new plan types, then move them to `PLAN_FREE`
INSERT INTO portal.portal_accounts (
    portal_account_id,
    organization_id,
    portal_plan_type,
    user_account_name,
    internal_account_name,
    portal_account_user_limit,
    portal_account_user_limit_interval,
    portal_account_user_limit_rps,
    billing_type,
    stripe_subscription_id,
    gcp_account_id,
    gcp_entitlement_id,
    deleted_at,
    created_at,
    updated_at
)
SELECT 
    -- casts each legacy ID as a uuid
    a.id::uuid as portal_account_id,
    NULL as organization_id,
    CASE 
        WHEN a.plan_type = 'ENTERPRISE' THEN 'PLAN_ENTERPRISE'
        WHEN a.plan_type = 'PLAN_UNLIMITED' THEN 'PLAN_UNLIMITED'
        ELSE 'PLAN_FREE'
    END as portal_plan_type,
    a.name as user_account_name,
    a.name as internal_account_name,
    CASE 
        WHEN a.monthly_user_limit = 0 THEN NULL
        ELSE a.monthly_user_limit 
    END as portal_account_user_limit,
    CASE 
        WHEN a.monthly_user_limit > 0 THEN 'month'::plan_interval
        ELSE NULL
    END as portal_account_user_limit_interval,
    NULL as portal_account_user_limit_rps,
    a.billing_type,
    ai.stripe_subscription_id,
    a.gcp_account_id,
    a.gcp_entitlement_id,
    CASE 
        WHEN a.deleted = true THEN a.deleted_at
        ELSE NULL
    END as deleted_at,
    a.created_at,
    a.updated_at
FROM "legacy-extract".accounts a
LEFT JOIN "legacy-extract".account_integrations ai ON a.id = ai.account_id;

-- Transform the legacy users to the new users table and make sure to grab all relevant data
INSERT INTO portal.portal_users (
    portal_user_id,
    portal_user_email,
    signed_up,
    portal_admin,
    deleted_at,
    created_at,
    updated_at
)
SELECT 
    u.id as portal_user_id,
    u.email as portal_user_email,
    COALESCE(u.signed_up, FALSE) as signed_up,
    FALSE as portal_admin,
    NULL as deleted_at,
    u.created_at,
    u.updated_at
FROM "legacy-extract".users u;

-- Update the sequence after inserting users with specific IDs
SELECT setval('portal_users_portal_user_id_seq', (SELECT MAX(portal_user_id) FROM portal_users));

-- Load the RBAC Table with the basic roles we have today, to be adjusted later
-- Note that we tag them with the `LEGACY_` prefix so we can parse these out later
INSERT INTO portal.rbac (
    role_id,
    role_name,
    permissions
)
VALUES 
    (DEFAULT, 'LEGACY_ADMIN', ARRAY[]),
    (DEFAULT, 'LEGACY_OWNER', ARRAY[]),
    (DEFAULT, 'LEGACY_MEMBER', ARRAY[]);

-- Transform the legacy account_users table into the new portal_account_rbac
-- Note that we transform the role names to include a `LEGACY_` prefix so these are easier
-- to parse.
INSERT INTO portal.portal_account_rbac (
    portal_account_id,
    portal_user_id,
    role_name,
    user_joined_account
)
SELECT 
    au.account_id::uuid as portal_account_id,
    au.user_id as portal_user_id,
    CASE 
        WHEN au.role_name = 'ADMIN' THEN 'LEGACY_ADMIN'
        WHEN au.role_name = 'OWNER' THEN 'LEGACY_OWNER'
        WHEN au.role_name = 'MEMBER' THEN 'LEGACY_MEMBER'
        ELSE au.role_name
    END as role_name,
    COALESCE(au.accepted, FALSE) as user_joined_account
FROM "legacy-extract".account_users au;

INSERT INTO portal.portal_applications (
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
    deleted_at,
    created_at,
    updated_at
)
SELECT 
    pa.id::uuid as portal_application_id,
    pa.account_id::uuid as portal_account_id,
    pa.name as portal_application_name,
    pa.app_emoji as emoji,
    NULL as portal_application_user_limit,
    NULL as portal_application_user_limit_interval,
    NULL as portal_application_user_limit_rps,
    pa.description as portal_application_description,
    (
        SELECT ARRAY_AGG(c.blockchain) 
        FROM "legacy-extract".chains c 
        WHERE c.id = ANY(pas.favorited_chain_ids)
    ) as favorite_service_ids,
    pas.secret_key as secret_key_hash,
    COALESCE(pas.secret_key_required, FALSE) as secret_key_required,
    CASE 
        WHEN pa.deleted = true THEN pa.deleted_at
        ELSE NULL
    END as deleted_at,
    pa.created_at,
    COALESCE(pas.updated_at, pa.updated_at) as updated_at
FROM "legacy-extract".portal_applications pa
LEFT JOIN "legacy-extract".portal_application_settings pas ON pa.id = pas.application_id;

-- Transform the legacy portal_application_whitelists to the new portal_application_allowlists
INSERT INTO portal.portal_application_allowlists (
    portal_application_id,
    type,
    value,
    service_id,
    created_at,
    updated_at
)
SELECT 
    paw.application_id::uuid as portal_application_id,
    CASE 
        WHEN paw.type = 'blockchains' THEN 'service_id'::allowlist_type
        WHEN paw.type = 'origins' THEN 'origin'::allowlist_type
        WHEN paw.type = 'contracts' THEN 'contract'::allowlist_type
        ELSE paw.type::allowlist_type
    END as type,
    paw.value,
    CASE 
        WHEN paw.type = 'blockchains' THEN c.blockchain
        ELSE NULL
    END as service_id,
    paw.created_at,
    paw.created_at as updated_at  -- Using created_at since there's no updated_at in legacy
FROM "legacy-extract".portal_application_whitelists paw
LEFT JOIN "legacy-extract".chains c ON paw.chain_id = c.id;

COMMIT;
