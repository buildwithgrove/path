-- This file is used by SQLC to autogenerate the Go code needed by the database driver. 
-- It contains all queries used for fetching user data by the Gateway.
-- See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries

-- name: SelectGatewayEndpoints :many
SELECT 
    ge.id,
    ge.account_id,
    ge.api_key,
    ge.api_key_required,
    ua.plan_type AS plan,
    p.rate_limit_throughput,
    p.rate_limit_capacity,
    p.rate_limit_capacity_period
FROM gateway_endpoints ge
LEFT JOIN user_accounts ua 
    ON ge.account_id = ua.id
LEFT JOIN plans p 
    ON ua.plan_type = p.type
GROUP BY 
    ge.id,
    ua.plan_type,
    p.rate_limit_throughput,
    p.rate_limit_capacity,
    p.rate_limit_capacity_period;
