# Portal DB <!-- omit in toc -->

The Portal DB is the house for all core business logic for both PATH and the Portal.

The Portal DB is a _highly opinionated_ implementation of a Postgres database that can be used to manage and administer both PATH and a UI on top of PATH.

## Table of Contents <!-- omit in toc -->

- [Quickstart](#quickstart)
  - [Quickstart - Fast Route](#quickstart---fast-route)
  - [Quickstart - Manual Route](#quickstart---manual-route)
- [Production Readiness](#production-readiness)
- [REST API Access](#rest-api-access)
- [Tools](#tools)
  - [`psql` (REQUIRED)](#psql-required)
  - [`dbeaver` (RECOMMENDED)](#dbeaver-recommended)
  - [Claude Postgres MCP Server (EXPERIMENTAL)](#claude-postgres-mcp-server-experimental)

## Quickstart

### Quickstart - Fast Route

```bash
cd portal-db
make quickstart
```

### Quickstart - Manual Route

You can manually populate and check the contents of the database following the instructions below.

<details>
<summary>Manual Database Population</summary>

We'll connect to the following gateway and applications:

- gateway - `pokt1lf0kekv9zcv9v3wy4v6jx2wh7v4665s8e0sl9s`
- solana app - `pokt1xd8jrccxtlzs8svrmg6gukn7umln7c2ww327xx`
- eth app - `pokt185tgfw9lxyuznh9rz89556l4p8dshdkjd5283d`
- xrplevm app - `pokt1gwxwgvlxlzk3ex59cx7lsswyvplf0rfhunxjhy`
- poly app - `pokt1hufj6cdgu83dluput6klhmh54vtrgtl3drttva`

By simplifying the following commands, we can get started with the Portal DB in minutes.

```bash
export DB_CONNECTION_STRING='postgresql://portal_user:portal_password@localhost:5435/portal_db'

cd portal-db
make portal-db-up

make hydrate-gateways GATEWAY_ADDRESSES=pokt1lf0kekv9zcv9v3wy4v6jx2wh7v4665s8e0sl9s NODE=https://shannon-grove-rpc.mainnet.poktroll.com NETWORK=pocket
make hydrate-services SERVICE_IDS='eth,poly,solana,xrplevm' NODE=https://shannon-grove-rpc.mainnet.poktroll.com NETWORK=pocket
make hydrate-applications APPLICATION_ADDRESSES='pokt1xd8jrccxtlzs8svrmg6gukn7umln7c2ww327xx,pokt185tgfw9lxyuznh9rz89556l4p8dshdkjd5283d,pokt1gwxwgvlxlzk3ex59cx7lsswyvplf0rfhunxjhy,pokt1hufj6cdgu83dluput6klhmh54vtrgtl3drttva' NODE=https://shannon-grove-rpc.mainnet.poktroll.com NETWORK=pocket

psql $DB_CONNECTION_STRING
SELECT * FROM gateways;
SELECT * FROM services;
SELECT * FROM applications;
```

</details>

## Production Readiness

Search for `TODO_PRODUCTION` in the codebase to see items that need to be addressed
before promoting the Portal DB/PostgREST stack to production.

For example, as of #467, the output of this command:

```bash
grep -r "TODO_PRODUCTION" ./portal-db
```

Shows:

1. **Externalize the JWT secret**: Replace the hardcoded development secret in `api/postgrest.conf` with a secure value managed via env vars or secrets storage.
2. **Synchronize JWT secret usage**: Update `api/scripts/postgrest-gen-jwt.sh` to source the same secret (or invoke PostgREST’s `/jwt` RPC when available) instead of embedding a development constant.

## REST API Access

The Portal DB includes a **PostgREST API** that automatically generates REST endpoints from your database schema. This provides instant HTTP access to all your data with authentication, filtering, and Go SDK generation.

**➡️ [View PostgREST API Documentation](api/README.md)** for setup, authentication, and SDK usage.

## Tools

This is a list of tools for interacting with the database.

### `psql` (REQUIRED)

**Installation**:

- 🍎 Mac: `brew install postgresql`
- 🅰️ Arch: `pacman -S postgresql`
- 🌀 Debian: `sudo apt-get install postgresql`

**Usage**:

```bash
export DB_CONNECTION_STRING='postgresql://portal_user:portal_password@localhost:5435/portal_db'
psql $DB_CONNECTION_STRING
```

### `dbeaver` (RECOMMENDED)

It is _highly recommended_ to use a GUI Database Explorer in conjunction with the Portal DB.

This allows a user to directly Create, Read, Update, and Delete (CRUD) database records in a GUI. We recommend `dbeaver`.

:::tip ERD - Entity Relationship Diagrams

One reason we recommend `dbeaver` is its native functionality of creating an ERD - a visual tool for seeing how tables are interrelated in SQL.

Once you have the Database running and `dbeaver` installed and configured, you can right-click on a schema and choose: "View Diagram" for an interactive ERD.

:::

**Install `dbeaver`**

- 🍎 Mac: `brew install --cask dbeaver-community`
- 🅰️ Arch: `paru -S dbeaver`
- 🌀 Debian:

```bash
sudo add-apt-repository ppa:serge-rider/dbeaver-ce
sudo apt-get update
sudo apt-get install dbeaver-ce
```

**`dbeaver` Connection String Setup**

- Open `dbeaver`
- File > New Connection
- Select Postgres for the DB Driver
- Enter the connection details:
  - `URL`: `jdbc:postgresql://127.0.0.1:5435/portal_db?sslmode=disable`
  - `Username`: `portal_user`
  - `Password`: `portal_password`
- Connect to the database and explore

![dbeaver connection](../docusaurus/static/img/portal_db_connection.png)

### Claude Postgres MCP Server (EXPERIMENTAL)

<details>
<summary>Experimental Postgres MCP Server</summary>

1. Install [postgres-mcp](https://github.com/crystaldba/postgres-mcp) using `pipx`.

   ```bash
   pipx install postgres-mcp
   ```

2. Update your [claude_desktop_config.json](claude_desktop_config.json) with the setting below. On macOS, you'll find it at `~/Library/Application Support/Claude/claude_desktop_config.json`.

   ```json
   {
     "mcpServers": {
       "postgres": {
         "command": "/Users/olshansky/.local/bin/postgres-mcp",
         "args": ["--access-mode=restricted"],
         "env": {
           "DATABASE_URI": "postgresql://portal_user:portal_password@localhost:5435/portal_db"
         }
       }
     }
   }
   ```

3. Restart Claude Desktop

4. Create a Claude Project with the following system prompt:

   ```text
   You are a professional software engineer and database administrator specializing in SQL query design and PostgreSQL database navigation.

   Your role is to:
   - Analyze the provided database schema.
   - Leverage the MCP server to validate schema details and explore available tables, columns, and relationships.
   - Generate accurate, efficient, and secure SQL queries that align with the user’s request.
   - Clearly explain your reasoning and the structure of the queries when helpful, but keep results concise and actionable.
   - Assume all queries target a PostgreSQL database unless explicitly stated otherwise.

   You must:
   - Use the schema as the source of truth for query construction.
   - Ask clarifying questions if user requests are ambiguous or under-specified.
   - Favor correctness, readability, and performance best practices in all SQL you produce.
   ```

5. Upload [schema/001_portal_init.sql](schema/001_portal_init.sql) as one of the files to the Claude Project.

6. Try using it by asking: `How many records are in my database?`

![claude_desktop_postgres_mcp](../docusaurus/static/img/claude_desktop_postgres_mcp.png)

</details>
