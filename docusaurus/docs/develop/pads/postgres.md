---
sidebar_position: 5
title: Postgres Data Source
description: Postgres data source example configuration
---

If the `POSTGRES_CONNECTION_STRING` environment variable is set, **PADS** will connect to the specified Postgres database.

**Postgres triggers are configured to stream updates** to the `Go External Authorization Server` in real time as changes are made to the connected Postgres database.

## Grove Portal DB Driver Background

<div align="center">
<a href="https://www.postgresql.org/">
<img src="https://static-00.iconduck.com/assets.00/postgresql-icon-1987x2048-v2fkmdaw.png" alt="PostgreSQL logo" width="150"/>
<div>https://www.postgresql.org/</div>
</a>
</div>
<br/>

:::info
The database driver implemented is an **extremely opinionated** implementation designed
for backwards compatibility with [Grove's Portal](https://portal.grove.city/).
:::

The [Grove Postgres Driver schema file](https://github.com/buildwithgrove/path-auth-data-server/blob/main/postgres/grove/sqlc/grove_schema.sql)
uses a subset of tables from the existing Grove Portal database schema, allowing `PATH` to source its authorization data from the existing Grove Portal DB.

It converts the data stored in the `portal_applications` table and its associated tables into the `proto.GatewayEndpoint` format expected by PATH's `Go External Authorization Server`.

It also listens for updates to the Grove Portal DB and streams updates to the `Go External Authorization Server` in real time as changes are made to the connected Postgres database.

The full Grove Portal DB schema is defined in the [Portal HTTP DB (PHD) repository](https://github.com/pokt-foundation/portal-http-db/blob/master/postgres-driver/sqlc/schema.sql).

```mermaid
flowchart LR
    subgraph "Grove Portal Database"
        DB[(Full Portal DB)]
        PT[portal_applications table]
        AT[associated tables]
        DB --- PT
        DB --- AT
    end

    subgraph "Grove Postgres Driver"
        GS[Grove Schema\nSubset of Portal DB]
        LC[Change Listener]
        PT --> GS
        AT --> GS
        DB --> LC
    end

    subgraph "PATH System"
        EAS[Go External\nAuthorization Server]
        PE[proto.GatewayEndpoint]
    end

    GS -->|Converts to| PE
    PE --> EAS
    LC -->|Streams real-time updates| EAS
```

### Entity Relationship Diagram

This ERD shows the subset of tables from the full Grove Portal DB schema that are used by the Grove Postgres Driver in PADS.

```mermaid
erDiagram
    PAY_PLANS {
        VARCHAR(25) plan_type PK
        INT monthly_relay_limit
        INT throughput_limit
    }

    ACCOUNTS {
        VARCHAR(10) id PK
        VARCHAR(25) plan_type FK
    }

    USERS {
        VARCHAR(10) id PK
    }

    USER_AUTH_PROVIDERS {
        SERIAL id PK
        VARCHAR(10) user_id FK
        VARCHAR(255) provider_user_id
        VARCHAR type
    }

    ACCOUNT_USERS {
        SERIAL id PK
        VARCHAR(10) user_id FK
        VARCHAR(10) account_id FK
    }

    PORTAL_APPLICATIONS {
        VARCHAR(24) id PK
        VARCHAR(10) account_id FK
    }

    PORTAL_APPLICATION_SETTINGS {
        SERIAL id PK
        VARCHAR(24) application_id FK
        VARCHAR(64) secret_key
        BOOLEAN secret_key_required
    }

    PORTAL_APPLICATION_CHANGES {
        SERIAL id PK
        VARCHAR(24) portal_app_id
        BOOLEAN is_delete
        TIMESTAMP changed_at
    }

    PAY_PLANS ||--o{ ACCOUNTS : "plan_type"
    ACCOUNTS ||--o{ ACCOUNT_USERS : "id"
    USERS ||--o{ ACCOUNT_USERS : "id"
    USERS ||--o{ USER_AUTH_PROVIDERS : "id"
    ACCOUNTS ||--o{ PORTAL_APPLICATIONS : "id"
    PORTAL_APPLICATIONS ||--o{ PORTAL_APPLICATION_SETTINGS : "id"
```

### SQLC Autogeneration

<div align="center">
<a href="https://docs.sqlc.dev/en/stable">
<img src="https://sqlc.dev/logo.png" alt="SQLC logo" width="150"/>
<div>https://docs.sqlc.dev/en/stable</div>
</a>
</div>
<br/>

The Postgres Driver uses `SQLC` to automatically convert SQL definitions into Go code.

The process is started by running `make gen_sqlc`, which reads two main SQL files:

1. [`grove_schema.sql`](https://github.com/buildwithgrove/path-auth-data-server/blob/main/postgres/grove/sqlc/grove_schema.sql): Defines the database structure
2. [`grove_queries.sql`](https://github.com/buildwithgrove/path-auth-data-server/blob/main/postgres/grove/sqlc/grove_queries.sql): Contains the database queries

Using the configuration in [sqlc.yaml](https://github.com/buildwithgrove/path-auth-data-server/blob/main/postgres/grove/sqlc/sqlc.yaml), SQLC generates the corresponding Go code and outputs it to the [postgres/grove/sqlc directory](https://github.com/buildwithgrove/path-auth-data-server/blob/main/postgres/grove/sqlc).

```mermaid
flowchart TD
    subgraph "Input Files"
        S[grove_schema.sql]
        Q[grove_queries.sql]
        C[sqlc.yaml]
    end

    M[make gen_sqlc] -->|triggers| SQLC[SQLC Generator]

    S -->|schema definition| SQLC
    Q -->|query definition| SQLC
    C -->|configuration| SQLC

    subgraph "Output"
        SQLC -->|generates| GO[Go Code]
        GO -->|written to| DIR[postgres/grove/sqlc/]
    end

    style M fill:#e1f5fe
    style SQLC fill:#fff3e0
    style GO fill:#f1f8e9
```

## Additional Postgres Implementations

As mentioned above, this is an **extremely opinionated** implementation designed for backwards compatibility with [Grove's Portal](https://portal.grove.city/).

Pull requests are welcome to support alternative Postgres data sources.

The only requirement is that the gRPC service definition in [`gateway_endpoint.proto`](https://github.com/buildwithgrove/path/blob/main/envoy/auth_server/proto/gateway_endpoint.proto) must be supported.

Alternatively, you may fork the [PADS repository](https://github.com/buildwithgrove/path-auth-data-server) and implement your own data source.
