# Portal Database API <!-- omit in toc -->

**PostgREST configuration** for the Portal Database.

PostgREST automatically generates a REST API from the PostgreSQL database schema.

## Table of Contents <!-- omit in toc -->

- [QuickStart](#quickstart)
- [Walkthrough](#walkthrough)
- [Authentication](#authentication)
  - [Auth Summary](#auth-summary)
  - [Database Roles Roles](#database-roles-roles)
  - [Testing auth locally](#testing-auth-locally)
  - [JWT Generation](#jwt-generation)
- [How it Works](#how-it-works)
- [📚 Resources](#-resources)

## QuickStart

Run `make quickstart` for a guided walkthrough.

Run `make` to see a list of available commands.

## Walkthrough

The following is a minimal walkthrough to get started running the Portal DB API
and sending a few requests locally:

```bash
# Start PostgreSQL + PostgREST (port 3000)
cd portal-db
make portal-db-up

# Add test data
make postgrest-hydrate-testdata

# Admin token (copy export command)
make postgrest-gen-jwt admin

# Reader token (optional)
make postgrest-gen-jwt reader

# Paste the export commands here
export JWT_TOKEN_ADMIN="..."
export JWT_TOKEN_READER="..."

# Test the API
curl http://localhost:3000/networks -H "Authorization: Bearer $JWT_TOKEN_READER" | jq
curl http://localhost:3000/organizations -H "Authorization: Bearer $JWT_TOKEN_READER" | jq
curl http://localhost:3000/portal_accounts -H "Authorization: Bearer $JWT_TOKEN_READER" | jq
curl -X POST http://localhost:3000/portal_applications \
  -H "Authorization: Bearer $JWT_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -H "Prefer: return=representation" \
  -d '{"portal_account_id":"10000000-0000-0000-0000-000000000004","portal_application_name":"CLI Quickstart App","secret_key_hash":"demo","secret_key_required":false}' \
  | jq

# Refresh OpenAPI spec before launching Swagger UI
make postgrest-generate-openapi

# Launch Swagger UI
make postgrest-swagger-ui
```

## Authentication

The PostgREST API is authenticated via the SQL migration in [002_postgrest_init.sql](../schema/002_postgrest_init.sql).

### Auth Summary

1. SSL Certs to connect to the DB
2. JWT to authenticate into the DB as a `portal_db_*` user
3. Top level roles authenticated into the DB subject to RLS (e.g. `portal_db_admin` or `portal_db_reader`)
4. Portal Application roles defined within the tables (see `rbac` of each table)

### Database Roles Roles

- `authenticator` - "Chameleon" role used exclusively by PostgREST for JWT authentication (no direct API access)
- `portal_db_admin` - JWT-backed role with read/write access (subject to RLS)
- `portal_db_reader` - JWT-backed role with read-only access (subject to RLS)
- `anon` - Default unauthenticated role with no privileges

### Testing auth locally

Run `make` from the `portal-db` directory shows the following scripts which can be used to test things locally:

```bash
=== 🔐 Authentication & Testing ===
test-postgrest-auth                Test JWT authentication flow
test-postgrest-portal-app-creation Test portal application creation and retrieval via authenticated Postgres flow
postgrest-gen-jwt                  Generate JWT token
```

### JWT Generation

```bash
# Admin JWT
make postgrest-gen-jwt admin

# Reader JWT
make postgrest-gen-jwt reader
```

## How it Works

**PostgREST** introspects PostgreSQL schema and auto-generates REST endpoints:

```bash
Database Schema → PostgREST → OpenAPI Spec → (Coming Soon) SDKs (Go, TypeScript, etc...)
```

## 📚 Resources

- [PostgREST Documentation](https://postgrest.org/en/stable/)
- [OpenAPI Specification](https://swagger.io/specification/)
