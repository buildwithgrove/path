-- This file is used by SQLC to autogenerate the Go code needed by the database driver. 
-- It contains all queries used for fetching endpoint data by the Gateway.
-- See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries

-- name: SelectGatewayEndpoints :many
SELECT 
    ge.id,
    ge.account_id,
    ua.plan_type AS plan,
    p.rate_limit_throughput,
    p.rate_limit_capacity,
    p.rate_limit_capacity_period,
    ARRAY_AGG(uap.provider_user_id)::VARCHAR[] AS provider_user_ids
FROM gateway_endpoints ge
LEFT JOIN user_accounts ua 
    ON ge.account_id = ua.id
LEFT JOIN plans p 
    ON ua.plan_type = p.type
LEFT JOIN account_users au
    ON ge.account_id = au.account_id
LEFT JOIN users u
    ON au.user_id = u.id
LEFT JOIN user_auth_providers uap
    ON u.id = uap.user_id
GROUP BY 
    ge.id,
    ua.plan_type,
    p.rate_limit_throughput,
    p.rate_limit_capacity,
    p.rate_limit_capacity_period;
    