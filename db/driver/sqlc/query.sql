-- This file is used by SQLC to autogenerate the Go code needed by the database driver. 
-- It contains all queries used for fetching user data by the Gateway.
-- See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries
-- 
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
    )::jsonb AS allowlists,
    a.plan_type AS plan,
    p.rate_limit_throughput
FROM user_apps u
    LEFT JOIN user_app_allowlists w ON u.id = w.user_app_id
    LEFT JOIN accounts a ON u.account_id = a.id
    LEFT JOIN plans p ON a.plan_type = p.type
GROUP BY u.id,
    a.plan_type,
    p.rate_limit_throughput;