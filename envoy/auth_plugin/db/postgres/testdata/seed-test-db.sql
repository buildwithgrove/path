-- This file updates the ephemeral Docker Postgres test database initialized in db/driver/driver_test.go
-- with just enough data to run the test of the database driver using an actual Postgres DB instance.
-- Insert into the 'plans' table
INSERT INTO plans (
        type,
        rate_limit_throughput,
        rate_limit_capacity,
        rate_limit_capacity_period
    )
VALUES ('PLAN_FREE', 30, 100000, 'daily'),
    ('PLAN_UNLIMITED', 0, NULL, NULL);

-- Insert into the 'user_accounts' table
INSERT INTO user_accounts (id, plan_type)
VALUES ('account_1', 'PLAN_FREE'),
    ('account_2', 'PLAN_UNLIMITED'),
    ('account_3', 'PLAN_FREE');

-- Insert into the 'gateway_endpoints' table
INSERT INTO gateway_endpoints (id, account_id, api_key, api_key_required)
VALUES ('endpoint_1', 'account_1', 'api_key_1', TRUE),
    ('endpoint_2', 'account_2', 'api_key_2', TRUE),
    ('endpoint_3', 'account_3', 'api_key_3', TRUE),
    ('endpoint_4', 'account_1', NULL, FALSE),
    ('endpoint_5', 'account_2', NULL, FALSE);
    