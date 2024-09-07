-- name: SelectUserApps :many
SELECT u.id,
    u.account_id,
    u.secret_key,
    u.secret_key_required,
    COALESCE(
        jsonb_object_agg(
            w.type,
            jsonb_build_object(w.value, '{}'::jsonb)
        ) FILTER (
            WHERE w.user_app_id IS NOT NULL
        ),
        '{}'::jsonb
    )::jsonb AS whitelists,
    a.plan_type AS plan,
    p.throughput_limit
FROM user_apps u
    LEFT JOIN user_app_whitelists w ON u.id = w.user_app_id
    LEFT JOIN accounts a ON u.account_id = a.id
    LEFT JOIN plans p ON a.plan_type = p.type
GROUP BY u.id,
    a.plan_type,
    p.throughput_limit;