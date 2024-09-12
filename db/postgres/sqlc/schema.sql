-- This file is used by SQLC to autogenerate the Go code needed by the database driver. 
-- It contains all tables required for storing user data needed by the Gateway.
-- See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries
--
CREATE TYPE rate_limit_capacity_period AS ENUM ('daily', 'weekly', 'monthly');
CREATE TABLE plans (
    id SERIAL PRIMARY KEY,
    type VARCHAR(255) NOT NULL UNIQUE,
    rate_limit_throughput INT,
    rate_limit_capacity INT,
    rate_limit_capacity_period rate_limit_capacity_period,
    CHECK (
        (
            rate_limit_capacity IS NOT NULL
            AND rate_limit_capacity_period IS NOT NULL
        )
        OR (
            rate_limit_capacity IS NULL
            AND rate_limit_capacity_period IS NULL
        )
    )
);
CREATE TABLE user_accounts (
    id VARCHAR(24) PRIMARY KEY,
    plan_type VARCHAR(255) NOT NULL REFERENCES plans(type)
);
CREATE TABLE gateway_endpoints (
    id VARCHAR(24) PRIMARY KEY,
    account_id VARCHAR(24) REFERENCES user_accounts(id) ON DELETE CASCADE,
    api_key VARCHAR(255),
    api_key_required BOOLEAN DEFAULT FALSE
);