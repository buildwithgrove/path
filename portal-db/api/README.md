# Portal Database API

<!-- TODO_DOCUMENTATION(@commoddity): Add section describing potential deployment to production using Pulumi, similar to how Portal database itself is deployed in the infra repo. -->

**PostgREST configuration** and **SDK generation tools** for the Portal Database.

PostgREST automatically generates a REST API from the PostgreSQL database schema.

## 🚀 Quick Start <!-- omit in toc -->

```bash
# From portal-db directory
make portal-db-up                    # Start PostgreSQL + PostgREST (port 3000)
make postgrest-hydrate-testdata      # Add test data
curl http://localhost:3000/networks | jq  # Test API
```

# Table of Contents <!-- omit in toc -->

- [Portal Database API](#portal-database-api)
  - [📁 Folder Structure](#-folder-structure)
  - [⚙️ Configuration](#️-configuration)
  - [💾 Database Transactions](#-database-transactions)
  - [🛠️ SDK Generation](#️-sdk-generation)
  - [🔧 Development Commands](#-development-commands)
  - [📚 Resources](#-resources)

## How it Works

**PostgREST** introspects PostgreSQL schema and auto-generates REST endpoints:

```
Database Schema → PostgREST → OpenAPI Spec → Go/TypeScript SDKs
```

## 📁 Folder Structure

```
api/
├── codegen/                # SDK generation scripts
├── openapi/                # Generated OpenAPI spec
├── scripts/                # Helper scripts (JWT, auth testing)
└── postgrest.conf          # PostgREST configuration
```

## ⚙️ Configuration

PostgREST configuration in `postgrest.conf`:

```ini
db-uri = "postgresql://authenticator:password@postgres:5432/portal_db"
db-schemas = "public,api"
server-port = 3000
```

Database roles defined in `../schema/002_postgrest_init.sql`:
- `anon` - Public data (networks, services)
- `authenticated` - User data (accounts, applications)

## Authentication (Optional)

JWT authentication available for protected endpoints:

```bash
make postgrest-gen-jwt      # Generate test token
make test-postgrest-auth    # Test authentication flow
```

See `scripts/gen-jwt.sh` for details.

## 💾 Database Transactions

PostgreSQL functions are auto-exposed as RPC endpoints:

```sql
CREATE FUNCTION public.create_portal_application(...) RETURNS JSON AS $$
  -- Multi-step transaction logic
$$ LANGUAGE plpgsql;
```

Usage: `curl -X POST http://localhost:3000/rpc/create_portal_application -d '{...}'`

Test: `make test-postgrest-portal-app-creation`

## 🛠️ SDK Generation

```bash
make postgrest-generate-all       # Generate OpenAPI spec + Go/TS SDKs
make postgrest-generate-openapi   # OpenAPI spec only
```

Generated SDKs: `../sdk/go/` and `../sdk/typescript/`

## 🔧 Development Commands

```bash
make portal-db-up                    # Start services
make portal-db-down                  # Stop services
make portal-db-logs                  # View logs
make postgrest-hydrate-testdata      # Add test data
make postgrest-generate-all          # Regenerate SDKs after schema changes
make test-postgrest-auth             # Test JWT authentication
make postgrest-gen-jwt               # Generate JWT token
```

**After schema changes:**
1. Edit `../schema/001_portal_init.sql`
2. `make portal-db-down && make portal-db-up`
3. `make postgrest-generate-all`

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
