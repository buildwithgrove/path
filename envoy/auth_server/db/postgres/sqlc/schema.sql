-- This file is used by SQLC to autogenerate the Go code needed by the database driver. 
-- It contains all tables required for storing endpoint data needed by the Gateway.
-- See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries

-- Create ENUM type for rate_limit_capacity_period
CREATE TYPE rate_limit_capacity_period AS ENUM ('daily', 'weekly', 'monthly');

-- Create 'plans' table
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

-- Create 'user_accounts' table
CREATE TABLE user_accounts (
    id VARCHAR(24) PRIMARY KEY,
    plan_type VARCHAR(255) NOT NULL REFERENCES plans(type)
);

-- Create 'users' table
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY
);

-- Create 'account_users' table
CREATE TABLE account_users (
    user_id VARCHAR(10) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id VARCHAR(10) NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, account_id)
);

-- Create 'gateway_endpoints' table
CREATE TABLE gateway_endpoints (
    id VARCHAR(24) PRIMARY KEY,
    account_id VARCHAR(24) REFERENCES user_accounts(id) ON DELETE CASCADE,
    api_key VARCHAR(255),
    api_key_required BOOLEAN DEFAULT FALSE
);
