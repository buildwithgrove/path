-- legacy-transform_up.sql (updated without UUID casts and fixed enum casting)

BEGIN;

-- First populate the Portal Plans
INSERT INTO public.portal_plans (
    portal_plan_type,
    portal_plan_type_description,
    plan_usage_limit,
    plan_usage_limit_interval,
    plan_rate_limit_rps,
    plan_application_limit
)
VALUES
    ('PLAN_FREE', 'Free tier with limited monthly usage and rate limited.', 1e6, 'month', 100, 2),
    ('PLAN_PAID', 'Paid tier with unlimited monthly usage (pay-as-you-go) and no rate limits.', 0, NULL, 0, 0),
    ('PLAN_LEGACY', 'Legacy plan for enterprise customers', 0, NULL, 0, 0),
    ('PLAN_INTERNAL', 'Internal plan for development purposes only', 0, NULL, 0, 0);

-- Transform the legacy accounts to the new accounts structure
INSERT INTO public.portal_accounts (
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
    a.id as portal_account_id,
    NULL as organization_id,
    CASE
        WHEN a.plan_type = 'ENTERPRISE' THEN 'PLAN_LEGACY'
        WHEN a.plan_type = 'PLAN_UNLIMITED' THEN 'PLAN_PAID'
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
FROM legacy_extract.accounts a
LEFT JOIN legacy_extract.account_integrations ai ON a.id = ai.account_id;

-- Transform the legacy users to the new users table
INSERT INTO public.portal_users (
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
FROM legacy_extract.users u;

-- Load the RBAC Table
INSERT INTO public.rbac (
    role_id,
    role_name,
    permissions
)
VALUES
    (DEFAULT, 'LEGACY_ADMIN', ARRAY[]::VARCHAR[]),
    (DEFAULT, 'LEGACY_OWNER', ARRAY[]::VARCHAR[]),
    (DEFAULT, 'LEGACY_MEMBER', ARRAY[]::VARCHAR[]);

-- Transform the legacy account_users table into the new portal_account_rbac
-- Join on email since we're using new auto-generated portal_user_id values
INSERT INTO public.portal_account_rbac (
    portal_account_id,
    portal_user_id,
    role_name,
    user_joined_account
)
SELECT
    au.account_id as portal_account_id,
    pu.portal_user_id,  -- Use the new auto-generated ID
    CASE
        WHEN au.role_name = 'ADMIN' THEN 'LEGACY_ADMIN'
        WHEN au.role_name = 'OWNER' THEN 'LEGACY_OWNER'
        WHEN au.role_name = 'MEMBER' THEN 'LEGACY_MEMBER'
        ELSE au.role_name
    END as role_name,
    COALESCE(au.accepted, FALSE) as user_joined_account
FROM legacy_extract.account_users au
JOIN legacy_extract.users lu ON au.user_id = lu.id
JOIN public.portal_users pu ON lu.email = pu.portal_user_email;

INSERT INTO public.portal_applications (
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
    pa.id as portal_application_id,
    pa.account_id as portal_account_id,
    LEFT(pa.name, 42) as portal_application_name,  -- Truncate to 42 chars
    pa.app_emoji as emoji,
    NULL as portal_application_user_limit,
    NULL as portal_application_user_limit_interval,
    NULL as portal_application_user_limit_rps,
    LEFT(pa.description, 255) as portal_application_description,
    (
        SELECT ARRAY_AGG(c.blockchain)
        FROM legacy_extract.chains c
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
FROM legacy_extract.portal_applications pa
LEFT JOIN legacy_extract.portal_application_settings pas ON pa.id = pas.application_id;

-- Transform the legacy portal_application_whitelists to the new portal_application_allowlists
INSERT INTO public.portal_application_allowlists (
    portal_application_id,
    type,
    value,
    service_id,
    created_at,
    updated_at
)
SELECT
    paw.application_id as portal_application_id,
    CASE
        WHEN paw.type = 'blockchains' THEN 'service_id'::allowlist_type
        WHEN paw.type = 'origins' THEN 'origin'::allowlist_type
        WHEN paw.type = 'contracts' THEN 'contract'::allowlist_type
    END as type,
    paw.value,
    CASE
        WHEN paw.type = 'blockchains' THEN c.blockchain
        ELSE NULL
    END as service_id,
    paw.created_at,
    paw.created_at as updated_at
FROM legacy_extract.portal_application_whitelists paw
LEFT JOIN legacy_extract.chains c ON paw.chain_id = c.id
WHERE paw.type IN ('blockchains', 'origins', 'contracts');  -- Only process known types

-- Transform legacy_extract.user_auth_providers to public.portal_user_auth
INSERT INTO public.portal_user_auth (
    portal_user_id,
    portal_auth_provider,
    portal_auth_type,
    auth_provider_user_id,
    federated,
    created_at,
    updated_at
)
SELECT
    uap.user_id as portal_user_id,
    'auth0'::portal_auth_provider as portal_auth_provider,  -- auth0 is the only provider supported during migration
    CASE
        WHEN uap.type = 'auth0_username' THEN 'auth0_username'::portal_auth_type
        WHEN uap.type = 'auth0_github' THEN 'auth0_github'::portal_auth_type
        WHEN uap.type = 'auth0_google' THEN 'auth0_google'::portal_auth_type
    END as portal_auth_type,
    uap.provider_user_id as auth_provider_user_id,
    COALESCE(uap.federated, FALSE) as federated,
    uap.created_at,
    uap.created_at as updated_at  -- Use created_at since legacy table doesn't have updated_at
FROM legacy_extract.user_auth_providers uap
WHERE uap.provider = 'auth0'  -- Only auth0 provider supported
  AND uap.type IN ('auth0_username', 'auth0_github', 'auth0_google');

COMMIT;
