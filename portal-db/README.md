# Portal DB <!-- omit in toc -->

The Portal DB is the house for all core business logic for both PATH and the Portal.

The Portal DB is a _highly opinionated_ implementation of a Postgres database that can be used to manage and administer both PATH and a UI on top of PATH.

## Table of Contents <!-- omit in toc -->

- [🌐 REST API Access](#-rest-api-access)
- [💻 REST API Client SDKs](#-rest-api-client-sdks)
- [Quickstart (for Grove Engineering)](#quickstart-for-grove-engineering)
- [Interacting with the database](#interacting-with-the-database)
  - [`make` Targets](#make-targets)
  - [`scripts`](#scripts)
- [Tools](#tools)
  - [`psql` (REQUIRED)](#psql-required)
  - [`dbeaver` (RECOMMENDED)](#dbeaver-recommended)
  - [Claude Postgres MCP Server (EXPERIMENTAL)](#claude-postgres-mcp-server-experimental)

## 🌐 REST API Access

The Portal DB includes a **PostgREST API** that automatically generates REST endpoints from your database schema. This provides instant HTTP access to all your data with authentication, filtering, and Go SDK generation.

**➡️ [View PostgREST API Documentation](api/README.md)** for setup, authentication, and SDK usage.

## 💻 REST API Client SDKs

The Portal DB includes client SDKs for both Go and TypeScript.

**➡️ [View Go SDK Documentation](sdk/go/README.md)**
**➡️ [View TypeScript SDK Documentation](sdk/typescript/README.md)**

:::warning TODO(@olshansk): Revisit docs location

Consider if this should be moved into `docusaurus/docs` so it is discoverable as part of [path.grove.city](https://path.grove.city/).

:::

## Quickstart (for Grove Engineering)

We'll connect to the following gateway and applications:

- gateway - `pokt1lf0kekv9zcv9v3wy4v6jx2wh7v4665s8e0sl9s`
- solana app - `pokt1xd8jrccxtlzs8svrmg6gukn7umln7c2ww327xx`
- eth app - `pokt185tgfw9lxyuznh9rz89556l4p8dshdkjd5283d`
- xrplevm app - `pokt1gwxwgvlxlzk3ex59cx7lsswyvplf0rfhunxjhy`
- poly app - `pokt1hufj6cdgu83dluput6klhmh54vtrgtl3drttva`

```bash
export DB_CONNECTION_STRING='postgresql://portal_user:portal_password@localhost:5435/portal_db'
make portal_db_up

make portal_db_hydrate_gateways pokt1lf0kekv9zcv9v3wy4v6jx2wh7v4665s8e0sl9s https://shannon-grove-rpc.mainnet.poktroll.com pocket
make portal_db_hydrate_services 'eth,poly,solana,xrplevm' https://shannon-grove-rpc.mainnet.poktroll.com pocket
make portal_db_hydrate_applications 'pokt1xd8jrccxtlzs8svrmg6gukn7umln7c2ww327xx,pokt185tgfw9lxyuznh9rz89556l4p8dshdkjd5283d,pokt1gwxwgvlxlzk3ex59cx7lsswyvplf0rfhunxjhy,pokt1hufj6cdgu83dluput6klhmh54vtrgtl3drttva' https://shannon-grove-rpc.mainnet.poktroll.com pocket

psql $DB_CONNECTION_STRING
SELECT * FROM gateways;
SELECT * FROM services;
SELECT * FROM applications;
```

## Interacting with the database

:::tip make helpers

You can run the following command to see all available `make` targets:

```bash
make | grep --line-buffered "portal"
```

:::

### `make` Targets

- `make portal_db_up` creates the Portal DB with the base schema (`./schema/001_portal_init.sql`) and runs the Portal DB on port `:5435`.
- `make portal_db_down` stops running the local Portal DB.
- `make portal_db_env` creates and inits the Database, and helps set up the local development environment.
- `make portal_db_clean` stops the local Portal DB and deletes the database and drops the schema.
- `make portal_db_status` Check status of portal-db PostgreSQL container
- `make portal_db_logs` Show logs from portal-db PostgreSQL container
- `make portal_db_connect` Connect to the portal database using psql

### `scripts`

Helper scripts exist to quickly populate the database with real data.

- `./scripts/hydrate-gateways.sh` - Retrieves all onchain data about a given `gateway` and populates the Portal DB
- `./scripts/hydrate-services.sh` - Retrieves all onchain data about a set of `services` and populates the Portal DB
- `./scripts/hydrate-applications.sh` - Retrieves all onchain data about a set of `applications` and populates the Portal DB

## Tools

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

::: warning EXPERIMENTAL

Using a postgres MCP server is experimental but worth a shot!

:::

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
