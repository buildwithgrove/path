# Portal Database API <!-- omit in toc -->

**PostgREST configuration** for the Portal Database.

PostgREST automatically generates a REST API from the PostgreSQL database schema.

## 🚀 Quick Start <!-- omit in toc -->

```bash
cd portal-db
make portal-db-up                    # Start PostgreSQL + PostgREST (port 3000)
make postgrest-hydrate-testdata      # Add test data
make postgrest-gen-jwt admin         # Admin token (copy export command)
make postgrest-gen-jwt reader        # Reader token (optional)

# Paste the export commands here
export JWT_TOKEN_ADMIN="..."
export JWT_TOKEN_READER="..."
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

You can run `make` to see a list of available commands.

You can also run `make quickstart` for a guided walkthrough.

## Table of Contents <!-- omit in toc -->

- [How it Works](#how-it-works)
- [⚙️ Configuration](#️-configuration)
- [Authentication](#authentication)
- [🔧 Development Commands](#-development-commands)
- [Query Examples](#query-examples)
- [📚 Resources](#-resources)

## How it Works

**PostgREST** introspects PostgreSQL schema and auto-generates REST endpoints:

```bash
Database Schema → PostgREST → OpenAPI Spec → Go/TypeScript SDKs
```

## ⚙️ Configuration

PostgREST configuration in `postgrest.conf`:

```ini
db-uri = "postgresql://authenticator:password@postgres:5432/portal_db"
db-schemas = "public,api"
server-port = 3000
```

Database roles defined in `../schema/002_postgrest_init.sql`:

- `authenticator` - Connection role used exclusively by PostgREST (no direct API access)
- `portal_db_admin` - JWT-backed role with read/write access (subject to RLS)
- `portal_db_reader` - JWT-backed role with read-only access (subject to RLS)
- `anon` - Default unauthenticated role with no privileges

## Authentication

All API access flows through manually issued JWTs.

The token must contain a `role` claim set to either `portal_db_admin` or `portal_db_reader`. The helper
script wires this up for local testing:

```bash
# From portal-db directory
./api/scripts/postgrest-gen-jwt.sh portal_db_reader user@example.com
export JWT_TOKEN=$(./api/scripts/postgrest-gen-jwt.sh --token-only portal_db_admin admin@example.com)
curl http://localhost:3000/organizations -H "Authorization: Bearer $JWT_TOKEN"
```

PostgREST validates the signature, sets `SET ROLE <role claim>`, and row-level
security policies in `../schema/002_postgrest_init.sql` scope the session based
on the selected role.

## 🔧 Development Commands

Run `make help` for a list of available commands.

## Query Examples

```bash
# Filtering
curl "http://localhost:3000/services?active=eq.true"

# Field selection
curl "http://localhost:3000/services?select=service_id,service_name"

# Sorting & pagination
curl "http://localhost:3000/services?order=service_name.asc&limit=10"

# Joins
curl "http://localhost:3000/services?select=*,service_endpoints(*)"
```

## 📚 Resources

- [PostgREST Documentation](https://postgrest.org/en/stable/)
- [OpenAPI Specification](https://swagger.io/specification/)
