-- This file updates the ephemeral Docker Postres test database initialized in db/driver/driver_test.go
-- with just enough data to run the test of the database driver using an actual Postgres DB instance.
INSERT INTO plans (type, rate_limit_throughput)
VALUES ('PLAN_FREE', 30),
    ('PLAN_UNLIMITED', 0);
INSERT INTO accounts (id, plan_type)
VALUES ('account_1', 'PLAN_FREE'),
    ('account_2', 'PLAN_UNLIMITED'),
    ('account_3', 'PLAN_FREE');
INSERT INTO user_apps (id, account_id, secret_key, secret_key_required)
VALUES ('user_app_1', 'account_1', 'secret_key_1', TRUE),
    ('user_app_2', 'account_2', 'secret_key_2', TRUE),
    ('user_app_3', 'account_3', 'secret_key_3', TRUE),
    ('user_app_4', 'account_1', NULL, FALSE),
    ('user_app_5', 'account_2', NULL, FALSE);
INSERT INTO user_app_allowlists (user_app_id, type, value)
VALUES ('user_app_2', 'contracts', 'contract_1'),
    ('user_app_3', 'methods', 'method_1'),
    ('user_app_4', 'origins', 'origin_1'),
    ('user_app_1', 'services', 'service_1'),
    ('user_app_5', 'user_agents', 'user_agent_1');