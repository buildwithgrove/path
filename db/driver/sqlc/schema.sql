CREATE TABLE plans (
    id SERIAL PRIMARY KEY,
    type VARCHAR(255) NOT NULL UNIQUE,
    throughput_limit INT NOT NULL
);
CREATE TABLE accounts (
    id VARCHAR(24) PRIMARY KEY,
    plan_type VARCHAR(255) NOT NULL REFERENCES plans(type)
);
CREATE TABLE user_apps (
    id VARCHAR(24) PRIMARY KEY,
    account_id VARCHAR(24) REFERENCES accounts(id) ON DELETE CASCADE,
    secret_key VARCHAR(255),
    secret_key_required BOOLEAN DEFAULT FALSE
);
CREATE TYPE allowlist_type AS ENUM (
    'contracts',
    'methods',
    'origins',
    'services',
    'user_agents'
);
CREATE TABLE user_app_allowlists (
    id SERIAL PRIMARY KEY,
    user_app_id VARCHAR(24) NOT NULL REFERENCES user_apps(id) ON DELETE CASCADE,
    type allowlist_type NOT NULL,
    value VARCHAR(255) NOT NULL,
    UNIQUE (user_app_id, type, value)
);